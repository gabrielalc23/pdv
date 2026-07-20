package auth

import (
	"context"
	"fmt"
	"net/netip"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/gabrielalc23/pdv/internal/platform/requestmeta"
)

const (
	passwordResetTTL     = 30 * time.Minute
	emailVerificationTTL = 24 * time.Hour
)

func (s *Service) createActionTokenInTx(ctx context.Context, q database.Querier, userID pgtype.UUID, purpose ActionTokenPurpose, ttl time.Duration, meta requestmeta.RequestMetadata) (string, error) {
	if _, err := q.InvalidatePreviousActionTokens(ctx, database.InvalidatePreviousActionTokensParams{UserID: userID, Purpose: purpose}); err != nil {
		return "", fmt.Errorf("invalidate previous action tokens: %w", err)
	}
	selector, err := randomUUID()
	if err != nil {
		return "", fmt.Errorf("generate action token selector: %w", err)
	}
	rawToken, secretHash, err := s.actionTokens.Generate(purpose, selector)
	if err != nil {
		return "", fmt.Errorf("generate action token: %w", err)
	}
	var requestedIP *netip.Addr
	if ip, parseErr := netip.ParseAddr(meta.ClientIP); parseErr == nil {
		requestedIP = &ip
	}
	row, err := q.CreateActionToken(ctx, database.CreateActionTokenParams{
		ID: selector, UserID: userID, Purpose: purpose, SecretHash: secretHash,
		ExpiresAt: pgtype.Timestamptz{Time: s.clock.Now().Add(ttl), Valid: true}, RequestedIp: requestedIP,
	})
	if err != nil {
		return "", fmt.Errorf("create action token: %w", err)
	}
	if row.ID != selector || len(row.SecretHash) != 32 {
		return "", fmt.Errorf("create action token: persisted token invariant failed")
	}
	return rawToken, nil
}
