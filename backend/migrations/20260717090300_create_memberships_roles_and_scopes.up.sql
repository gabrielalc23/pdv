CREATE TYPE membership_status AS ENUM (
    'ACTIVE',
    'SUSPENDED',
    'REMOVED'
);

CREATE TYPE permission_scope_level AS ENUM (
    'ORGANIZATION',
    'STORE'
);

CREATE TYPE role_assignment_scope AS ENUM (
    'ORGANIZATION',
    'STORE'
);

CREATE TABLE organization_memberships (
    id UUID DEFAULT uuidv7 () NOT NULL,
    organization_id UUID NOT NULL,
    user_id UUID NOT NULL,
    status membership_status DEFAULT 'ACTIVE' NOT NULL,
    default_store_id UUID,
    authorization_version BIGINT DEFAULT 1 NOT NULL,
    joined_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    suspended_at TIMESTAMPTZ,
    removed_at TIMESTAMPTZ,
    created_by_user_id UUID NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    CONSTRAINT organization_memberships_pkey PRIMARY KEY (id),
    CONSTRAINT organization_memberships_organization_id_id_unique UNIQUE (organization_id, id),
    CONSTRAINT organization_memberships_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES organizations (id) ON DELETE RESTRICT,
    CONSTRAINT organization_memberships_user_id_fkey FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE RESTRICT,
    CONSTRAINT organization_memberships_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES users (id) ON DELETE RESTRICT,
    CONSTRAINT organization_memberships_default_store_fkey FOREIGN KEY (
        organization_id,
        default_store_id
    ) REFERENCES stores (organization_id, id) ON DELETE RESTRICT,
    CONSTRAINT organization_memberships_authorization_version_positive CHECK (authorization_version > 0),
    CONSTRAINT organization_memberships_status_timestamps_check CHECK (
        (
            status = 'ACTIVE'
            AND suspended_at IS NULL
            AND removed_at IS NULL
        )
        OR (
            status = 'SUSPENDED'
            AND suspended_at IS NOT NULL
            AND removed_at IS NULL
        )
        OR (
            status = 'REMOVED'
            AND removed_at IS NOT NULL
        )
    ),
    CONSTRAINT organization_memberships_joined_at_check CHECK (joined_at >= created_at),
    CONSTRAINT organization_memberships_suspended_at_check CHECK (
        suspended_at IS NULL
        OR suspended_at >= joined_at
    ),
    CONSTRAINT organization_memberships_removed_at_check CHECK (
        removed_at IS NULL
        OR removed_at >= joined_at
    ),
    CONSTRAINT organization_memberships_lifecycle_order_check CHECK (
        suspended_at IS NULL
        OR removed_at IS NULL
        OR suspended_at <= removed_at
    ),
    CONSTRAINT organization_memberships_updated_at_check CHECK (updated_at >= created_at)
);

CREATE UNIQUE INDEX idx_organization_memberships_active_user ON organization_memberships (organization_id, user_id)
WHERE
    status <> 'REMOVED';

CREATE INDEX idx_organization_memberships_user_status ON organization_memberships (user_id, status);

CREATE INDEX idx_organization_memberships_organization_status ON organization_memberships (organization_id, status);

CREATE INDEX idx_organization_memberships_default_store ON organization_memberships (
    organization_id,
    default_store_id
);

CREATE FUNCTION validate_membership_status_transition()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF OLD.status = 'REMOVED' AND NEW.status <> 'REMOVED' THEN
        RAISE EXCEPTION 'a removed membership cannot be reactivated'
            USING ERRCODE = '23514',
                  CONSTRAINT = 'organization_memberships_removed_terminal';
    END IF;

    RETURN NEW;
END;
$$;

CREATE TRIGGER trg_organization_memberships_validate_status_transition
BEFORE UPDATE OF status ON organization_memberships
FOR EACH ROW
EXECUTE FUNCTION validate_membership_status_transition();

CREATE TRIGGER trg_organization_memberships_touch_updated_at
BEFORE UPDATE ON organization_memberships
FOR EACH ROW
EXECUTE FUNCTION touch_updated_at();

COMMENT ON
TABLE organization_memberships IS 'Associates a global user with an organization while preserving membership history.';

COMMENT ON COLUMN organization_memberships.id IS 'Internal UUIDv7 membership identifier.';

COMMENT ON COLUMN organization_memberships.organization_id IS 'Organization tenant to which the user belongs.';

COMMENT ON COLUMN organization_memberships.user_id IS 'Global user represented by this membership.';

COMMENT ON COLUMN organization_memberships.status IS 'Lifecycle status within the organization; removed memberships are terminal.';

COMMENT ON COLUMN organization_memberships.default_store_id IS 'Optional default store within the same organization.';

