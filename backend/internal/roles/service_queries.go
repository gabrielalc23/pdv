package roles

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/gabrielalc23/pdv/internal/platform/authcontext"
	"github.com/gabrielalc23/pdv/internal/platform/authz"
)

func (s *Service) ListScopes(ctx context.Context, actor authcontext.Principal) (ScopeListResponse, error) {
	if err := requireActorScope(actor, authz.ScopeScopesRead); err != nil {
		return ScopeListResponse{}, err
	}
	rows, err := s.store.ListPermissionScopes(ctx)
	if err != nil {
		return ScopeListResponse{}, fmt.Errorf("%w: list permission scopes: %w", ErrDependencyUnavailable, err)
	}
	response := ScopeListResponse{Data: make([]ScopeResponse, 0, len(rows))}
	for _, row := range rows {
		response.Data = append(response.Data, mapScope(row))
	}
	return response, nil
}

func (s *Service) ListRoles(ctx context.Context, actor authcontext.Principal) (RoleListResponse, error) {
	if err := requireActorScope(actor, authz.ScopeRolesRead); err != nil {
		return RoleListResponse{}, err
	}
	rows, err := s.store.ListRolesWithScopes(ctx, actor.OrganizationID)
	if err != nil {
		return RoleListResponse{}, fmt.Errorf("%w: list roles: %w", ErrDependencyUnavailable, err)
	}
	response := RoleListResponse{Data: make([]RoleResponse, 0, len(rows))}
	for _, row := range rows {
		response.Data = append(response.Data, mapRole(row.ID, row.Key, row.Name, row.Description, row.AssignmentScope, row.IsSystem, row.IsMutable, row.IsActive, row.ScopeCodes, row.CreatedAt, row.UpdatedAt))
	}
	return response, nil
}

func (s *Service) GetRole(ctx context.Context, actor authcontext.Principal, rawRoleID string) (RoleResponse, error) {
	if err := requireActorScope(actor, authz.ScopeRolesRead); err != nil {
		return RoleResponse{}, err
	}
	roleID, err := parseUUID(rawRoleID, "roleId")
	if err != nil {
		return RoleResponse{}, err
	}
	row, err := s.store.GetRoleWithScopes(ctx, actor.OrganizationID, roleID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return RoleResponse{}, ErrRoleNotFound
		}
		return RoleResponse{}, fmt.Errorf("%w: get role: %w", ErrDependencyUnavailable, err)
	}
	return mapRole(row.ID, row.Key, row.Name, row.Description, row.AssignmentScope, row.IsSystem, row.IsMutable, row.IsActive, row.ScopeCodes, row.CreatedAt, row.UpdatedAt), nil
}
