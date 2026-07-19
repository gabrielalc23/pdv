package organizations

import (
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gabrielalc23/pdv/internal/platform/database"
)

func mapOrganization(row database.Organization) OrganizationResponse {
	return OrganizationResponse{
		ID:                   uuidString(row.ID),
		Name:                 row.Name,
		Slug:                 row.Slug,
		Status:               string(row.Status),
		Timezone:             row.Timezone,
		Locale:               row.Locale,
		Currency:             row.Currency,
		AuthorizationVersion: row.AuthorizationVersion,
		ArchivedAt:           optionalTime(row.ArchivedAt),
		CreatedAt:            row.CreatedAt.Time,
		UpdatedAt:            row.UpdatedAt.Time,
	}
}

func mapMembership(row database.ListUserActiveMembershipsRow) OrganizationMembershipResponse {
	organization := OrganizationSummaryResponse{
		ID:                   uuidString(row.OrganizationID),
		Name:                 row.OrganizationName,
		Slug:                 row.OrganizationSlug,
		Status:               string(row.OrganizationStatus),
		Timezone:             row.OrganizationTimezone,
		Locale:               row.OrganizationLocale,
		Currency:             row.OrganizationCurrency,
		AuthorizationVersion: row.OrganizationAuthorizationVersion,
	}
	response := OrganizationMembershipResponse{
		MembershipID:                   uuidString(row.MembershipID),
		MembershipStatus:               string(row.MembershipStatus),
		MembershipAuthorizationVersion: row.MembershipAuthorizationVersion,
		JoinedAt:                       row.JoinedAt.Time,
		Organization:                   organization,
	}
	if row.DefaultStoreID.Valid {
		response.DefaultStore = &StoreResponse{
			ID:     uuidString(row.DefaultStoreID),
			Code:   row.DefaultStoreCode.String,
			Name:   row.DefaultStoreName.String,
			Status: string(row.DefaultStoreStatus.StoreStatus),
		}
	}
	return response
}

func mapStore(row database.Store) StoreResponse {
	return StoreResponse{
		ID:       uuidString(row.ID),
		Code:     row.Code,
		Name:     row.Name,
		Status:   string(row.Status),
		Timezone: row.Timezone,
	}
}

func optionalTime(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	t := value.Time
	return &t
}

func uuidString(id pgtype.UUID) string {
	if !id.Valid {
		return ""
	}
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", id.Bytes[0:4], id.Bytes[4:6], id.Bytes[6:8], id.Bytes[8:10], id.Bytes[10:16])
}
