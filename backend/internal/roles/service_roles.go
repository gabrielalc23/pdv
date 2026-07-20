package roles

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gabrielalc23/pdv/internal/audit"
	"github.com/gabrielalc23/pdv/internal/platform/authcontext"
	"github.com/gabrielalc23/pdv/internal/platform/authz"
	"github.com/gabrielalc23/pdv/internal/platform/database"
)

func (s *Service) CreateRole(ctx context.Context, actor authcontext.Principal, input UpsertRoleInput) (RoleResponse, error) {
	if err := requireActorScope(actor, authz.ScopeRolesCreate); err != nil {
		return RoleResponse{}, err
	}
	normalized, err := normalizeRoleInput(input)
	if err != nil {
		return RoleResponse{}, err
	}

	var response RoleResponse
	var sessionIDs []pgtype.UUID
	err = s.txProvider.WithTx(ctx, func(tx TxStore) error {
		if err := validateRoleScopes(ctx, tx, actor, normalized.assignmentScope, normalized.scopes); err != nil {
			return err
		}
		role, err := tx.CreateRole(ctx, database.CreateRoleParams{
			OrganizationID:        actor.OrganizationID,
			Key:                   normalized.key,
			Name:                  normalized.name,
			Description:           normalized.description,
			AssignmentScope:       normalized.assignmentScope,
			IsSystem:              false,
			IsMutable:             true,
			CreatedByMembershipID: actor.MembershipID,
		})
		if err != nil {
			return translatePersistenceError(err)
		}
		if err := tx.ReplaceRoleScopes(ctx, actor.OrganizationID, role.ID, normalized.scopes); err != nil {
			return translatePersistenceError(err)
		}
		if err := tx.IncrementOrganizationAuthorizationVersion(ctx, actor.OrganizationID); err != nil {
			return fmt.Errorf("%w: increment organization authorization version: %w", ErrDependencyUnavailable, err)
		}
		sessionIDs, err = tx.ListSessionIDsForOrganization(ctx, actor.OrganizationID)
		if err != nil {
			return fmt.Errorf("%w: list affected sessions: %w", ErrDependencyUnavailable, err)
		}
		if err := writeAudit(ctx, tx, actor, audit.EventRoleCreated, "role", role.ID, audit.Metadata{
			"key":              role.Key,
			"assignment_scope": string(role.AssignmentScope),
			"scope_count":      len(normalized.scopes),
		}); err != nil {
			return err
		}
		response = mapRole(role.ID, role.Key, role.Name, role.Description, role.AssignmentScope, role.IsSystem, role.IsMutable, role.IsActive, normalized.scopes, role.CreatedAt, role.UpdatedAt)
		return nil
	})
	if err != nil {
		return RoleResponse{}, fmt.Errorf("create role transaction: %w", err)
	}
	s.invalidateSessions(ctx, sessionIDs)
	s.invalidateOrganizationVersion(ctx, actor.OrganizationID)
	return response, nil
}

func (s *Service) UpdateRole(ctx context.Context, actor authcontext.Principal, rawRoleID string, input UpsertRoleInput) (RoleResponse, error) {
	if err := requireActorScope(actor, authz.ScopeRolesUpdate); err != nil {
		return RoleResponse{}, err
	}
	roleID, err := parseUUID(rawRoleID, "roleId")
	if err != nil {
		return RoleResponse{}, err
	}
	normalized, err := normalizeRoleInput(input)
	if err != nil {
		return RoleResponse{}, err
	}

	var response RoleResponse
	var sessionIDs []pgtype.UUID
	err = s.txProvider.WithTx(ctx, func(tx TxStore) error {
		current, err := tx.GetRoleWithScopes(ctx, actor.OrganizationID, roleID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrRoleNotFound
			}
			return fmt.Errorf("%w: get role: %w", ErrDependencyUnavailable, err)
		}
		if current.IsSystem || !current.IsMutable {
			return ErrSystemRoleImmutable
		}
		if normalized.key != current.Key {
			return validationError("key", "cannot be changed after role creation")
		}
		if normalized.assignmentScope != current.AssignmentScope {
			return validationError("assignmentScope", "cannot be changed after role creation")
		}
		if err := tx.LockRoleForScopeChange(ctx, actor.OrganizationID, roleID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrSystemRoleImmutable
			}
			return fmt.Errorf("%w: lock role: %w", ErrDependencyUnavailable, err)
		}
		current, err = tx.GetRoleWithScopes(ctx, actor.OrganizationID, roleID)
		if err != nil {
			return fmt.Errorf("%w: reload locked role: %w", ErrDependencyUnavailable, err)
		}
		if err := validateRoleScopes(ctx, tx, actor, current.AssignmentScope, normalized.scopes); err != nil {
			return err
		}

		unchanged := current.Name == normalized.name && current.Description == normalized.description && slices.Equal(current.ScopeCodes, normalized.scopes)
		if unchanged {
			response = mapRole(current.ID, current.Key, current.Name, current.Description, current.AssignmentScope, current.IsSystem, current.IsMutable, current.IsActive, current.ScopeCodes, current.CreatedAt, current.UpdatedAt)
			return nil
		}
		updated, err := tx.UpdateRole(ctx, database.UpdateRoleParams{Name: normalized.name, Description: normalized.description, OrganizationID: actor.OrganizationID, RoleID: roleID})
		if err != nil {
			return translatePersistenceError(err)
		}
		if err := tx.ReplaceRoleScopes(ctx, actor.OrganizationID, roleID, normalized.scopes); err != nil {
			return translatePersistenceError(err)
		}
		if err := tx.IncrementOrganizationAuthorizationVersion(ctx, actor.OrganizationID); err != nil {
			return fmt.Errorf("%w: increment organization authorization version: %w", ErrDependencyUnavailable, err)
		}
		sessionIDs, err = tx.ListSessionIDsForOrganization(ctx, actor.OrganizationID)
		if err != nil {
			return fmt.Errorf("%w: list affected sessions: %w", ErrDependencyUnavailable, err)
		}
		if err := writeAudit(ctx, tx, actor, audit.EventRoleUpdated, "role", roleID, audit.Metadata{
			"key":                  current.Key,
			"previous_scope_count": len(current.ScopeCodes),
			"scope_count":          len(normalized.scopes),
		}); err != nil {
			return err
		}
		response = mapRole(updated.ID, updated.Key, updated.Name, updated.Description, updated.AssignmentScope, updated.IsSystem, updated.IsMutable, updated.IsActive, normalized.scopes, updated.CreatedAt, updated.UpdatedAt)
		return nil
	})
	if err != nil {
		return RoleResponse{}, fmt.Errorf("update role transaction: %w", err)
	}
	s.invalidateSessions(ctx, sessionIDs)
	s.invalidateOrganizationVersion(ctx, actor.OrganizationID)
	return response, nil
}