COMMENT ON COLUMN organization_memberships.authorization_version IS 'Monotonic version used to invalidate membership authorization.';

COMMENT ON COLUMN organization_memberships.created_by_user_id IS 'Global user that initiated creation of the membership.';

CREATE TABLE permission_scopes (
    code VARCHAR(100) NOT NULL,
    resource VARCHAR(60) NOT NULL,
    action VARCHAR(60) NOT NULL,
    scope_level permission_scope_level NOT NULL,
    description VARCHAR(255) NOT NULL,
    is_assignable BOOLEAN DEFAULT TRUE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    CONSTRAINT permission_scopes_pkey PRIMARY KEY (code),
    CONSTRAINT permission_scopes_resource_not_blank CHECK (BTRIM (resource) <> ''),
    CONSTRAINT permission_scopes_action_not_blank CHECK (BTRIM (action) <> ''),
    CONSTRAINT permission_scopes_description_not_blank CHECK (BTRIM (description) <> ''),
    CONSTRAINT permission_scopes_resource_format CHECK (
        resource ~ '^[a-z][a-z0-9_]*$'
    ),
    CONSTRAINT permission_scopes_action_format CHECK (
        action ~ '^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$'
    ),
    CONSTRAINT permission_scopes_code_components_check CHECK (
        code = resource || '.' || action
    )
);

COMMENT ON
TABLE permission_scopes IS 'Platform-managed catalog of explicit authorization scopes.';

COMMENT ON COLUMN permission_scopes.code IS 'Lowercase scope code composed from resource and action.';

COMMENT ON COLUMN permission_scopes.scope_level IS 'Context level at which the scope can become effective.';

COMMENT ON COLUMN permission_scopes.is_assignable IS 'Whether the scope may be granted to custom roles.';

CREATE TABLE roles (
    id UUID DEFAULT uuidv7 () NOT NULL,
    organization_id UUID NOT NULL,
    key VARCHAR(80) NOT NULL,
    name VARCHAR(120) NOT NULL,
    description VARCHAR(255),
    assignment_scope role_assignment_scope NOT NULL,
    is_system BOOLEAN DEFAULT FALSE NOT NULL,
    is_mutable BOOLEAN DEFAULT TRUE NOT NULL,
    is_active BOOLEAN DEFAULT TRUE NOT NULL,
    created_by_membership_id UUID,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    CONSTRAINT roles_pkey PRIMARY KEY (id),
    CONSTRAINT roles_organization_id_id_unique UNIQUE (organization_id, id),
    CONSTRAINT roles_organization_id_key_unique UNIQUE (organization_id, key),
    CONSTRAINT roles_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES organizations (id) ON DELETE RESTRICT,
    CONSTRAINT roles_created_by_membership_fkey FOREIGN KEY (
        organization_id,
        created_by_membership_id
    ) REFERENCES organization_memberships (organization_id, id) ON DELETE RESTRICT,
    CONSTRAINT roles_key_not_blank CHECK (BTRIM (key) <> ''),
    CONSTRAINT roles_name_not_blank CHECK (BTRIM (name) <> ''),
    CONSTRAINT roles_key_format CHECK (
        key ~ '^[a-z0-9]+(?:_[a-z0-9]+)*$'
    ),
    CONSTRAINT roles_system_immutable_check CHECK (
        NOT is_system
        OR NOT is_mutable
    ),
    CONSTRAINT roles_owner_shape_check CHECK (
        key <> 'owner'
        OR (
            assignment_scope = 'ORGANIZATION'
            AND is_system
            AND NOT is_mutable
            AND is_active
        )
    ),
    CONSTRAINT roles_updated_at_check CHECK (updated_at >= created_at)
);

CREATE TRIGGER trg_roles_touch_updated_at
BEFORE UPDATE ON roles
FOR EACH ROW
EXECUTE FUNCTION touch_updated_at();

COMMENT ON
TABLE roles IS 'Organization-owned RBAC roles composed from platform permission scopes.';

COMMENT ON COLUMN roles.organization_id IS 'Organization tenant that owns the role.';

COMMENT ON COLUMN roles.key IS 'Stable lowercase snake_case role key unique within the organization.';

COMMENT ON COLUMN roles.assignment_scope IS 'Determines whether bindings apply organization-wide or to one store.';

COMMENT ON COLUMN roles.is_system IS 'Identifies a role created and managed by the platform bootstrap.';

COMMENT ON COLUMN roles.is_mutable IS 'Controls whether administrative flows may modify the role.';

COMMENT ON COLUMN roles.is_active IS 'Controls whether bindings for the role grant authorization.';

COMMENT ON COLUMN roles.created_by_membership_id IS 'Optional creator membership from the same organization.';

