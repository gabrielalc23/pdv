package sessions

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/gabrielalc23/pdv/internal/platform/requestmeta"
)

func (s *Service) RevokeCurrentSession(ctx context.Context, userID, sessionID pgtype.UUID, reqMeta requestmeta.RequestMetadata) (RevokeResult, error) {
	var result RevokeResult

	err := s.provider.WithTx(ctx, func(q Querier) error {
		row, err := q.RevokeSession(ctx, database.RevokeSessionParams{
			SessionID:    sessionID,
			UserID:       userID,
			RevokeReason: pgtype.Text{String: "user_logged_out", Valid: true},
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrSessionNotFound
			}
			return fmt.Errorf("%w: revoke session: %w", ErrDependencyUnavailable, err)
		}

		_, err = q.RevokeSessionRefreshTokens(ctx, sessionID)
		if err != nil {
			return fmt.Errorf("%w: revoke refresh tokens: %w", ErrDependencyUnavailable, err)
		}

		_ = writeAuditEvent(ctx, q, "auth.logged_out", database.AuditOutcomeSUCCESS, map[string]any{
			"reason": "user_logged_out",
		}, reqMeta)

		result = RevokeResult{
			Session: sessionFromDBRow(database.AuthSession{
				ID:           row.ID,
				UserID:       row.UserID,
				Status:       row.Status,
				RevokedAt:    pgtype.Timestamptz{Time: row.RevokedAt.Time, Valid: row.RevokedAt.Valid},
				RevokeReason: row.RevokeReason,
			}),
			MustClearCookies: true,
		}

		return nil
	})

	return result, err
}

func (s *Service) RevokeSession(ctx context.Context, actorUserID, targetSessionID pgtype.UUID, reqMeta requestmeta.RequestMetadata) (RevokeResult, error) {
	var result RevokeResult

	err := s.provider.WithTx(ctx, func(q Querier) error {
		session, err := q.GetAuthSessionByID(ctx, targetSessionID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrSessionNotFound
			}
			return fmt.Errorf("%w: get session: %w", ErrDependencyUnavailable, err)
		}

		if session.UserID != actorUserID {
			return ErrSessionNotFound
		}

		if session.Status != database.AuthSessionStatusACTIVE {
			result = RevokeResult{
				Session:          sessionFromDBRow(session),
				MustClearCookies: session.ID == actorUserID,
			}
			return nil
		}

		row, err := q.RevokeSession(ctx, database.RevokeSessionParams{
			SessionID:    targetSessionID,
			UserID:       actorUserID,
			RevokeReason: pgtype.Text{String: "user_revoked", Valid: true},
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrSessionNotFound
			}
			return fmt.Errorf("%w: revoke session: %w", ErrDependencyUnavailable, err)
		}

		_, err = q.RevokeSessionRefreshTokens(ctx, targetSessionID)
		if err != nil {
			return fmt.Errorf("%w: revoke refresh tokens: %w", ErrDependencyUnavailable, err)
		}

		_ = writeAuditEvent(ctx, q, "session.revoked", database.AuditOutcomeSUCCESS, map[string]any{
			"reason":          "user_revoked",
			"session_id":      uuidStr(targetSessionID),
			"previous_status": "ACTIVE",
			"new_status":      "REVOKED",
		}, reqMeta)

		isCurrent := targetSessionID == actorUserID
		result = RevokeResult{
			Session: sessionFromDBRow(database.AuthSession{
				ID:           row.ID,
				UserID:       row.UserID,
				Status:       row.Status,
				RevokedAt:    pgtype.Timestamptz{Time: row.RevokedAt.Time, Valid: row.RevokedAt.Valid},
				RevokeReason: row.RevokeReason,
			}),
			MustClearCookies: isCurrent,
		}

		return nil
	})

	return result, err
}

func (s *Service) RevokeAllUserSessions(ctx context.Context, userID, currentSessionID pgtype.UUID, reqMeta requestmeta.RequestMetadata) (RevokeAllResult, error) {
	var result RevokeAllResult

	err := s.provider.WithTx(ctx, func(q Querier) error {
		rows, err := q.RevokeAllUserSessions(ctx, database.RevokeAllUserSessionsParams{
			UserID:       userID,
			RevokeReason: pgtype.Text{String: "user_logged_out_all", Valid: true},
		})
		if err != nil {
			return fmt.Errorf("%w: revoke all user sessions: %w", ErrDependencyUnavailable, err)
		}

		for _, row := range rows {
			_, revokeErr := q.RevokeSessionRefreshTokens(ctx, row.ID)
			if revokeErr != nil {
				return fmt.Errorf("%w: revoke refresh tokens for session %s: %w", ErrDependencyUnavailable, uuidStr(row.ID), revokeErr)
			}
			result.AffectedSessionIDs = append(result.AffectedSessionIDs, row.ID)
		}

		currentRevoked := false
		for _, id := range result.AffectedSessionIDs {
			if id == currentSessionID {
				currentRevoked = true
				break
			}
		}

		_ = writeAuditEvent(ctx, q, "auth.logged_out_all", database.AuditOutcomeSUCCESS, map[string]any{
			"reason":                 "user_logged_out_all",
			"affected_session_count": len(rows),
		}, reqMeta)

		result.MustClearCookies = currentRevoked

		return nil
	})

	return result, err
}
