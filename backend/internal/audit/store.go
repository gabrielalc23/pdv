package audit

import (
	"context"
	"net/netip"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gabrielalc23/pdv/internal/platform/database"
)

type Writer interface {
	Write(ctx context.Context, q database.Querier, event Event) error
}

type Event struct {
	OrganizationID    pgtype.UUID
	StoreID           pgtype.UUID
	ActorUserID       pgtype.UUID
	ActorMembershipID pgtype.UUID
	SessionID         pgtype.UUID
	EventType         string
	EntityType        pgtype.Text
	EntityID          pgtype.UUID
	RequestID         pgtype.Text
	IPAddress         *netip.Addr
	UserAgent         pgtype.Text
	Outcome           database.AuditOutcome
	Metadata          []byte
}
