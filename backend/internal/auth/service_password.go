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
	"github.com/gabrielalc23/pdv/internal/platform/password"
	"github.com/gabrielalc23/pdv/internal/platform/ratelimit"
	"github.com/gabrielalc23/pdv/internal/platform/requestmeta"
)

func (s *Service) ForgotPassword(ctx context.Context, input EmailActionRequest, meta requestmeta.RequestMetadata) error {
	normalizedEmail, err := validateActionEmail(input.Email)
	if err != nil {
		return err
	}
	fingerprint := ratelimit.Fingerprint(s.cfg.RateLimitKey, normalizedEmail)
	user, lookupErr := s.store.Queries.GetUserForActionByNormalizedEmail(ctx, normalizedEmail)
	if lookupErr != nil && !errors.Is(lookupErr, pgx.ErrNoRows) {
		return fmt.Errorf("%w: look up password reset user: %w", ErrDependencyUnavailable, lookupErr)
	}

	eligible := lookupErr == nil && user.Status == database.UserStatusACTIVE && user.HasPassword
	if !eligible {
		return s.auditGenericPasswordResetRequest(ctx, fingerprint, meta)
	}

	var rawToken string
	err = s.store.WithTx(ctx, func(tx *database.Tx) error {
		q := tx.Queries
		if _, err := q.LockUserForActionTokenChange(ctx, user.ID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil
			}
			return fmt.Errorf("lock password reset user: %w", err)
		}
		current, err := q.GetUserForActionByNormalizedEmail(ctx, normalizedEmail)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil
			}
			return fmt.Errorf("reload password reset user: %w", err)
		}
		if current.Status != database.UserStatusACTIVE || !current.HasPassword {
			return nil
		}
		rawToken, err = s.createActionTokenInTx(ctx, q, current.ID, ActionTokenPurposePasswordReset, passwordResetTTL, meta)
		if err != nil {
			return err
		}
		metadata := audit.NewMetadata()
		metadata.Set("email_fingerprint", fingerprint)
		metadata.Set("requested_via", "public")
		if err := s.writeAudit(ctx, q, audit.EventAuthPasswordResetReq, database.AuditOutcomeSUCCESS, current.ID, pgtype.UUID{}, pgtype.UUID{}, pgtype.UUID{}, pgtype.UUID{}, meta, metadata); err != nil {
			return fmt.Errorf("write password reset request audit: %w", err)
		}
		user = current
		return nil
	})
	if err != nil {
		return fmt.Errorf("%w: request password reset: %w", ErrDependencyUnavailable, err)
	}
	if rawToken != "" {
		link := s.mailLinks.BuildPasswordReset(rawToken)
		mailCtx, cancel := mailDeliveryContext(ctx)
		defer cancel()
		if err := s.mailer.SendPasswordReset(mailCtx, user.Email, user.DisplayName, link); err != nil {
			slog.Error("password reset email delivery failed", "component", "mailer", "message_type", "password_reset")
		}
	}
	return nil
}

func (s *Service) auditGenericPasswordResetRequest(ctx context.Context, fingerprint string, meta requestmeta.RequestMetadata) error {
	err := s.store.WithTx(ctx, func(tx *database.Tx) error {
		metadata := audit.NewMetadata()
		metadata.Set("email_fingerprint", fingerprint)
		metadata.Set("requested_via", "public")
		return s.writeAudit(ctx, tx.Queries, audit.EventAuthPasswordResetReq, database.AuditOutcomeSUCCESS, pgtype.UUID{}, pgtype.UUID{}, pgtype.UUID{}, pgtype.UUID{}, pgtype.UUID{}, meta, metadata)
	})
	if err != nil {
		return fmt.Errorf("%w: write generic password reset audit: %w", ErrDependencyUnavailable, err)
	}
	return nil
}

