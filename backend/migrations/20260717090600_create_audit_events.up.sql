CREATE TYPE audit_outcome AS ENUM (
    'SUCCESS',
    'FAILURE'
);

CREATE TABLE security_audit_events (
    id UUID DEFAULT uuidv7() NOT NULL,
    organization_id UUID,
    store_id UUID,
    actor_user_id UUID,
    actor_membership_id UUID,
    session_id UUID,
    event_type VARCHAR(100) NOT NULL,
    entity_type VARCHAR(80),
    entity_id UUID,
    request_id VARCHAR(100),
    ip_address INET,
    user_agent VARCHAR(512),
    outcome audit_outcome NOT NULL,
    metadata JSONB DEFAULT '{}'::jsonb NOT NULL,
    occurred_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    CONSTRAINT security_audit_events_pkey PRIMARY KEY (id),
    CONSTRAINT security_audit_events_organization_id_fkey
        FOREIGN KEY (organization_id)
        REFERENCES organizations (id) ON DELETE RESTRICT,
    CONSTRAINT security_audit_events_store_fkey
        FOREIGN KEY (organization_id, store_id)
        REFERENCES stores (organization_id, id) ON DELETE RESTRICT,
    CONSTRAINT security_audit_events_actor_user_id_fkey
        FOREIGN KEY (actor_user_id)
        REFERENCES users (id) ON DELETE RESTRICT,
    CONSTRAINT security_audit_events_actor_membership_fkey
        FOREIGN KEY (organization_id, actor_membership_id)
        REFERENCES organization_memberships (organization_id, id)
        ON DELETE RESTRICT,
    CONSTRAINT security_audit_events_store_context_check
        CHECK (store_id IS NULL OR organization_id IS NOT NULL),
    CONSTRAINT security_audit_events_membership_context_check
        CHECK (actor_membership_id IS NULL OR organization_id IS NOT NULL),
    CONSTRAINT security_audit_events_event_type_not_blank
        CHECK (BTRIM(event_type) <> ''),
    CONSTRAINT security_audit_events_entity_type_not_blank CHECK (
        entity_type IS NULL OR BTRIM(entity_type) <> ''
    ),
    CONSTRAINT security_audit_events_request_id_not_blank CHECK (
        request_id IS NULL OR BTRIM(request_id) <> ''
    ),
    CONSTRAINT security_audit_events_user_agent_not_blank CHECK (
        user_agent IS NULL OR BTRIM(user_agent) <> ''
    ),
    CONSTRAINT security_audit_events_metadata_object CHECK (
        jsonb_typeof(metadata) = 'object'
    )
);

CREATE INDEX idx_security_audit_events_organization_occurred_at ON security_audit_events (
    organization_id,
    occurred_at DESC
);

CREATE INDEX idx_security_audit_events_organization_store_occurred_at ON security_audit_events (
    organization_id,
    store_id,
    occurred_at DESC
);

CREATE INDEX idx_security_audit_events_actor_user_occurred_at ON security_audit_events (
    actor_user_id,
    occurred_at DESC
);

CREATE INDEX idx_security_audit_events_event_type_occurred_at ON security_audit_events (event_type, occurred_at DESC);

CREATE INDEX idx_security_audit_events_entity ON security_audit_events (entity_type, entity_id);

CREATE INDEX idx_security_audit_events_actor_membership_fkey ON security_audit_events (
    organization_id,
    actor_membership_id
);

CREATE FUNCTION validate_security_audit_actor()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF NEW.actor_membership_id IS NOT NULL AND NOT EXISTS (
        SELECT 1
        FROM organization_memberships
        WHERE organization_id = NEW.organization_id
          AND id = NEW.actor_membership_id
          AND user_id = NEW.actor_user_id
    ) THEN
        RAISE EXCEPTION 'the audit membership must belong to the actor user'
            USING ERRCODE = '23514',
                  CONSTRAINT = 'security_audit_events_actor_membership_user_check';
    END IF;

    RETURN NEW;
END;
$$;

CREATE TRIGGER trg_security_audit_events_validate_actor
BEFORE INSERT ON security_audit_events
FOR EACH ROW
EXECUTE FUNCTION validate_security_audit_actor();

CREATE FUNCTION prevent_security_audit_event_mutation()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    RAISE EXCEPTION 'security audit events are append-only'
        USING ERRCODE = '23514',
              CONSTRAINT = 'security_audit_events_immutable';
END;
$$;

CREATE TRIGGER trg_security_audit_events_prevent_mutation
BEFORE UPDATE OR DELETE ON security_audit_events
FOR EACH ROW
EXECUTE FUNCTION prevent_security_audit_event_mutation();

CREATE TRIGGER trg_security_audit_events_prevent_truncate
BEFORE TRUNCATE ON security_audit_events
FOR EACH STATEMENT
EXECUTE FUNCTION prevent_security_audit_event_mutation();

COMMENT ON
TABLE security_audit_events IS 'Append-only security and authorization event history.';

COMMENT ON COLUMN security_audit_events.organization_id IS 'Optional tenant context; null for identity and pre-authentication events.';

COMMENT ON COLUMN security_audit_events.store_id IS 'Optional store context within the organization.';

COMMENT ON COLUMN security_audit_events.actor_user_id IS 'Global user that initiated the event when known.';

COMMENT ON COLUMN security_audit_events.actor_membership_id IS 'Tenant membership that initiated the event when known.';

COMMENT ON COLUMN security_audit_events.session_id IS 'Historical session identifier retained without a foreign key so audit survives session cleanup.';

COMMENT ON COLUMN security_audit_events.event_type IS 'Stable application-defined event code.';

COMMENT ON COLUMN security_audit_events.metadata IS 'Non-sensitive structured metadata; tokens, credentials, cookies, and secrets are forbidden.';
