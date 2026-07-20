package sessions

import (
	"net/netip"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gabrielalc23/pdv/internal/platform/database"
)

func sessionFromDBRow(row database.AuthSession) Session {
	s := Session{
		ID:             row.ID,
		UserID:         row.UserID,
		Status:         string(row.Status),
		ClientID:       row.ClientID,
		DeviceName:     row.DeviceName,
		UserAgent:      row.UserAgent,
		IPAddress:      row.IpAddress,
		ContextKind:    string(row.ContextKind),
		OrganizationID: row.CurrentOrganizationID,
		MembershipID:   row.CurrentMembershipID,
		StoreID:        row.CurrentStoreID,
		RevokedAt:      row.RevokedAt,
		RevokeReason:   row.RevokeReason,
	}

	if row.IdleExpiresAt.Valid {
		s.IdleExpiresAt = row.IdleExpiresAt.Time
	}
	if row.AbsoluteExpiresAt.Valid {
		s.AbsoluteExpiresAt = row.AbsoluteExpiresAt.Time
	}
	if row.LastSeenAt.Valid {
		s.LastSeenAt = row.LastSeenAt.Time
	}
	if row.CreatedAt.Valid {
		s.CreatedAt = row.CreatedAt.Time
	}
	if row.UpdatedAt.Valid {
		s.UpdatedAt = row.UpdatedAt.Time
	}

	return s
}

func sessionToListItem(s Session, currentSessionID pgtype.UUID) SessionListItem {
	item := SessionListItem{
		ID:        uuidStr(s.ID),
		ClientID:  s.ClientID,
		Status:    s.Status,
		IsCurrent: s.ID == currentSessionID,
	}

	if s.DeviceName.Valid {
		item.DeviceName = s.DeviceName.String
	}
	if s.UserAgent.Valid {
		item.UserAgent = s.UserAgent.String
	}
	if s.IPAddress != nil {
		item.IPAddress = s.IPAddress.String()
	}
	if !s.LastSeenAt.IsZero() {
		item.LastSeenAt = s.LastSeenAt.Format(time.RFC3339)
	}
	if !s.CreatedAt.IsZero() {
		item.CreatedAt = s.CreatedAt.Format(time.RFC3339)
	}
	if !s.IdleExpiresAt.IsZero() {
		item.IdleExpiresAt = s.IdleExpiresAt.Format(time.RFC3339)
	}
	if !s.AbsoluteExpiresAt.IsZero() {
		item.AbsoluteExpiresAt = s.AbsoluteExpiresAt.Format(time.RFC3339)
	}

	return item
}

func uuidFromPtr(ptr *netip.Addr) pgtype.UUID {
	if ptr == nil {
		return pgtype.UUID{}
	}
	var id pgtype.UUID
	_ = id.Scan(ptr.String())
	return id
}
