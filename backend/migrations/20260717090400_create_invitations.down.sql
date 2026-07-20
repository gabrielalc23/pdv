DROP TRIGGER IF EXISTS trg_roles_validate_invitation_bindings ON roles;
DROP TRIGGER IF EXISTS trg_invitation_role_bindings_require_role
ON invitation_role_bindings;
DROP TRIGGER IF EXISTS trg_organization_invitations_require_role
ON organization_invitations;
DROP TRIGGER IF EXISTS trg_invitation_role_bindings_validate_owner
ON invitation_role_bindings;
DROP TRIGGER IF EXISTS trg_invitation_role_bindings_validate_scope
ON invitation_role_bindings;
DROP TRIGGER IF EXISTS trg_organization_invitations_validate_owner_update
ON organization_invitations;

DROP TABLE IF EXISTS invitation_role_bindings;

DROP TRIGGER IF EXISTS trg_organization_invitations_touch_updated_at
ON organization_invitations;
DROP TRIGGER IF EXISTS trg_organization_invitations_validate_status_transition
ON organization_invitations;
DROP TABLE IF EXISTS organization_invitations;

DROP FUNCTION IF EXISTS ensure_invitation_has_role_binding();
DROP FUNCTION IF EXISTS validate_owner_invitation_binding();
DROP FUNCTION IF EXISTS validate_owner_invitation_update();
DROP FUNCTION IF EXISTS validate_role_update_for_invitations();
DROP FUNCTION IF EXISTS validate_invitation_status_transition();

DROP TYPE IF EXISTS invitation_status;
