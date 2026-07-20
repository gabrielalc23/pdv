DROP TABLE IF EXISTS auth_refresh_tokens;
DROP FUNCTION IF EXISTS validate_refresh_token_lineage();

DROP TRIGGER IF EXISTS trg_auth_sessions_validate_status_transition
ON auth_sessions;
DROP TRIGGER IF EXISTS trg_auth_sessions_validate_context ON auth_sessions;
DROP TRIGGER IF EXISTS trg_auth_sessions_touch_updated_at ON auth_sessions;
DROP TABLE IF EXISTS auth_sessions;

DROP TRIGGER IF EXISTS trg_organization_memberships_prevent_user_change
ON organization_memberships;

DROP FUNCTION IF EXISTS validate_auth_session_status_transition();
DROP FUNCTION IF EXISTS validate_auth_session_context();
DROP FUNCTION IF EXISTS prevent_membership_user_change();

DROP TYPE IF EXISTS auth_context_kind;
DROP TYPE IF EXISTS auth_session_status;
