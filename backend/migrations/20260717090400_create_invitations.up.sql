CREATE TYPE invitation_status AS ENUM (
    'PENDING',
    'ACCEPTED',
    'REVOKED',
    'EXPIRED'
);

CREATE TABLE organization_invitations (
    id UUID DEFAULT uuidv7 () NOT NULL,
    organization_id UUID NOT NULL,
    email VARCHAR(320) NOT NULL,
    email_normalized VARCHAR(320) NOT NULL,
    status invitation_status DEFAULT 'PENDING' NOT NULL,
    secret_hash BYTEA NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    accepted_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ,
    invited_by_membership_id UUID NOT NULL,
    accepted_by_membership_id UUID,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    CONSTRAINT organization_invitations_pkey PRIMARY KEY (id),
    CONSTRAINT organization_invitations_organization_id_id_unique UNIQUE (organization_id, id),
    CONSTRAINT organization_invitations_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES organizations (id) ON DELETE RESTRICT,
    CONSTRAINT organization_invitations_inviter_fkey FOREIGN KEY (
        organization_id,
        invited_by_membership_id
    ) REFERENCES organization_memberships (organization_id, id) ON DELETE RESTRICT,
    CONSTRAINT organization_invitations_accepted_membership_fkey FOREIGN KEY (
        organization_id,
        accepted_by_membership_id
    ) REFERENCES organization_memberships (organization_id, id) ON DELETE RESTRICT,
    CONSTRAINT organization_invitations_email_not_blank CHECK (BTRIM (email) <> ''),
    CONSTRAINT organization_invitations_email_normalized_not_blank CHECK (
        BTRIM (email_normalized) <> ''
    ),
    CONSTRAINT organization_invitations_email_normalized_canonical CHECK (
        email_normalized = LOWER(BTRIM (email_normalized))
    ),
    CONSTRAINT organization_invitations_secret_hash_length CHECK (
        OCTET_LENGTH(secret_hash) = 32
    ),
    CONSTRAINT organization_invitations_expires_at_check CHECK (expires_at > created_at),
    CONSTRAINT organization_invitations_status_timestamps_check CHECK (
        (
            status = 'PENDING'
            AND accepted_at IS NULL
            AND revoked_at IS NULL
        )
        OR (
            status = 'ACCEPTED'
            AND accepted_at IS NOT NULL
            AND revoked_at IS NULL
        )
        OR (
            status = 'REVOKED'
            AND accepted_at IS NULL
            AND revoked_at IS NOT NULL
        )
        OR (
            status = 'EXPIRED'
            AND accepted_at IS NULL
            AND revoked_at IS NULL
        )
    ),
    CONSTRAINT organization_invitations_accepted_at_check CHECK (
        accepted_at IS NULL
        OR (
            accepted_at >= created_at
            AND accepted_at <= expires_at
        )
    ),
    CONSTRAINT organization_invitations_revoked_at_check CHECK (
        revoked_at IS NULL
        OR revoked_at >= created_at
    ),
    CONSTRAINT organization_invitations_updated_at_check CHECK (updated_at >= created_at)
);

CREATE UNIQUE INDEX idx_organization_invitations_pending_email ON organization_invitations (
    organization_id,
    email_normalized
)
WHERE
    status = 'PENDING';

CREATE TRIGGER trg_organization_invitations_touch_updated_at
BEFORE UPDATE ON organization_invitations
FOR EACH ROW
EXECUTE FUNCTION touch_updated_at();

CREATE FUNCTION validate_invitation_status_transition()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF OLD.status <> 'PENDING' AND NEW.status <> OLD.status THEN
        RAISE EXCEPTION 'a terminal invitation cannot return to another status'
            USING ERRCODE = '23514',
                  CONSTRAINT = 'organization_invitations_status_terminal';
    END IF;

    RETURN NEW;
END;
$$;

CREATE TRIGGER trg_organization_invitations_validate_status_transition
BEFORE UPDATE OF status ON organization_invitations
FOR EACH ROW
EXECUTE FUNCTION validate_invitation_status_transition();

COMMENT ON
TABLE organization_invitations IS 'Stores opaque one-time invitations to join an organization.';

COMMENT ON COLUMN organization_invitations.id IS 'UUIDv7 selector included in the external invitation token.';

COMMENT ON COLUMN organization_invitations.organization_id IS 'Organization tenant issuing the invitation.';

