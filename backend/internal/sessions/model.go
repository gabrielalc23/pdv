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

