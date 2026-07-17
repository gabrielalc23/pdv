CREATE TYPE auth_session_status AS ENUM (
    'ACTIVE',
    'REVOKED',
    'COMPROMISED',
    'EXPIRED'
);

CREATE TYPE auth_context_kind AS ENUM (
    'IDENTITY',
    'ORGANIZATION',
    'STORE'
);

CREATE TABLE auth_sessions (
    id UUID DEFAULT uuidv7() NOT NULL,
    user_id UUID NOT NULL,
    status auth_session_status DEFAULT 'ACTIVE' NOT NULL,
    client_id VARCHAR(50) NOT NULL,
    device_name VARCHAR(150),
    user_agent VARCHAR(512),
    ip_address INET,
    context_kind auth_context_kind DEFAULT 'IDENTITY' NOT NULL,
    current_organization_id UUID,
    current_membership_id UUID,
    current_store_id UUID,
    idle_expires_at TIMESTAMPTZ NOT NULL,
    absolute_expires_at TIMESTAMPTZ NOT NULL,
    last_seen_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    revoked_at TIMESTAMPTZ,
    revoke_reason VARCHAR(100),
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    CONSTRAINT auth_sessions_pkey PRIMARY KEY (id),
    CONSTRAINT auth_sessions_user_id_fkey FOREIGN KEY (user_id)
        REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT auth_sessions_current_membership_fkey
        FOREIGN KEY (current_organization_id, current_membership_id)
        REFERENCES organization_memberships (organization_id, id)
        ON DELETE RESTRICT,
    CONSTRAINT auth_sessions_current_store_fkey
        FOREIGN KEY (current_organization_id, current_store_id)
        REFERENCES stores (organization_id, id) ON DELETE RESTRICT,
    CONSTRAINT auth_sessions_client_id_not_blank CHECK (BTRIM(client_id) <> ''),
    CONSTRAINT auth_sessions_device_name_not_blank CHECK (
        device_name IS NULL OR BTRIM(device_name) <> ''
    ),
    CONSTRAINT auth_sessions_user_agent_not_blank CHECK (
        user_agent IS NULL OR BTRIM(user_agent) <> ''
    ),
    CONSTRAINT auth_sessions_context_check CHECK (
        (
            context_kind = 'IDENTITY'
            AND current_organization_id IS NULL
            AND current_membership_id IS NULL
            AND current_store_id IS NULL
        )
        OR (
            context_kind = 'ORGANIZATION'
            AND current_organization_id IS NOT NULL
            AND current_membership_id IS NOT NULL
            AND current_store_id IS NULL
        )
        OR (
            context_kind = 'STORE'
            AND current_organization_id IS NOT NULL
            AND current_membership_id IS NOT NULL
            AND current_store_id IS NOT NULL
        )
    ),
    CONSTRAINT auth_sessions_idle_expires_at_check
        CHECK (idle_expires_at > created_at),
    CONSTRAINT auth_sessions_absolute_expires_at_check
        CHECK (absolute_expires_at > created_at),
    CONSTRAINT auth_sessions_expiration_order_check
        CHECK (idle_expires_at <= absolute_expires_at),
    CONSTRAINT auth_sessions_status_check CHECK (
        (status = 'ACTIVE' AND revoked_at IS NULL AND revoke_reason IS NULL)
        OR (status <> 'ACTIVE' AND revoked_at IS NOT NULL)
    ),
    CONSTRAINT auth_sessions_revoke_reason_not_blank CHECK (
        revoke_reason IS NULL OR BTRIM(revoke_reason) <> ''
    ),
    CONSTRAINT auth_sessions_last_seen_at_check
        CHECK (last_seen_at >= created_at),
    CONSTRAINT auth_sessions_revoked_at_check CHECK (
        revoked_at IS NULL OR revoked_at >= created_at
    ),
    CONSTRAINT auth_sessions_updated_at_check CHECK (updated_at >= created_at)
);

CREATE INDEX idx_auth_sessions_user_status_last_seen_at
ON auth_sessions (user_id, status, last_seen_at DESC);

CREATE INDEX idx_auth_sessions_current_membership_status
ON auth_sessions (
    current_organization_id,
    current_membership_id,
    status
);

CREATE INDEX idx_auth_sessions_current_store
ON auth_sessions (current_organization_id, current_store_id);

CREATE INDEX idx_auth_sessions_absolute_expires_at
ON auth_sessions (absolute_expires_at);

CREATE INDEX idx_auth_sessions_idle_expires_at
ON auth_sessions (idle_expires_at);

