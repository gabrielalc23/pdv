package memberships

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gabrielalc23/pdv/internal/platform/database"
)

type SessionRevoker interface {
	RevokeMembershipSessions(context.Context, TxStore, pgtype.UUID, pgtype.UUID, pgtype.UUID, string) ([]pgtype.UUID, error)
}

type CacheInvalidator interface {
	InvalidateSession(context.Context, pgtype.UUID)
	InvalidateMembershipAuthorizationVersion(context.Context, pgtype.UUID)
}

type sessionRevoker struct{}

func NewSessionRevoker() SessionRevoker { return sessionRevoker{} }

func (sessionRevoker) RevokeMembershipSessions(ctx context.Context, tx TxStore, organizationID, membershipID, userID pgtype.UUID, reason string) ([]pgtype.UUID, error) {
	ids, err := tx.ListSessionIDsForMembership(ctx, database.ListSessionIDsForMembershipParams{
		OrganizationID: organizationID,
		MembershipID:   membershipID,
	})
	if err != nil {
		return nil, fmt.Errorf("list membership sessions: %w", err)
	}
	for _, id := range ids {
		_, err := tx.RevokeSession(ctx, database.RevokeSessionParams{
			SessionID:    id,
			UserID:       userID,
			RevokeReason: pgtype.Text{String: reason, Valid: true},
		})
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("revoke membership session: %w", err)
		}
		if _, err := tx.RevokeSessionRefreshTokens(ctx, id); err != nil {
			return nil, fmt.Errorf("revoke membership refresh tokens: %w", err)
		}
	}
	return ids, nil
}
