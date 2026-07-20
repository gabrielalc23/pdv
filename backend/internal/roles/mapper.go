package roles

import (
	"slices"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gabrielalc23/pdv/internal/platform/database"
)

func mapScope(row database.PermissionScope) ScopeResponse {
	return ScopeResponse{
		Code:         row.Code,
		Resource:     row.Resource,
		Action:       row.Action,
		ScopeLevel:   string(row.ScopeLevel),
		Description:  row.Description,
		IsAssignable: row.IsAssignable,
		CreatedAt:    timestamp(row.CreatedAt),
	}
}

func mapRole(id pgtype.UUID, key, name string, description pgtype.Text, assignmentScope database.RoleAssignmentScope, isSystem, isMutable, isActive bool, scopes []string, createdAt, updatedAt pgtype.Timestamptz) RoleResponse {
	response := RoleResponse{
		ID:              id.String(),
		Key:             key,
		Name:            name,
		AssignmentScope: string(assignmentScope),
		IsSystem:        isSystem,
		IsMutable:       isMutable,
		IsActive:        isActive,
		Scopes:          slices.Clone(scopes),
		CreatedAt:       timestamp(createdAt),
		UpdatedAt:       timestamp(updatedAt),
	}
	if response.Scopes == nil {
		response.Scopes = []string{}
	}
	if description.Valid {
		value := description.String
		response.Description = &value
	}
	return response
}

func mapBinding(id, membershipID pgtype.UUID, role database.GetRoleWithScopesForOrganizationRow, store *database.Store, expiresAt, createdAt pgtype.Timestamptz) BindingResponse {
	response := BindingResponse{
		ID:           id.String(),
		MembershipID: membershipID.String(),
		Role: BindingRoleResponse{
			ID:              role.ID.String(),
			Key:             role.Key,
			Name:            role.Name,
			AssignmentScope: string(role.AssignmentScope),
		},
		ExpiresAt: nullableTimestamp(expiresAt),
		CreatedAt: timestamp(createdAt),
	}
	if store != nil {
		response.Store = &BindingStoreResponse{ID: store.ID.String(), Code: store.Code, Name: store.Name}
	}
	return response
}

func timestamp(value pgtype.Timestamptz) time.Time {
	if !value.Valid {
		return time.Time{}
	}
	return value.Time.UTC()
}

func nullableTimestamp(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	result := value.Time.UTC()
	return &result
}
