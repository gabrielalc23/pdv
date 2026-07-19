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
	return s.revokeSession(ctx, userID, sessionID, sessionID, "user_logged_out", "auth.logged_out", reqMeta)
}

// RevokeSession preserves the original API. Call RevokeSessionWithCurrent when
// the caller needs MustClearCookies to identify the current session.
func (s *Service) RevokeSession(ctx context.Context, actorUserID, targetSessionID pgtype.UUID, reqMeta requestmeta.RequestMetadata) (RevokeResult, error) {
	return s.RevokeSessionWithCurrent(ctx, actorUserID, targetSessionID, pgtype.UUID{}, reqMeta)
}

func (s *Service) RevokeSessionWithCurrent(ctx context.Context, actorUserID, targetSessionID, currentSessionID pgtype.UUID, reqMeta requestmeta.RequestMetadata) (RevokeResult, error) {
	return s.revokeSession(ctx, actorUserID, targetSessionID, currentSessionID, "user_revoked", "session.revoked", reqMeta)
}

func (s *Service) revokeSession(
	ctx context.Context,
	actorUserID pgtype.UUID,
	targetSessionID pgtype.UUID,
	currentSessionID pgtype.UUID,
	reason string,
	eventType string,
	reqMeta requestmeta.RequestMetadata,
) (RevokeResult, error) {
	var result RevokeResult

	err := s.provider.WithTx(ctx, func(q Querier) error {
		row, err := q.GetAuthSessionForUpdate(ctx, targetSessionID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("%w: target session does not exist", ErrSessionNotFound)
			}
			return fmt.Errorf("%w: get target session: %w", ErrDependencyUnavailable, err)
		}
		if row.UserID != actorUserID {
			return fmt.Errorf("%w: target session does not exist", ErrSessionNotFound)
		}

		session := sessionFromDBRow(authSessionFromForUpdate(row))
		previousStatus := row.Status
		statusChanged := row.Status == database.AuthSessionStatusACTIVE
		if statusChanged {
			revoked, err := q.RevokeSession(ctx, database.RevokeSessionParams{
				SessionID:    targetSessionID,
				UserID:       actorUserID,
				RevokeReason: pgtype.Text{String: reason, Valid: true},
			})
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return fmt.Errorf("%w: target session is not active", ErrSessionNotFound)
				}
				return fmt.Errorf("%w: revoke session: %w", ErrDependencyUnavailable, err)
			}

			session.Status = string(revoked.Status)
			session.RevokedAt = revoked.RevokedAt
			session.RevokeReason = revoked.RevokeReason
			if revoked.UpdatedAt.Valid {
				session.UpdatedAt = revoked.UpdatedAt.Time
			}
		}

		if _, err := q.RevokeSessionRefreshTokens(ctx, targetSessionID); err != nil {
			return fmt.Errorf("%w: revoke refresh tokens: %w", ErrDependencyUnavailable, err)
		}

		if statusChanged {
			if err := writeAuditEventWithSubject(ctx, q, eventType, database.AuditOutcomeSUCCESS, map[string]any{
				"reason":          reason,
				"session_id":      uuidStr(targetSessionID),
				"previous_status": string(previousStatus),
				"new_status":      string(database.AuthSessionStatusREVOKED),
			}, reqMeta, auditSubject{UserID: row.UserID, MembershipID: row.CurrentMembershipID, OrganizationID: row.CurrentOrganizationID, StoreID: row.CurrentStoreID, SessionID: row.ID}); err != nil {
				return fmt.Errorf("%w: write session revoke audit event: %w", ErrDependencyUnavailable, err)
			}
		}

		result = RevokeResult{
			Session:          session,
			MustClearCookies: targetSessionID == currentSessionID,
		}
		return nil
	})
	if err != nil {
		return RevokeResult{}, fmt.Errorf("revoke session transaction: %w", err)
	}

	s.invalidateSession(ctx, targetSessionID)
	return result, nil
}

func (s *Service) RevokeAllUserSessions(ctx context.Context, userID, currentSessionID pgtype.UUID, reqMeta requestmeta.RequestMetadata) (RevokeAllResult, error) {
	var result RevokeAllResult

	err := s.provider.WithTx(ctx, func(q Querier) error {
		ids, err := s.RevokeAllUserSessionsInTx(ctx, q, userID, "user_logged_out_all")
		if err != nil {
			return err
		}
		result.AffectedSessionIDs = ids

		for _, id := range result.AffectedSessionIDs {
			if id == currentSessionID {
				result.MustClearCookies = true
				break
			}
		}

		if err := writeAuditEventWithSubject(ctx, q, "auth.logged_out_all", database.AuditOutcomeSUCCESS, map[string]any{
			"reason":                 "user_logged_out_all",
			"affected_session_count": len(ids),
		}, reqMeta, auditSubject{UserID: userID, SessionID: currentSessionID}); err != nil {
			return fmt.Errorf("%w: write revoke-all audit event: %w", ErrDependencyUnavailable, err)
		}

		return nil
	})
	if err != nil {
		return RevokeAllResult{}, fmt.Errorf("revoke all sessions transaction: %w", err)
	}

	for _, id := range result.AffectedSessionIDs {
		s.invalidateSession(ctx, id)
	}
	return result, nil
}

// RevokeAllUserSessionsInTx composes session and refresh-token revocation into
// a caller-owned transaction. It does not commit, write audit data, or invalidate caches.
func (s *Service) RevokeAllUserSessionsInTx(ctx context.Context, q Querier, userID pgtype.UUID, reason string) ([]pgtype.UUID, error) {
	ids, err := q.RevokeAllActiveUserSessions(ctx, database.RevokeAllActiveUserSessionsParams{
		UserID:       userID,
		RevokeReason: pgtype.Text{String: reason, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("%w: revoke all user sessions: %w", ErrDependencyUnavailable, err)
	}
	if _, err := q.RevokeAllUserRefreshTokens(ctx, userID); err != nil {
		return nil, fmt.Errorf("%w: revoke all user refresh tokens: %w", ErrDependencyUnavailable, err)
	}
	return ids, nil
}
