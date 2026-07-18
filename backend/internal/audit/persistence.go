package audit

import (
	"context"
	"fmt"

	"github.com/gabrielalc23/pdv/internal/platform/database"
)

type writerImpl struct{}

func NewWriter() Writer {
	return &writerImpl{}
}

func (w *writerImpl) Write(ctx context.Context, q database.Querier, event Event) error {
	if event.EventType == "" {
		return fmt.Errorf("%w: event_type is required", ErrWriteFailed)
	}

	params := database.CreateAuditEventParams{
		OrganizationID:    event.OrganizationID,
		StoreID:           event.StoreID,
		ActorUserID:       event.ActorUserID,
		ActorMembershipID: event.ActorMembershipID,
		SessionID:         event.SessionID,
		EventType:         event.EventType,
		EntityType:        event.EntityType,
		EntityID:          event.EntityID,
		RequestID:         event.RequestID,
		IpAddress:         event.IPAddress,
		UserAgent:         event.UserAgent,
		Outcome:           event.Outcome,
		Metadata:          event.Metadata,
	}

	if _, err := q.CreateAuditEvent(ctx, params); err != nil {
		return fmt.Errorf("%w: %w", ErrWriteFailed, err)
	}
	return nil
}