COMMENT ON COLUMN organization_invitations.email IS 'Trimmed invitee email preserving display casing.';

COMMENT ON COLUMN organization_invitations.email_normalized IS 'Trimmed lowercase email used for pending-invitation uniqueness.';

COMMENT ON COLUMN organization_invitations.secret_hash IS 'Thirty-two-byte hash of the opaque invitation secret.';

COMMENT ON COLUMN organization_invitations.invited_by_membership_id IS 'Membership that issued the invitation.';

COMMENT ON COLUMN organization_invitations.accepted_by_membership_id IS 'Membership created or selected when the invitation was accepted.';

CREATE TABLE invitation_role_bindings (
    id UUID DEFAULT uuidv7 () NOT NULL,
    organization_id UUID NOT NULL,
    invitation_id UUID NOT NULL,
    role_id UUID NOT NULL,
    store_id UUID,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    CONSTRAINT invitation_role_bindings_pkey PRIMARY KEY (id),
    CONSTRAINT invitation_role_bindings_invitation_fkey FOREIGN KEY (
        organization_id,
        invitation_id
    ) REFERENCES organization_invitations (organization_id, id) ON DELETE RESTRICT,
    CONSTRAINT invitation_role_bindings_role_fkey FOREIGN KEY (organization_id, role_id) REFERENCES roles (organization_id, id) ON DELETE RESTRICT,
    CONSTRAINT invitation_role_bindings_store_fkey FOREIGN KEY (organization_id, store_id) REFERENCES stores (organization_id, id) ON DELETE RESTRICT
);

CREATE UNIQUE INDEX idx_invitation_role_bindings_org_unique ON invitation_role_bindings (
    organization_id,
    invitation_id,
    role_id
)
WHERE
    store_id IS NULL;

CREATE UNIQUE INDEX idx_invitation_role_bindings_store_unique ON invitation_role_bindings (
    organization_id,
    invitation_id,
    role_id,
    store_id
)
WHERE
    store_id IS NOT NULL;

CREATE TRIGGER trg_invitation_role_bindings_validate_scope
BEFORE INSERT OR UPDATE ON invitation_role_bindings
FOR EACH ROW
EXECUTE FUNCTION validate_role_binding_scope();

CREATE FUNCTION validate_owner_invitation_binding()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    target_role_key VARCHAR(80);
    target_role_is_system BOOLEAN;
    inviter_membership_id UUID;
BEGIN
    SELECT key, is_system
    INTO target_role_key, target_role_is_system
    FROM roles
    WHERE organization_id = NEW.organization_id
      AND id = NEW.role_id
    FOR SHARE;

    IF NOT FOUND OR NOT target_role_is_system OR target_role_key <> 'owner' THEN
        RETURN NEW;
    END IF;

    SELECT invited_by_membership_id
    INTO inviter_membership_id
    FROM organization_invitations
    WHERE organization_id = NEW.organization_id
      AND id = NEW.invitation_id
    FOR UPDATE;

    IF NOT FOUND THEN
        RETURN NEW;
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM organization_memberships AS membership
        JOIN membership_role_bindings AS binding
          ON binding.organization_id = membership.organization_id
         AND binding.membership_id = membership.id
        WHERE membership.organization_id = NEW.organization_id
          AND membership.id = inviter_membership_id
          AND membership.status = 'ACTIVE'
          AND binding.role_id = NEW.role_id
          AND binding.store_id IS NULL
          AND (binding.expires_at IS NULL OR binding.expires_at > NOW())
    ) THEN
        RAISE EXCEPTION 'only an active owner can invite another owner'
            USING ERRCODE = '23514',
                  CONSTRAINT = 'invitation_role_bindings_owner_inviter_check';
    END IF;

    RETURN NEW;
END;
$$;

CREATE TRIGGER trg_invitation_role_bindings_validate_owner
BEFORE INSERT OR UPDATE ON invitation_role_bindings
FOR EACH ROW
EXECUTE FUNCTION validate_owner_invitation_binding();