CREATE TABLE role_scopes (
    organization_id UUID NOT NULL,
    role_id UUID NOT NULL,
    scope_code VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    CONSTRAINT role_scopes_pkey PRIMARY KEY (
        organization_id,
        role_id,
        scope_code
    ),
    CONSTRAINT role_scopes_role_fkey FOREIGN KEY (organization_id, role_id) REFERENCES roles (organization_id, id) ON DELETE RESTRICT,
    CONSTRAINT role_scopes_scope_code_fkey FOREIGN KEY (scope_code) REFERENCES permission_scopes (code) ON DELETE RESTRICT
);

CREATE FUNCTION validate_role_scope_assignment()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    role_scope role_assignment_scope;
    role_is_system BOOLEAN;
    permission_level permission_scope_level;
    permission_is_assignable BOOLEAN;
BEGIN
    SELECT assignment_scope, is_system
    INTO role_scope, role_is_system
    FROM roles
    WHERE organization_id = NEW.organization_id
      AND id = NEW.role_id
    FOR SHARE;

    IF NOT FOUND THEN
        RETURN NEW;
    END IF;

    SELECT scope_level, is_assignable
    INTO permission_level, permission_is_assignable
    FROM permission_scopes
    WHERE code = NEW.scope_code;

    IF NOT FOUND THEN
        RETURN NEW;
    END IF;

    IF role_scope = 'STORE' AND permission_level <> 'STORE' THEN
        RAISE EXCEPTION 'a store role can contain only store-level scopes'
            USING ERRCODE = '23514',
                  CONSTRAINT = 'role_scopes_assignment_scope_check';
    END IF;

    IF NOT permission_is_assignable AND NOT role_is_system THEN
        RAISE EXCEPTION 'a non-assignable scope cannot be added to a custom role'
            USING ERRCODE = '23514',
                  CONSTRAINT = 'role_scopes_assignable_check';
    END IF;

    RETURN NEW;
END;
$$;

CREATE TRIGGER trg_role_scopes_validate_assignment
BEFORE INSERT OR UPDATE ON role_scopes
FOR EACH ROW
EXECUTE FUNCTION validate_role_scope_assignment();

COMMENT ON
TABLE role_scopes IS 'Associates organization roles with platform permission scopes.';

COMMENT ON COLUMN role_scopes.organization_id IS 'Tenant key repeated to enforce same-organization relationships.';

COMMENT ON COLUMN role_scopes.scope_code IS 'Platform scope granted through the role.';

CREATE TABLE membership_role_bindings (
    id UUID DEFAULT uuidv7 () NOT NULL,
    organization_id UUID NOT NULL,
    membership_id UUID NOT NULL,
    role_id UUID NOT NULL,
    store_id UUID,
    created_by_membership_id UUID NOT NULL,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    CONSTRAINT membership_role_bindings_pkey PRIMARY KEY (id),
    CONSTRAINT membership_role_bindings_membership_fkey FOREIGN KEY (
        organization_id,
        membership_id
    ) REFERENCES organization_memberships (organization_id, id) ON DELETE RESTRICT,
    CONSTRAINT membership_role_bindings_role_fkey FOREIGN KEY (organization_id, role_id) REFERENCES roles (organization_id, id) ON DELETE RESTRICT,
    CONSTRAINT membership_role_bindings_creator_fkey FOREIGN KEY (
        organization_id,
        created_by_membership_id
    ) REFERENCES organization_memberships (organization_id, id) ON DELETE RESTRICT,
    CONSTRAINT membership_role_bindings_store_fkey FOREIGN KEY (organization_id, store_id) REFERENCES stores (organization_id, id) ON DELETE RESTRICT,
    CONSTRAINT membership_role_bindings_expires_at_check CHECK (
        expires_at IS NULL
        OR expires_at > created_at
    )
);

CREATE UNIQUE INDEX idx_membership_role_bindings_org_unique ON membership_role_bindings (
    organization_id,
    membership_id,
    role_id
)
WHERE
    store_id IS NULL;

CREATE UNIQUE INDEX idx_membership_role_bindings_store_unique ON membership_role_bindings (
    organization_id,
    membership_id,
    role_id,
    store_id
)
WHERE
    store_id IS NOT NULL;

CREATE FUNCTION validate_role_binding_scope()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    binding_scope role_assignment_scope;
    bound_role_key VARCHAR(80);
    bound_role_is_system BOOLEAN;
