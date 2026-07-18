package sessions

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s *Service) ListUserSessions(ctx context.Context, userID, currentSessionID pgtype.UUID) ([]SessionListItem, error) {
	rows, err := s.store.ListUserSessions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("%w: list sessions: %w", ErrDependencyUnavailable, err)
	}

	items := make([]SessionListItem, 0, len(rows))
	for _, row := range rows {
		session := sessionFromDBRow(row)
		item := sessionToListItem(session, currentSessionID)
		items = append(items, item)
	}

	if len(items) == 0 {
		return items, nil
	}

	return items, nil
}

func (s *Service) GetSessionByID(ctx context.Context, id pgtype.UUID) (Session, error) {
	row, err := s.store.GetAuthSessionByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Session{}, ErrSessionNotFound
		}
		return Session{}, fmt.Errorf("%w: get session: %w", ErrDependencyUnavailable, err)
	}

	return sessionFromDBRow(row), nil
}
