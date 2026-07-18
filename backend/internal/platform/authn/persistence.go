package authn

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gabrielalc23/pdv/internal/platform/database"
)

type persistenceStore struct {
	q *database.Queries
}

func NewPersistenceStore(q *database.Queries) *persistenceStore {
	return &persistenceStore{q: q}
}

func (s *persistenceStore) getSessionState(ctx context.Context, sessionID pgtype.UUID) (sessionState, error) {
	row, err := s.q.GetAuthSessionState(ctx, sessionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sessionState{}, fmt.Errorf("%w: session not found", ErrSessionRevoked)
		}
		return sessionState{}, fmt.Errorf("%w: %w", ErrDependencyUnavailable, err)
	}
	return sessionStateFromRow(row), nil
}

func (s *persistenceStore) touchSession(ctx context.Context, sessionID, userID pgtype.UUID, idleExpiresAt time.Time) error {
	_, err := s.q.TouchSession(ctx, database.TouchSessionParams{
		SessionID:     sessionID,
		UserID:        userID,
		IdleExpiresAt: pgtype.Timestamptz{Time: idleExpiresAt, Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("touch session: %w", err)
	}
	return nil
}
