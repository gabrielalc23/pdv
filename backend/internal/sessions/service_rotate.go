package sessions

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gabrielalc23/pdv/internal/platform/database"
)

func (s *Service) RotateRefreshToken(ctx context.Context, input RotateInput) (RotateResult, error) {
	parsed, err := s.codec.Parse(input.RawRefreshToken)
	if err != nil {
		return RotateResult{}, err
	}

	var result RotateResult

	err = s.provider.WithTx(ctx, func(q Querier) error {
		token, err := q.GetRefreshTokenForUpdate(ctx, parsed.Selector)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("%w: token not found", ErrRefreshTokenInvalid)
			}
			return fmt.Errorf("%w: %w", ErrDependencyUnavailable, err)
		}

		if !s.codec.VerifySecret(parsed.Secret, token.SecretHash) {
			return fmt.Errorf("%w: invalid secret", ErrRefreshTokenInvalid)
		}

		if token.RevokedAt.Valid {
			return fmt.Errorf("%w: token is revoked", ErrRefreshTokenInvalid)
		}

		if token.ExpiresAt.Valid && s.clock.Now().After(token.ExpiresAt.Time) {
			return fmt.Errorf("%w: token expired", ErrRefreshTokenExpired)
		}

		if token.ConsumedAt.Valid {
			return s.handleReuse(ctx, q, token, input)
		}

		sessionRow, err := q.GetAuthSessionForUpdate(ctx, token.SessionID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("%w: session not found", ErrSessionNotFound)
			}
			return fmt.Errorf("%w: %w", ErrDependencyUnavailable, err)
		}

		if sessionRow.Status != database.AuthSessionStatusACTIVE {
			return fmt.Errorf("%w: session status is %s", errFromSessionStatus(sessionRow.Status), sessionRow.Status)
		}

		now := s.clock.Now()
		if sessionRow.IdleExpiresAt.Valid && now.After(sessionRow.IdleExpiresAt.Time) {
			return fmt.Errorf("%w: session idle expired", ErrSessionExpired)
		}
		if sessionRow.AbsoluteExpiresAt.Valid && now.After(sessionRow.AbsoluteExpiresAt.Time) {
			return fmt.Errorf("%w: session absolute expired", ErrSessionExpired)
		}

		newIdleExpiresAt := now.Add(s.cfg.RefreshIdleTTL)
		if sessionRow.AbsoluteExpiresAt.Valid && newIdleExpiresAt.After(sessionRow.AbsoluteExpiresAt.Time) {
			newIdleExpiresAt = sessionRow.AbsoluteExpiresAt.Time
		}

		childRaw, childHash, err := s.codec.Generate(token.SessionID)
		if err != nil {
			return fmt.Errorf("generate child token: %w", err)
		}

		child, err := q.CreateRefreshToken(ctx, database.CreateRefreshTokenParams{
			SessionID:     token.SessionID,
			ParentTokenID: token.ID,
			SecretHash:    childHash,
			ExpiresAt:     sessionRow.AbsoluteExpiresAt,
		})
		if err != nil {
			return fmt.Errorf("create child refresh token: %w", err)
		}

		_, err = q.ConsumeAndReplaceRefreshToken(ctx, database.ConsumeAndReplaceRefreshTokenParams{
			ReplacedByTokenID: child.ID,
			SessionID:         token.SessionID,
			ID:                token.ID,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("%w: token already consumed", ErrRefreshTokenReused)
			}
			return fmt.Errorf("consume and replace refresh token: %w", err)
		}

		_, _ = q.TouchSession(ctx, database.TouchSessionParams{
			SessionID:     token.SessionID,
			UserID:        sessionRow.UserID,
			IdleExpiresAt: pgtype.Timestamptz{Time: newIdleExpiresAt, Valid: true},
		})

		meta := map[string]any{
			"previous_token_id": uuidStr(token.ID),
			"new_token_id":      uuidStr(child.ID),
		}
		_ = writeAuditEvent(ctx, q, "auth.refreshed", database.AuditOutcomeSUCCESS, meta, input.RequestMeta)

		result = RotateResult{
			Session:          sessionFromDBRow(authSessionFromForUpdate(sessionRow)),
			RawRefreshToken:  childRaw,
			ExpiresIn:        time.Until(newIdleExpiresAt),
			MustClearCookies: false,
		}

		return nil
	})

	return result, err
}

func (s *Service) handleReuse(ctx context.Context, q Querier, token database.AuthRefreshToken, input RotateInput) error {
	_, err := q.MarkSessionCompromised(ctx, database.MarkSessionCompromisedParams{
		SessionID:    token.SessionID,
		RevokeReason: pgtype.Text{String: "refresh_token_reused", Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("%w: session not found for reuse handling", ErrSessionNotFound)
		}
		return fmt.Errorf("%w: mark session compromised: %w", ErrDependencyUnavailable, err)
	}

	_, err = q.RevokeSessionRefreshTokens(ctx, token.SessionID)
	if err != nil {
		return fmt.Errorf("%w: revoke session refresh tokens: %w", ErrDependencyUnavailable, err)
	}

	_ = writeAuditEvent(ctx, q, "auth.refresh.reused", database.AuditOutcomeFAILURE, map[string]any{
		"token_id":        uuidStr(token.ID),
		"session_id":      uuidStr(token.SessionID),
		"previous_status": "ACTIVE",
		"new_status":      "COMPROMISED",
	}, input.RequestMeta)

	return fmt.Errorf("%w: token was already consumed", ErrRefreshTokenReused)
}

func errFromSessionStatus(status database.AuthSessionStatus) error {
	switch status {
	case database.AuthSessionStatusREVOKED:
		return ErrSessionRevoked
	case database.AuthSessionStatusCOMPROMISED:
		return ErrSessionCompromised
	case database.AuthSessionStatusEXPIRED:
		return ErrSessionExpired
	default:
		return ErrSessionRevoked
	}
}

func authSessionFromForUpdate(row database.GetAuthSessionForUpdateRow) database.AuthSession {
	return database.AuthSession{
		ID:                    row.ID,
		UserID:                row.UserID,
		Status:                row.Status,
		ClientID:              row.ClientID,
		DeviceName:            row.DeviceName,
		UserAgent:             row.UserAgent,
		IpAddress:             row.IpAddress,
		ContextKind:           row.ContextKind,
		CurrentOrganizationID: row.CurrentOrganizationID,
		CurrentMembershipID:   row.CurrentMembershipID,
		CurrentStoreID:        row.CurrentStoreID,
		IdleExpiresAt:         row.IdleExpiresAt,
		AbsoluteExpiresAt:     row.AbsoluteExpiresAt,
		LastSeenAt:            row.LastSeenAt,
		RevokedAt:             row.RevokedAt,
		RevokeReason:          row.RevokeReason,
		CreatedAt:             row.CreatedAt,
		UpdatedAt:             row.UpdatedAt,
	}
}
