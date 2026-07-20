DROP TRIGGER IF EXISTS trg_roles_validate_update ON roles;

DROP TRIGGER IF EXISTS trg_membership_role_bindings_validate_scope
ON membership_role_bindings;
DROP TABLE IF EXISTS membership_role_bindings;

DROP TRIGGER IF EXISTS trg_role_scopes_validate_assignment ON role_scopes;
DROP TABLE IF EXISTS role_scopes;

DROP FUNCTION IF EXISTS validate_role_binding_scope();
DROP FUNCTION IF EXISTS validate_role_scope_assignment();
DROP FUNCTION IF EXISTS validate_role_update();

DROP TRIGGER IF EXISTS trg_roles_touch_updated_at ON roles;
DROP TABLE IF EXISTS roles;

DROP TABLE IF EXISTS permission_scopes;

DROP TRIGGER IF EXISTS trg_organization_memberships_touch_updated_at
ON organization_memberships;
DROP TRIGGER IF EXISTS trg_organization_memberships_validate_status_transition
ON organization_memberships;
DROP TABLE IF EXISTS organization_memberships;

DROP FUNCTION IF EXISTS validate_membership_status_transition();

DROP TYPE IF EXISTS role_assignment_scope;
DROP TYPE IF EXISTS permission_scope_level;
DROP TYPE IF EXISTS membership_status;