BEGIN
    SELECT assignment_scope, key, is_system
    INTO binding_scope, bound_role_key, bound_role_is_system
    FROM roles
    WHERE organization_id = NEW.organization_id
      AND id = NEW.role_id
    FOR SHARE;

    IF NOT FOUND THEN
        RETURN NEW;
    END IF;

    IF binding_scope = 'ORGANIZATION' AND NEW.store_id IS NOT NULL THEN
        RAISE EXCEPTION 'an organization role binding cannot include a store'
            USING ERRCODE = '23514',
                  CONSTRAINT = 'role_binding_store_forbidden';
    END IF;

    IF binding_scope = 'STORE' AND NEW.store_id IS NULL THEN
        RAISE EXCEPTION 'a store role binding requires a store'
            USING ERRCODE = '23514',
                  CONSTRAINT = 'role_binding_store_required';
    END IF;

    IF TG_TABLE_NAME = 'membership_role_bindings'
       AND bound_role_is_system
       AND bound_role_key = 'owner'
       AND to_jsonb(NEW) ->> 'expires_at' IS NOT NULL THEN
        RAISE EXCEPTION 'an owner role binding cannot expire'
            USING ERRCODE = '23514',
                  CONSTRAINT = 'owner_binding_expiration_forbidden';
    END IF;

    RETURN NEW;
END;
$$;

CREATE TRIGGER trg_membership_role_bindings_validate_scope
BEFORE INSERT OR UPDATE ON membership_role_bindings
FOR EACH ROW
EXECUTE FUNCTION validate_role_binding_scope();

CREATE FUNCTION validate_role_update()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF NOT OLD.is_mutable AND ROW(
        NEW.organization_id,
        NEW.key,
        NEW.name,
        NEW.description,
        NEW.assignment_scope,
        NEW.is_system,
        NEW.is_mutable,
        NEW.is_active,
        NEW.created_by_membership_id
    ) IS DISTINCT FROM ROW(
        OLD.organization_id,
        OLD.key,
        OLD.name,
        OLD.description,
        OLD.assignment_scope,
        OLD.is_system,
        OLD.is_mutable,
        OLD.is_active,
        OLD.created_by_membership_id
    ) THEN
        RAISE EXCEPTION 'an immutable role cannot be modified'
            USING ERRCODE = '23514',
                  CONSTRAINT = 'roles_immutable_update';
    END IF;

    IF NEW.is_system IS DISTINCT FROM OLD.is_system THEN
        RAISE EXCEPTION 'the system designation of a role cannot be changed'
            USING ERRCODE = '23514',
                  CONSTRAINT = 'roles_system_designation_immutable';
    END IF;

    IF NEW.assignment_scope IS DISTINCT FROM OLD.assignment_scope THEN
        IF NEW.assignment_scope = 'STORE' AND EXISTS (
            SELECT 1
            FROM role_scopes
            JOIN permission_scopes
              ON permission_scopes.code = role_scopes.scope_code
            WHERE role_scopes.organization_id = OLD.organization_id
              AND role_scopes.role_id = OLD.id
              AND permission_scopes.scope_level = 'ORGANIZATION'
        ) THEN
            RAISE EXCEPTION 'the role scope change conflicts with an organization-level permission scope'
                USING ERRCODE = '23514',
                      CONSTRAINT = 'roles_assignment_scope_check';
        END IF;

        IF EXISTS (
            SELECT 1
            FROM membership_role_bindings
            WHERE organization_id = OLD.organization_id
              AND role_id = OLD.id
              AND (
                  (NEW.assignment_scope = 'STORE' AND store_id IS NULL)
                  OR (NEW.assignment_scope = 'ORGANIZATION' AND store_id IS NOT NULL)
              )
        ) THEN
            RAISE EXCEPTION 'the role scope change conflicts with an existing membership binding'
                USING ERRCODE = '23514',
                      CONSTRAINT = 'roles_assignment_binding_check';
        END IF;
    END IF;

    RETURN NEW;
END;
$$;

CREATE TRIGGER trg_roles_validate_update
BEFORE UPDATE ON roles
FOR EACH ROW
EXECUTE FUNCTION validate_role_update();

COMMENT ON
TABLE membership_role_bindings IS 'Grants organization or store roles to memberships within one tenant.';

COMMENT ON COLUMN membership_role_bindings.organization_id IS 'Tenant key repeated across every binding relationship.';

COMMENT ON COLUMN membership_role_bindings.membership_id IS 'Membership receiving the role.';

COMMENT ON COLUMN membership_role_bindings.role_id IS 'Role granted by this binding.';

COMMENT ON COLUMN membership_role_bindings.store_id IS 'Required for store roles and forbidden for organization roles.';

COMMENT ON COLUMN membership_role_bindings.created_by_membership_id IS 'Membership that created the binding.';

COMMENT ON COLUMN membership_role_bindings.expires_at IS 'Optional instant after which the binding grants no authorization.';
