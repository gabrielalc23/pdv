DROP TRIGGER IF EXISTS trg_security_audit_events_prevent_truncate
ON security_audit_events;
DROP TRIGGER IF EXISTS trg_security_audit_events_prevent_mutation
ON security_audit_events;
DROP TRIGGER IF EXISTS trg_security_audit_events_validate_actor
ON security_audit_events;

DROP TABLE IF EXISTS security_audit_events;
DROP FUNCTION IF EXISTS prevent_security_audit_event_mutation();
DROP FUNCTION IF EXISTS validate_security_audit_actor();
DROP TYPE IF EXISTS audit_outcome;