func (s *Service) ResetPassword(ctx context.Context, input PasswordResetRequest, meta requestmeta.RequestMetadata) error {
	parsed, err := s.actionTokens.Parse(input.Token, ActionTokenPurposePasswordReset)
	if err != nil {
		return ErrInvalidRequest
	}
	var sessionIDs []pgtype.UUID
	var changedUserID pgtype.UUID
	err = s.store.WithTx(ctx, func(tx *database.Tx) error {
		q := tx.Queries
		userID, err := q.LockActionTokenOwner(ctx, database.LockActionTokenOwnerParams{ID: parsed.Selector, Purpose: ActionTokenPurposePasswordReset})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrInvalidRequest
			}
			return fmt.Errorf("lock password reset token owner: %w", err)
		}
		user, err := q.GetUserWithPasswordForUpdate(ctx, userID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrInvalidRequest
			}
			return fmt.Errorf("lock password credential: %w", err)
		}
		changedUserID = user.ID
		token, err := q.GetActionTokenForUpdate(ctx, database.GetActionTokenForUpdateParams{ID: parsed.Selector, Purpose: ActionTokenPurposePasswordReset})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrInvalidRequest
			}
			return fmt.Errorf("lock password reset token: %w", err)
		}
		if !s.actionTokens.VerifySecret(parsed.Secret, token.SecretHash) || token.UserID != user.ID || token.ConsumedAt.Valid {
			return ErrInvalidRequest
		}
		if !token.ExpiresAt.Valid || !s.clock.Now().Before(token.ExpiresAt.Time) {
			return ErrActionTokenExpired
		}
		if err := mapUserStatus(user.Status); err != nil {
			return err
		}
		if err := s.validateNewPassword(input.NewPassword, user.EmailNormalized); err != nil {
			return err
		}
		passwordHash, err := s.hasher.Hash(input.NewPassword)
		if err != nil {
			return fmt.Errorf("hash new password: %w", err)
		}
		version, err := q.UpdateUserPasswordAndIncrementVersion(ctx, database.UpdateUserPasswordAndIncrementVersionParams{PasswordHash: passwordHash, UserID: user.ID})
		if err != nil {
			return fmt.Errorf("update reset password: %w", err)
		}
		if _, err := q.ConsumeActionToken(ctx, database.ConsumeActionTokenParams{ID: token.ID, UserID: user.ID, Purpose: ActionTokenPurposePasswordReset}); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrInvalidRequest
			}
			return fmt.Errorf("consume password reset token: %w", err)
		}
		if _, err := q.InvalidatePreviousActionTokens(ctx, database.InvalidatePreviousActionTokensParams{UserID: user.ID, Purpose: ActionTokenPurposePasswordReset}); err != nil {
			return fmt.Errorf("invalidate remaining reset tokens: %w", err)
		}
		sessionIDs, err = q.ListAllUserSessionIDs(ctx, user.ID)
		if err != nil {
			return fmt.Errorf("list sessions for password reset: %w", err)
		}
		affected, err := s.sessions.RevokeAllUserSessionsInTx(ctx, q, user.ID, "password_reset")
		if err != nil {
			return err
		}
		metadata := audit.NewMetadata()
		metadata.Set("affected_session_count", len(affected))
		metadata.Set("previous_password_version", version.PreviousPasswordVersion)
		metadata.Set("new_password_version", version.NewPasswordVersion)
		if err := s.writeAuditForEntity(ctx, q, audit.EventAuthPasswordResetComp, database.AuditOutcomeSUCCESS, user.ID, pgtype.UUID{}, pgtype.UUID{}, pgtype.UUID{}, pgtype.UUID{}, pgtype.Text{String: "auth_action_token", Valid: true}, token.ID, meta, metadata); err != nil {
			return fmt.Errorf("write completed password reset audit: %w", err)
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, ErrInvalidRequest) || errors.Is(err, ErrActionTokenExpired) || errors.Is(err, ErrWeakPassword) || errors.Is(err, ErrCommonPassword) || errors.Is(err, ErrUserSuspended) || errors.Is(err, ErrUserDisabled) {
			return err
		}
		return fmt.Errorf("%w: reset password: %w", ErrDependencyUnavailable, err)
	}
	s.invalidator.InvalidateUserPasswordVersion(ctx, changedUserID)
	for _, sessionID := range sessionIDs {
		s.invalidator.InvalidateSession(ctx, sessionID)
	}
	return nil
}

