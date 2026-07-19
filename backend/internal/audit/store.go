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

type ReadStore interface {
	ListAuditEvents(ctx context.Context, params database.ListAuditEventsParams) ([]database.SecurityAuditEvent, error)
	CountAuditEvents(ctx context.Context, params database.CountAuditEventsParams) (int64, error)
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

type readStore struct {
	q database.Querier
}

func NewReadStore(q database.Querier) ReadStore {
	return &readStore{q: q}
}

func (s *readStore) ListAuditEvents(ctx context.Context, params database.ListAuditEventsParams) ([]database.SecurityAuditEvent, error) {
	return s.q.ListAuditEvents(ctx, params)
}

func (s *readStore) CountAuditEvents(ctx context.Context, params database.CountAuditEventsParams) (int64, error) {
	return s.q.CountAuditEvents(ctx, params)
}