CREATE TRIGGER trg_auth_sessions_touch_updated_at
BEFORE UPDATE ON auth_sessions
FOR EACH ROW
EXECUTE FUNCTION touch_updated_at();

CREATE FUNCTION validate_auth_session_context()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF NEW.context_kind <> 'IDENTITY'
       AND NEW.current_organization_id IS NOT NULL
       AND NEW.current_membership_id IS NOT NULL
       AND NOT EXISTS (
        SELECT 1
        FROM organization_memberships
        WHERE organization_id = NEW.current_organization_id
          AND id = NEW.current_membership_id
          AND user_id = NEW.user_id
    ) THEN
        RAISE EXCEPTION 'the session membership must belong to the session user'
            USING ERRCODE = '23514',
                  CONSTRAINT = 'auth_sessions_user_membership_check';
    END IF;

    RETURN NEW;
END;
$$;

CREATE TRIGGER trg_auth_sessions_validate_context
BEFORE INSERT OR UPDATE OF user_id, context_kind, current_organization_id, current_membership_id, current_store_id
ON auth_sessions
FOR EACH ROW
EXECUTE FUNCTION validate_auth_session_context();

CREATE FUNCTION validate_auth_session_status_transition()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF OLD.status <> 'ACTIVE' AND NEW.status <> OLD.status THEN
        RAISE EXCEPTION 'a terminal session status cannot transition'
            USING ERRCODE = '23514',
                  CONSTRAINT = 'auth_sessions_status_terminal';
    END IF;

    RETURN NEW;
END;
$$;

CREATE TRIGGER trg_auth_sessions_validate_status_transition
BEFORE UPDATE OF status ON auth_sessions
FOR EACH ROW
EXECUTE FUNCTION validate_auth_session_status_transition();

CREATE FUNCTION prevent_membership_user_change()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF NEW.user_id IS DISTINCT FROM OLD.user_id THEN
        RAISE EXCEPTION 'a membership cannot be reassigned to another user'
            USING ERRCODE = '23514',
                  CONSTRAINT = 'organization_memberships_user_immutable';
    END IF;

    RETURN NEW;
END;
$$;

CREATE TRIGGER trg_organization_memberships_prevent_user_change
BEFORE UPDATE OF user_id ON organization_memberships
FOR EACH ROW
EXECUTE FUNCTION prevent_membership_user_change();

COMMENT ON TABLE auth_sessions IS 'Stores server-authoritative user sessions and their active authorization context.';
COMMENT ON COLUMN auth_sessions.context_kind IS 'Active identity, organization, or store context represented by access tokens.';
COMMENT ON COLUMN auth_sessions.current_organization_id IS 'Organization selected for tenant contexts.';
COMMENT ON COLUMN auth_sessions.current_membership_id IS 'Membership selected for tenant contexts.';
COMMENT ON COLUMN auth_sessions.current_store_id IS 'Store selected only for store context.';
COMMENT ON COLUMN auth_sessions.idle_expires_at IS 'Sliding inactivity deadline capped by absolute_expires_at.';
COMMENT ON COLUMN auth_sessions.absolute_expires_at IS 'Maximum lifetime of the session.';
COMMENT ON COLUMN auth_sessions.revoked_at IS 'Terminal-state timestamp for revoked, compromised, or expired sessions.';

CREATE TABLE auth_refresh_tokens (
    id UUID DEFAULT uuidv7() NOT NULL,
    session_id UUID NOT NULL,
    parent_token_id UUID,
    replaced_by_token_id UUID,
    secret_hash BYTEA NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    consumed_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    CONSTRAINT auth_refresh_tokens_pkey PRIMARY KEY (id),
    CONSTRAINT auth_refresh_tokens_session_id_id_unique UNIQUE (session_id, id),
    CONSTRAINT auth_refresh_tokens_session_id_fkey FOREIGN KEY (session_id)
        REFERENCES auth_sessions (id) ON DELETE CASCADE,
    CONSTRAINT auth_refresh_tokens_parent_token_id_fkey
        FOREIGN KEY (session_id, parent_token_id)
        REFERENCES auth_refresh_tokens (session_id, id) ON DELETE RESTRICT,
    CONSTRAINT auth_refresh_tokens_replaced_by_token_id_fkey
        FOREIGN KEY (session_id, replaced_by_token_id)
        REFERENCES auth_refresh_tokens (session_id, id) ON DELETE RESTRICT,
    CONSTRAINT auth_refresh_tokens_secret_hash_length CHECK (
        OCTET_LENGTH(secret_hash) = 32
    ),
    CONSTRAINT auth_refresh_tokens_expires_at_check
        CHECK (expires_at > created_at),
    CONSTRAINT auth_refresh_tokens_parent_not_self CHECK (
        parent_token_id IS NULL OR parent_token_id <> id
    ),
    CONSTRAINT auth_refresh_tokens_replacement_not_self CHECK (
        replaced_by_token_id IS NULL OR replaced_by_token_id <> id
    ),
    CONSTRAINT auth_refresh_tokens_consumed_at_check CHECK (
        consumed_at IS NULL OR consumed_at >= created_at
    ),
    CONSTRAINT auth_refresh_tokens_revoked_at_check CHECK (
        revoked_at IS NULL OR revoked_at >= created_at
    )
);