func (s *Service) ActivateRole(ctx context.Context, actor authcontext.Principal, rawRoleID string) (RoleResponse, error) {
	return s.setRoleActive(ctx, actor, rawRoleID, true)
}

func (s *Service) DeactivateRole(ctx context.Context, actor authcontext.Principal, rawRoleID string) (RoleResponse, error) {
	return s.setRoleActive(ctx, actor, rawRoleID, false)
}

func (s *Service) setRoleActive(ctx context.Context, actor authcontext.Principal, rawRoleID string, active bool) (RoleResponse, error) {
	if err := requireActorScope(actor, authz.ScopeRolesStatusUpdate); err != nil {
		return RoleResponse{}, err
	}
	roleID, err := parseUUID(rawRoleID, "roleId")
	if err != nil {
		return RoleResponse{}, err
	}
	var response RoleResponse
	var sessionIDs []pgtype.UUID
	err = s.txProvider.WithTx(ctx, func(tx TxStore) error {
		current, err := tx.GetRoleWithScopes(ctx, actor.OrganizationID, roleID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrRoleNotFound
			}
			return fmt.Errorf("%w: get role: %w", ErrDependencyUnavailable, err)
		}
		if current.IsSystem || !current.IsMutable {
			return ErrSystemRoleImmutable
		}
		if err := tx.LockRoleForScopeChange(ctx, actor.OrganizationID, roleID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrSystemRoleImmutable
			}
			return fmt.Errorf("%w: lock role: %w", ErrDependencyUnavailable, err)
		}
		current, err = tx.GetRoleWithScopes(ctx, actor.OrganizationID, roleID)
		if err != nil {
			return fmt.Errorf("%w: reload locked role: %w", ErrDependencyUnavailable, err)
		}
		if current.IsActive == active {
			response = mapRole(current.ID, current.Key, current.Name, current.Description, current.AssignmentScope, current.IsSystem, current.IsMutable, current.IsActive, current.ScopeCodes, current.CreatedAt, current.UpdatedAt)
			return nil
		}
		if active {
			if err := ensureGrantSubset(ctx, tx, actor, current.ScopeCodes); err != nil {
				return err
			}
		}
		updated, err := tx.UpdateRoleStatus(ctx, database.UpdateRoleStatusParams{IsActive: active, OrganizationID: actor.OrganizationID, RoleID: roleID})
		if err != nil {
			return translatePersistenceError(err)
		}
		if err := tx.IncrementOrganizationAuthorizationVersion(ctx, actor.OrganizationID); err != nil {
			return fmt.Errorf("%w: increment organization authorization version: %w", ErrDependencyUnavailable, err)
		}
		sessionIDs, err = tx.ListSessionIDsForOrganization(ctx, actor.OrganizationID)
		if err != nil {
			return fmt.Errorf("%w: list affected sessions: %w", ErrDependencyUnavailable, err)
		}
		if err := writeAudit(ctx, tx, actor, audit.EventRoleStatusChanged, "role", roleID, audit.Metadata{
			"key":             current.Key,
			"previous_active": current.IsActive,
			"active":          active,
		}); err != nil {
			return err
		}
		response = mapRole(updated.ID, updated.Key, updated.Name, updated.Description, updated.AssignmentScope, updated.IsSystem, updated.IsMutable, updated.IsActive, current.ScopeCodes, updated.CreatedAt, updated.UpdatedAt)
		return nil
	})
	if err != nil {
		return RoleResponse{}, fmt.Errorf("update role status transaction: %w", err)
	}
	s.invalidateSessions(ctx, sessionIDs)
	s.invalidateOrganizationVersion(ctx, actor.OrganizationID)
	return response, nil
}

func translatePersistenceError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrRoleNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.ConstraintName {
		case "roles_organization_id_key_unique":
			return ErrRoleKeyAlreadyInUse
		case "roles_immutable_update", "roles_system_designation_immutable":
			return ErrSystemRoleImmutable
		case "role_scopes_assignment_scope_check":
			return ErrScopeLevelInvalid
		case "role_scopes_assignable_check":
			return ErrScopeNotAssignable
		case "role_binding_store_required":
			return validationError("storeId", "is required for a store role")
		case "role_binding_store_forbidden":
			return validationError("storeId", "must be null for an organization role")
		case "owner_binding_expiration_forbidden":
			return validationError("expiresAt", "owner binding cannot expire")
		}
	}
	return fmt.Errorf("%w: database operation: %w", ErrDependencyUnavailable, err)
}
