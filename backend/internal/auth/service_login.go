package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/gabrielalc23/pdv/internal/audit"
	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/gabrielalc23/pdv/internal/platform/ratelimit"
	"github.com/gabrielalc23/pdv/internal/platform/requestmeta"
	"github.com/gabrielalc23/pdv/internal/sessions"
)

func (s *Service) Login(ctx context.Context, input LoginRequest, meta requestmeta.RequestMetadata) (AuthResult, error) {
	input.Email = strings.TrimSpace(input.Email)
	if err := validateEmail(input.Email); err != nil {
		return AuthResult{}, ErrInvalidCredentials
	}
	if err := validateClientID(input.ClientID); err != nil {
		return AuthResult{}, err
	}
	input.DeviceName = strings.TrimSpace(input.DeviceName)
	if len(input.DeviceName) > 150 {
		return AuthResult{}, validationError("deviceName", "Nome do dispositivo inválido.")
	}
	normalized := normalizeEmail(input.Email)
	fingerprint := ratelimit.Fingerprint(s.cfg.RateLimitKey, normalized)
	user, err := s.store.Queries.GetUserByNormalizedEmail(ctx, normalized)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			_, _, _ = s.hasher.Verify(input.Password, s.dummyHash)
			if auditErr := s.auditLoginFailure(ctx, meta, input.ClientID, fingerprint); auditErr != nil {
				return AuthResult{}, auditErr
			}
			return AuthResult{}, ErrInvalidCredentials
		}
		return AuthResult{}, mapPersistenceError(err)
	}
	match, _, verifyErr := s.hasher.Verify(input.Password, user.PasswordHash)
	if verifyErr != nil || !match {
		if auditErr := s.auditLoginFailure(ctx, meta, input.ClientID, fingerprint); auditErr != nil {
			return AuthResult{}, auditErr
		}
		return AuthResult{}, ErrInvalidCredentials
	}
	switch user.Status {
	case database.UserStatusSUSPENDED:
		return AuthResult{}, ErrUserSuspended
	case database.UserStatusDISABLED:
		return AuthResult{}, ErrUserDisabled
	}
	if s.cfg.RequireVerifiedEmail && !user.EmailVerifiedAt.Valid {
		return AuthResult{}, ErrEmailNotVerified
	}
	var created sessions.CreateSessionResult
	var authResult AuthResult
	err = s.store.WithTx(ctx, func(tx *database.Tx) error {
		lockedUser, err := tx.Queries.GetUserWithPasswordForUpdate(ctx, user.ID)
		if err != nil {
			return fmt.Errorf("lock login credential: %w", err)
		}
		if lockedUser.PasswordHash != user.PasswordHash || lockedUser.PasswordVersion != user.PasswordVersion {
			return ErrInvalidCredentials
		}
		if err := mapUserStatus(lockedUser.Status); err != nil {
			return err
		}
		if s.cfg.RequireVerifiedEmail && !lockedUser.EmailVerifiedAt.Valid {
			return ErrEmailNotVerified
		}
		selected, err := resolveLoginContext(ctx, tx.Queries, lockedUser.ID, input.ClientID, input.OrganizationID, input.StoreID)
		if err != nil {
			return err
		}
		created, err = s.sessions.CreateSessionInTx(ctx, tx.Queries, sessions.CreateSessionInput{UserID: lockedUser.ID, ClientID: input.ClientID, DeviceName: nullableText(input.DeviceName), UserAgent: nullableText(meta.UserAgent), IPAddress: parseIP(meta.ClientIP), ContextKind: sessions.ContextKind(selected.Kind), OrganizationID: selected.OrganizationID, MembershipID: selected.MembershipID, StoreID: selected.StoreID})
		if err != nil {
			return fmt.Errorf("create login session: %w", err)
		}
		if _, err := tx.Queries.UpdateUserLastLogin(ctx, lockedUser.ID); err != nil {
			return fmt.Errorf("update last login: %w", err)
		}
		metadata := audit.NewMetadata()
		metadata.Set("client_id", input.ClientID)
		if err := s.writeAudit(ctx, tx.Queries, audit.EventAuthLoginSucceeded, database.AuditOutcomeSUCCESS, lockedUser.ID, selected.MembershipID, selected.OrganizationID, selected.StoreID, created.Session.ID, meta, metadata); err != nil {
			return fmt.Errorf("write login audit: %w", err)
		}
		state := State{UserID: lockedUser.ID, Email: lockedUser.Email, DisplayName: lockedUser.DisplayName, EmailVerified: lockedUser.EmailVerifiedAt.Valid, PasswordVersion: lockedUser.PasswordVersion, SessionID: created.Session.ID, ClientID: created.Session.ClientID, DeviceName: created.Session.DeviceName.String, SessionStatus: database.AuthSessionStatus(created.Session.Status), CreatedAt: created.Session.CreatedAt, IdleExpiresAt: created.Session.IdleExpiresAt, AbsoluteExpires: created.Session.AbsoluteExpiresAt, Context: selected}
		response, err := s.issue(state, state.CreatedAt)
		if err != nil {
			return err
		}
		authResult = AuthResult{Response: response, RawRefreshToken: created.RawRefreshToken, RefreshExpires: state.AbsoluteExpires}
		return nil
	})
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			if auditErr := s.auditLoginFailure(ctx, meta, input.ClientID, fingerprint); auditErr != nil {
				return AuthResult{}, auditErr
			}
		}
		return AuthResult{}, err
	}
	return authResult, nil
}

func (s *Service) auditLoginFailure(ctx context.Context, meta requestmeta.RequestMetadata, clientID, fingerprint string) error {
	metadata := audit.NewMetadata()
	metadata.Set("client_id", clientID)
	metadata.Set("email_fingerprint", fingerprint)
	metadata.Set("reason_category", "credentials_rejected")
	if err := s.writeAudit(ctx, s.store.Queries, audit.EventAuthLoginFailed, database.AuditOutcomeFAILURE, database.User{}.ID, database.OrganizationMembership{}.ID, database.Organization{}.ID, database.Store{}.ID, database.AuthSession{}.ID, meta, metadata); err != nil {
		return fmt.Errorf("%w: write failed login audit: %w", ErrDependencyUnavailable, err)
	}
	return nil
}
