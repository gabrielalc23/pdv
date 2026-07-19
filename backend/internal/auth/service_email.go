package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gabrielalc23/pdv/internal/audit"
	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/gabrielalc23/pdv/internal/platform/requestmeta"
)

func (s *Service) ResendVerification(ctx context.Context, input EmailActionRequest, meta requestmeta.RequestMetadata) error {
	normalizedEmail, err := validateActionEmail(input.Email)
	if err != nil {
		return err
	}
	user, err := s.store.Queries.GetUserForActionByNormalizedEmail(ctx, normalizedEmail)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("%w: look up verification user: %w", ErrDependencyUnavailable, err)
	}
	if user.Status != database.UserStatusACTIVE || user.EmailVerifiedAt.Valid {
		return nil
	}

	var rawToken string
	err = s.store.WithTx(ctx, func(tx *database.Tx) error {
		q := tx.Queries
		if _, err := q.LockUserForActionTokenChange(ctx, user.ID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil
			}
			return fmt.Errorf("lock verification user: %w", err)
		}
		current, err := q.GetUserForActionByNormalizedEmail(ctx, normalizedEmail)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil
			}
			return fmt.Errorf("reload verification user: %w", err)
		}
		if current.Status != database.UserStatusACTIVE || current.EmailVerifiedAt.Valid {
			return nil
		}
		rawToken, err = s.createActionTokenInTx(ctx, q, current.ID, ActionTokenPurposeEmailVerification, emailVerificationTTL, meta)
		if err != nil {
			return err
		}
		user = current
		return nil
	})
	if err != nil {
		return fmt.Errorf("%w: resend verification: %w", ErrDependencyUnavailable, err)
	}
	if rawToken != "" {
		link := s.mailLinks.BuildEmailVerification(rawToken)
		mailCtx, cancel := mailDeliveryContext(ctx)
		defer cancel()
		if err := s.mailer.SendEmailVerification(mailCtx, user.Email, user.DisplayName, link); err != nil {
			slog.Error("verification email delivery failed", "component", "mailer", "message_type", "verification")
		}
	}
	return nil
}

func (s *Service) VerifyEmail(ctx context.Context, input EmailVerifyRequest, meta requestmeta.RequestMetadata) error {
	parsed, err := s.actionTokens.Parse(input.Token, ActionTokenPurposeEmailVerification)
	if err != nil {
		return ErrInvalidRequest
	}
	err = s.store.WithTx(ctx, func(tx *database.Tx) error {
		q := tx.Queries
		userID, err := q.LockActionTokenOwner(ctx, database.LockActionTokenOwnerParams{ID: parsed.Selector, Purpose: ActionTokenPurposeEmailVerification})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrInvalidRequest
			}
			return fmt.Errorf("lock verification token owner: %w", err)
		}
		token, err := q.GetActionTokenForUpdate(ctx, database.GetActionTokenForUpdateParams{ID: parsed.Selector, Purpose: ActionTokenPurposeEmailVerification})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrInvalidRequest
			}
			return fmt.Errorf("lock verification token: %w", err)
		}
		if !s.actionTokens.VerifySecret(parsed.Secret, token.SecretHash) || token.UserID != userID {
			return ErrInvalidRequest
		}
		user, err := q.GetUserByID(ctx, userID)
		if err != nil {
			return fmt.Errorf("load verification user: %w", err)
		}
		if token.ConsumedAt.Valid {
			verifiedByToken, err := q.HasSecurityAuditEventForEntity(ctx, database.HasSecurityAuditEventForEntityParams{
				EventType: audit.EventAuthEmailVerified, EntityType: pgtype.Text{String: "auth_action_token", Valid: true}, EntityID: token.ID,
			})
			if err != nil {
				return fmt.Errorf("check verification idempotency: %w", err)
			}
			if user.EmailVerifiedAt.Valid && verifiedByToken {
				return nil
			}
			return ErrInvalidRequest
		}
		if !token.ExpiresAt.Valid || !s.clock.Now().Before(token.ExpiresAt.Time) {
			return ErrActionTokenExpired
		}
		if err := mapUserStatus(user.Status); err != nil {
			return err
		}
		if _, err := q.VerifyUserEmail(ctx, user.ID); err != nil {
			return fmt.Errorf("mark email verified: %w", err)
		}
		if _, err := q.ConsumeActionToken(ctx, database.ConsumeActionTokenParams{ID: token.ID, UserID: user.ID, Purpose: ActionTokenPurposeEmailVerification}); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrInvalidRequest
			}
			return fmt.Errorf("consume verification token: %w", err)
		}
		if _, err := q.InvalidatePreviousActionTokens(ctx, database.InvalidatePreviousActionTokensParams{UserID: user.ID, Purpose: ActionTokenPurposeEmailVerification}); err != nil {
			return fmt.Errorf("invalidate remaining verification tokens: %w", err)
		}
		metadata := audit.NewMetadata()
		metadata.Set("requested_via", "public")
		if err := s.writeAuditForEntity(ctx, q, audit.EventAuthEmailVerified, database.AuditOutcomeSUCCESS, user.ID, pgtype.UUID{}, pgtype.UUID{}, pgtype.UUID{}, pgtype.UUID{}, pgtype.Text{String: "auth_action_token", Valid: true}, token.ID, meta, metadata); err != nil {
			return fmt.Errorf("write email verified audit: %w", err)
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, ErrInvalidRequest) || errors.Is(err, ErrActionTokenExpired) || errors.Is(err, ErrUserSuspended) || errors.Is(err, ErrUserDisabled) {
			return err
		}
		return fmt.Errorf("%w: verify email: %w", ErrDependencyUnavailable, err)
	}
	return nil
}
