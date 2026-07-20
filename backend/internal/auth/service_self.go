package auth

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gabrielalc23/pdv/internal/platform/requestmeta"
)

func (s *Service) Me(ctx context.Context, sessionID pgtype.UUID) (MeResponse, error) {
	state, err := s.loadState(ctx, sessionID)
	if err != nil {
		return MeResponse{}, err
	}
	mapped := mapAuthResponse(state, "", 0)
	return MeResponse{User: mapped.User, Session: mapped.Session, Context: mapped.Context}, nil
}

func (s *Service) ListSessions(ctx context.Context, userID, currentSessionID pgtype.UUID) (SessionsResponse, error) {
	rows, err := s.sessions.ListUserSessions(ctx, userID, currentSessionID)
	if err != nil {
		return SessionsResponse{}, err
	}
	data := make([]SessionListItem, 0, len(rows))
	for _, row := range rows {
		data = append(data, SessionListItem{ID: row.ID, ClientID: row.ClientID, DeviceName: row.DeviceName, IPAddress: row.IPAddress, UserAgent: row.UserAgent, IsCurrent: row.IsCurrent, Status: row.Status, LastSeenAt: row.LastSeenAt, CreatedAt: row.CreatedAt, IdleExpiresAt: row.IdleExpiresAt, AbsoluteExpiresAt: row.AbsoluteExpiresAt})
	}
	return SessionsResponse{Data: data}, nil
}

func (s *Service) RevokeOwnSession(ctx context.Context, userID, currentSessionID, targetSessionID pgtype.UUID, meta requestmeta.RequestMetadata) (bool, error) {
	result, err := s.sessions.RevokeSessionWithCurrent(ctx, userID, targetSessionID, currentSessionID, meta)
	if err != nil {
		return false, err
	}
	return result.MustClearCookies, nil
}