func (s *Service) ChangePassword(ctx context.Context, userID, currentSessionID pgtype.UUID, input ChangePasswordRequest, meta requestmeta.RequestMetadata) error {
	var sessionIDs []pgtype.UUID
	err := s.store.WithTx(ctx, func(tx *database.Tx) error {
		q := tx.Queries
		user, err := q.GetUserWithPasswordForUpdate(ctx, userID)
		if err != nil {
			return fmt.Errorf("load current password: %w", err)
		}
		if err := mapUserStatus(user.Status); err != nil {
			return err
		}
		match, _, err := s.hasher.Verify(input.CurrentPassword, user.PasswordHash)
		if err != nil {
			return fmt.Errorf("verify current password: %w", err)
		}
		if !match {
			return ErrInvalidCredentials
		}
		if err := s.validateNewPassword(input.NewPassword, user.EmailNormalized); err != nil {
			return err
		}
		passwordHash, err := s.hasher.Hash(input.NewPassword)
		if err != nil {
			return fmt.Errorf("hash changed password: %w", err)
		}
		version, err := q.UpdateUserPasswordAndIncrementVersion(ctx, database.UpdateUserPasswordAndIncrementVersionParams{PasswordHash: passwordHash, UserID: user.ID})
		if err != nil {
			return fmt.Errorf("change password: %w", err)
		}
		if _, err := q.InvalidatePreviousActionTokens(ctx, database.InvalidatePreviousActionTokensParams{UserID: user.ID, Purpose: ActionTokenPurposePasswordReset}); err != nil {
			return fmt.Errorf("invalidate reset tokens after password change: %w", err)
		}
		sessionIDs, err = q.ListAllUserSessionIDs(ctx, user.ID)
		if err != nil {
			return fmt.Errorf("list sessions for password change: %w", err)
		}
		affected, err := s.sessions.RevokeAllUserSessionsInTx(ctx, q, user.ID, "password_changed")
		if err != nil {
			return err
		}
		metadata := audit.NewMetadata()
		metadata.Set("affected_session_count", len(affected))
		metadata.Set("previous_password_version", version.PreviousPasswordVersion)
		metadata.Set("new_password_version", version.NewPasswordVersion)
		if err := s.writeAudit(ctx, q, audit.EventAuthPasswordChanged, database.AuditOutcomeSUCCESS, user.ID, pgtype.UUID{}, pgtype.UUID{}, pgtype.UUID{}, currentSessionID, meta, metadata); err != nil {
			return fmt.Errorf("write password changed audit: %w", err)
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) || errors.Is(err, ErrWeakPassword) || errors.Is(err, ErrCommonPassword) || errors.Is(err, ErrUserSuspended) || errors.Is(err, ErrUserDisabled) {
			return err
		}
		return fmt.Errorf("%w: change password: %w", ErrDependencyUnavailable, err)
	}
	s.invalidator.InvalidateUserPasswordVersion(ctx, userID)
	for _, sessionID := range sessionIDs {
		s.invalidator.InvalidateSession(ctx, sessionID)
	}
	return nil
}

func (s *Service) validateNewPassword(value, normalizedEmail string) error {
	if value == "" {
		return ErrWeakPassword
	}
	if err := s.policy.Validate(value, normalizedEmail, s.blocklist); err != nil {
		if errors.Is(err, password.ErrPasswordCommon) {
			return ErrCommonPassword
		}
		return ErrWeakPassword
	}
	return nil
}

func validateActionEmail(value string) (string, error) {
	normalized := normalizeEmail(value)
	if err := validateEmail(normalized); err != nil {
		return "", err
	}
	return normalized, nil
}

func mapUserStatus(status database.UserStatus) error {
	switch status {
	case database.UserStatusACTIVE:
		return nil
	case database.UserStatusSUSPENDED:
		return ErrUserSuspended
	case database.UserStatusDISABLED:
		return ErrUserDisabled
	default:
		return ErrDependencyUnavailable
	}
}
