package memberships

import (
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gabrielalc23/pdv/internal/platform/database"
)

func mapListRow(row database.ListMembershipsRow) MembershipResponse {
	response := MembershipResponse{
		ID:                   row.ID.String(),
		OrganizationID:       row.OrganizationID.String(),
		User:                 UserResponse{ID: row.UserID.String(), Email: row.Email, DisplayName: row.DisplayName, Status: string(row.UserStatus)},
		Status:               string(row.Status),
		AuthorizationVersion: row.AuthorizationVersion,
		RoleBindings:         []RoleBindingResponse{},
		JoinedAt:             row.JoinedAt.Time,
		CreatedAt:            row.CreatedAt.Time,
		UpdatedAt:            row.UpdatedAt.Time,
		SuspendedAt:          timeString(row.SuspendedAt),
		RemovedAt:            timeString(row.RemovedAt),
	}
	if row.DefaultStoreID.Valid {
		response.DefaultStore = &StoreResponse{ID: row.DefaultStoreID.String(), Code: row.DefaultStoreCode.String, Name: row.DefaultStoreName.String}
	}
	return response
}

func mapDetailRow(row database.GetMembershipForOrganizationRow, bindings []database.ListMemberRoleBindingsRow) MembershipResponse {
	response := MembershipResponse{
		ID:                   row.ID.String(),
		OrganizationID:       row.OrganizationID.String(),
		User:                 UserResponse{ID: row.UserID.String(), Email: row.Email, DisplayName: row.DisplayName, Status: string(row.UserStatus), EmailVerifiedAt: timeString(row.EmailVerifiedAt)},
		Status:               string(row.Status),
		AuthorizationVersion: row.AuthorizationVersion,
		RoleBindings:         make([]RoleBindingResponse, 0, len(bindings)),
		JoinedAt:             row.JoinedAt.Time,
		CreatedAt:            row.CreatedAt.Time,
		UpdatedAt:            row.UpdatedAt.Time,
		SuspendedAt:          timeString(row.SuspendedAt),
		RemovedAt:            timeString(row.RemovedAt),
	}
	if row.DefaultStoreID.Valid {
		response.DefaultStore = &StoreResponse{ID: row.DefaultStoreID.String(), Code: row.DefaultStoreCode.String, Name: row.DefaultStoreName.String}
		if row.DefaultStoreStatus.Valid {
			response.DefaultStore.Status = string(row.DefaultStoreStatus.StoreStatus)
		}
	}
	for _, binding := range bindings {
		item := RoleBindingResponse{
			ID:              binding.ID.String(),
			RoleID:          binding.RoleID.String(),
			RoleKey:         binding.RoleKey,
			RoleName:        binding.RoleName,
			AssignmentScope: string(binding.AssignmentScope),
			ExpiresAt:       timeString(binding.ExpiresAt),
			IsActive:        binding.IsActive,
			CreatedAt:       binding.CreatedAt.Time,
		}
		if binding.StoreID.Valid {
			item.Store = &StoreResponse{ID: binding.StoreID.String(), Code: binding.StoreCode.String, Name: binding.StoreName.String}
		}
		response.RoleBindings = append(response.RoleBindings, item)
	}
	return response
}

func timeString(value pgtype.Timestamptz) *string {
	if !value.Valid {
		return nil
	}
	formatted := value.Time.UTC().Format("2006-01-02T15:04:05.999999999Z07:00")
	return &formatted
}