CREATE UNIQUE INDEX idx_auth_refresh_tokens_secret_hash
ON auth_refresh_tokens (secret_hash);

CREATE INDEX idx_auth_refresh_tokens_session_created_at
ON auth_refresh_tokens (session_id, created_at DESC);

CREATE UNIQUE INDEX idx_auth_refresh_tokens_parent_unique
ON auth_refresh_tokens (parent_token_id)
WHERE parent_token_id IS NOT NULL;

CREATE UNIQUE INDEX idx_auth_refresh_tokens_replacement_unique
ON auth_refresh_tokens (replaced_by_token_id)
WHERE replaced_by_token_id IS NOT NULL;

CREATE INDEX idx_auth_refresh_tokens_active
ON auth_refresh_tokens (session_id, expires_at)
WHERE consumed_at IS NULL AND revoked_at IS NULL;

CREATE FUNCTION validate_refresh_token_lineage()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    affected_ids UUID[];
BEGIN
    IF TG_OP = 'INSERT' THEN
        affected_ids := ARRAY[NEW.id];
    ELSIF TG_OP = 'DELETE' THEN
        affected_ids := ARRAY[OLD.id];
    ELSE
        affected_ids := ARRAY[OLD.id, NEW.id];
    END IF;

    IF EXISTS (
        SELECT 1
        FROM auth_refresh_tokens AS child
        WHERE (
            child.id = ANY(affected_ids)
            OR child.parent_token_id = ANY(affected_ids)
        )
          AND child.parent_token_id IS NOT NULL
          AND NOT EXISTS (
              SELECT 1
              FROM auth_refresh_tokens AS parent
              WHERE parent.session_id = child.session_id
                AND parent.id = child.parent_token_id
                AND parent.replaced_by_token_id = child.id
          )
    ) OR EXISTS (
        SELECT 1
        FROM auth_refresh_tokens AS parent
        WHERE (
            parent.id = ANY(affected_ids)
            OR parent.replaced_by_token_id = ANY(affected_ids)
        )
          AND parent.replaced_by_token_id IS NOT NULL
          AND NOT EXISTS (
              SELECT 1
              FROM auth_refresh_tokens AS child
              WHERE child.session_id = parent.session_id
                AND child.id = parent.replaced_by_token_id
                AND child.parent_token_id = parent.id
          )
    ) THEN
        RAISE EXCEPTION 'refresh token parent and replacement links must be reciprocal'
            USING ERRCODE = '23514',
                  CONSTRAINT = 'auth_refresh_tokens_lineage_consistency';
    END IF;

    RETURN NULL;
END;
$$;

CREATE CONSTRAINT TRIGGER trg_auth_refresh_tokens_validate_lineage
AFTER INSERT OR UPDATE OR DELETE ON auth_refresh_tokens
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW
EXECUTE FUNCTION validate_refresh_token_lineage();

COMMENT ON TABLE auth_refresh_tokens IS 'Stores selectors and HMAC-SHA-256 hashes for opaque rotating refresh tokens.';
COMMENT ON COLUMN auth_refresh_tokens.id IS 'UUIDv7 selector included in the external refresh token.';
COMMENT ON COLUMN auth_refresh_tokens.parent_token_id IS 'Token from which this token was rotated.';
COMMENT ON COLUMN auth_refresh_tokens.replaced_by_token_id IS 'Child token that replaced this token after consumption.';
COMMENT ON COLUMN auth_refresh_tokens.secret_hash IS 'Thirty-two-byte HMAC of the secret; raw refresh tokens are never stored.';
COMMENT ON COLUMN auth_refresh_tokens.consumed_at IS 'Timestamp of the one permitted rotation.';
COMMENT ON COLUMN auth_refresh_tokens.revoked_at IS 'Timestamp at which the token was explicitly revoked.';
