package sessions

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// ResolveRefreshSessionID verifies an opaque refresh token without mutating it.
// It exists so HTTP can validate session-bound CSRF before rotation starts.
func (s *Service) ResolveRefreshSessionID(ctx context.Context, rawToken string) (pgtype.UUID, error) {
	parsed, err := s.codec.Parse(rawToken)
	if err != nil {
		return pgtype.UUID{}, fmt.Errorf("parse refresh token: %w", err)
	}
	token, err := s.store.GetRefreshTokenForUpdate(ctx, parsed.Selector)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgtype.UUID{}, fmt.Errorf("%w: token not found", ErrRefreshTokenInvalid)
		}
		return pgtype.UUID{}, fmt.Errorf("%w: resolve refresh token: %w", ErrDependencyUnavailable, err)
	}
	if !s.codec.VerifySecret(parsed.Secret, token.SecretHash) {
		return pgtype.UUID{}, fmt.Errorf("%w: invalid secret", ErrRefreshTokenInvalid)
	}
	if token.ExpiresAt.Valid && !s.clock.Now().Before(token.ExpiresAt.Time) {
		return pgtype.UUID{}, fmt.Errorf("%w: token expired", ErrRefreshTokenExpired)
	}
	return token.SessionID, nil
}
