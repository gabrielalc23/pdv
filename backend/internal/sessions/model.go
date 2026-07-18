package sessions

import (
	"net/netip"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type Session struct {
	ID                pgtype.UUID
	UserID            pgtype.UUID
	Status            string
	ClientID          string
	DeviceName        pgtype.Text
	UserAgent         pgtype.Text
	IPAddress         *netip.Addr
	ContextKind       string
	OrganizationID    pgtype.UUID
	MembershipID      pgtype.UUID
	StoreID           pgtype.UUID
	IdleExpiresAt     time.Time
	AbsoluteExpiresAt time.Time
	LastSeenAt        time.Time
	RevokedAt         pgtype.Timestamptz
	RevokeReason      pgtype.Text
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type SessionListItem struct {
	ID                string `json:"id"`
	ClientID          string `json:"clientId"`
	DeviceName        string `json:"deviceName,omitempty"`
	IPAddress         string `json:"ipAddress,omitempty"`
	UserAgent         string `json:"userAgent,omitempty"`
	IsCurrent         bool   `json:"isCurrent"`
	Status            string `json:"status"`
	LastSeenAt        string `json:"lastSeenAt"`
	CreatedAt         string `json:"createdAt"`
	IdleExpiresAt     string `json:"idleExpiresAt"`
	AbsoluteExpiresAt string `json:"absoluteExpiresAt"`
}
