package authn

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gabrielalc23/pdv/internal/platform/database"
)

type sessionState struct {
	SessionID                        pgtype.UUID
	SessionStatus                    database.AuthSessionStatus
	UserID                           pgtype.UUID
	ClientID                         string
	ContextKind                      database.AuthContextKind
	OrganizationID                   pgtype.UUID
	MembershipID                     pgtype.UUID
	StoreID                          pgtype.UUID
	IdleExpiresAt                    time.Time
	AbsoluteExpiresAt                time.Time
	LastSeenAt                       time.Time
	UserStatus                       database.UserStatus
	PasswordVersion                  int64
	OrganizationStatus               database.NullOrganizationStatus
	OrganizationAuthorizationVersion pgtype.Int8
	MembershipStatus                 database.NullMembershipStatus
	MembershipAuthorizationVersion   pgtype.Int8
	StoreStatus                      database.NullStoreStatus
}

func sessionStateFromRow(row database.GetAuthSessionStateRow) sessionState {
	return sessionState{
		SessionID:                        row.ID,
		SessionStatus:                    row.Status,
		UserID:                           row.UserID,
		ClientID:                         row.ClientID,
		ContextKind:                      row.ContextKind,
		OrganizationID:                   row.CurrentOrganizationID,
		MembershipID:                     row.CurrentMembershipID,
		StoreID:                          row.CurrentStoreID,
		IdleExpiresAt:                    row.IdleExpiresAt.Time,
		AbsoluteExpiresAt:                row.AbsoluteExpiresAt.Time,
		LastSeenAt:                       row.LastSeenAt.Time,
		UserStatus:                       row.UserStatus,
		PasswordVersion:                  row.PasswordVersion,
		OrganizationStatus:               row.OrganizationStatus,
		OrganizationAuthorizationVersion: row.OrganizationAuthorizationVersion,
		MembershipStatus:                 row.MembershipStatus,
		MembershipAuthorizationVersion:   row.MembershipAuthorizationVersion,
		StoreStatus:                      row.StoreStatus,
	}
}

type cachedSessionPayload struct {
	Version       int    `json:"version"`
	Status        string `json:"status"`
	UserID        string `json:"user_id"`
	ContextKind   string `json:"context_kind"`
	OrgID         string `json:"org_id,omitempty"`
	MembershipID  string `json:"membership_id,omitempty"`
	StoreID       string `json:"store_id,omitempty"`
	IdleExpiresAt int64  `json:"idle_expires_at"`
	AbsExpiresAt  int64  `json:"abs_expires_at"`
	PasswordVer   int64  `json:"pv"`
	OrgAuthVer    *int64 `json:"oav,omitempty"`
	MemAuthVer    *int64 `json:"mav,omitempty"`
}