CREATE FUNCTION validate_owner_invitation_update()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    owner_role_id UUID;
BEGIN
    IF NEW.organization_id IS NOT DISTINCT FROM OLD.organization_id
       AND NEW.invited_by_membership_id IS NOT DISTINCT FROM OLD.invited_by_membership_id THEN
        RETURN NEW;
    END IF;

    SELECT role.id
    INTO owner_role_id
    FROM invitation_role_bindings AS binding
    JOIN roles AS role
      ON role.organization_id = binding.organization_id
     AND role.id = binding.role_id
    WHERE binding.organization_id = OLD.organization_id
      AND binding.invitation_id = OLD.id
      AND role.is_system
      AND role.key = 'owner'
    LIMIT 1;

    IF NOT FOUND THEN
        RETURN NEW;
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM organization_memberships AS membership
        JOIN membership_role_bindings AS binding
          ON binding.organization_id = membership.organization_id
         AND binding.membership_id = membership.id
        WHERE membership.organization_id = NEW.organization_id
          AND membership.id = NEW.invited_by_membership_id
          AND membership.status = 'ACTIVE'
          AND binding.role_id = owner_role_id
          AND binding.store_id IS NULL
          AND (binding.expires_at IS NULL OR binding.expires_at > NOW())
    ) THEN
        RAISE EXCEPTION 'only an active owner can remain the inviter of an owner invitation'
            USING ERRCODE = '23514',
                  CONSTRAINT = 'organization_invitations_owner_inviter_check';
    END IF;

    RETURN NEW;
END;
$$;

CREATE TRIGGER trg_organization_invitations_validate_owner_update
BEFORE UPDATE OF organization_id, invited_by_membership_id
ON organization_invitations
FOR EACH ROW
EXECUTE FUNCTION validate_owner_invitation_update();

CREATE FUNCTION validate_role_update_for_invitations()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF NEW.assignment_scope IS DISTINCT FROM OLD.assignment_scope
       AND EXISTS (
           SELECT 1
           FROM invitation_role_bindings
           WHERE organization_id = OLD.organization_id
             AND role_id = OLD.id
             AND (
                 (NEW.assignment_scope = 'STORE' AND store_id IS NULL)
                 OR (NEW.assignment_scope = 'ORGANIZATION' AND store_id IS NOT NULL)
             )
       ) THEN
        RAISE EXCEPTION 'the role scope change conflicts with an existing invitation binding'
            USING ERRCODE = '23514',
                  CONSTRAINT = 'roles_assignment_invitation_binding_check';
    END IF;

    RETURN NEW;
END;
$$;

CREATE TRIGGER trg_roles_validate_invitation_bindings
BEFORE UPDATE OF assignment_scope ON roles
FOR EACH ROW
EXECUTE FUNCTION validate_role_update_for_invitations();

CREATE FUNCTION ensure_invitation_has_role_binding()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    checked_organization_id UUID;
    checked_invitation_id UUID;
BEGIN
    IF TG_TABLE_NAME = 'organization_invitations' THEN
        checked_organization_id := NEW.organization_id;
        checked_invitation_id := NEW.id;
    ELSE
        checked_organization_id := OLD.organization_id;
        checked_invitation_id := OLD.invitation_id;
    END IF;

    PERFORM 1
    FROM organization_invitations
    WHERE organization_id = checked_organization_id
      AND id = checked_invitation_id
    FOR UPDATE;

    IF FOUND AND NOT EXISTS (
        SELECT 1
        FROM invitation_role_bindings
        WHERE organization_id = checked_organization_id
          AND invitation_id = checked_invitation_id
    ) THEN
        RAISE EXCEPTION 'an invitation must contain at least one role binding'
            USING ERRCODE = '23514',
                  CONSTRAINT = 'organization_invitations_role_required';
    END IF;

    RETURN NULL;
END;
$$;

CREATE CONSTRAINT TRIGGER trg_organization_invitations_require_role
AFTER INSERT OR UPDATE ON organization_invitations
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW
EXECUTE FUNCTION ensure_invitation_has_role_binding();

CREATE CONSTRAINT TRIGGER trg_invitation_role_bindings_require_role
AFTER DELETE OR UPDATE ON invitation_role_bindings
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW
EXECUTE FUNCTION ensure_invitation_has_role_binding();

COMMENT ON
TABLE invitation_role_bindings IS 'Stores the organization or store roles granted when an invitation is accepted.';

COMMENT ON COLUMN invitation_role_bindings.organization_id IS 'Tenant key repeated across every invitation binding relationship.';

COMMENT ON COLUMN invitation_role_bindings.invitation_id IS 'Invitation that will produce the role binding.';

COMMENT ON COLUMN invitation_role_bindings.role_id IS 'Role granted when the invitation is accepted.';

COMMENT ON COLUMN invitation_role_bindings.store_id IS 'Required for store roles and forbidden for organization roles.';
