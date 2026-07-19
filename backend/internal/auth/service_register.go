package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/netip"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gabrielalc23/pdv/internal/audit"
	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/gabrielalc23/pdv/internal/platform/password"
	"github.com/gabrielalc23/pdv/internal/platform/requestmeta"
	"github.com/gabrielalc23/pdv/internal/sessions"
)

func (s *Service) Register(ctx context.Context, input RegisterRequest, meta requestmeta.RequestMetadata) (RegisterResult, error) {
	if !s.cfg.RegistrationEnabled {
		return RegisterResult{}, ErrRegistrationDisabled
	}
	if err := validateRegister(&input); err != nil {
		return RegisterResult{}, err
	}
	normalizedEmail := normalizeEmail(input.Email)
	if err := s.policy.Validate(input.Password, normalizedEmail, s.blocklist); err != nil {
		if errors.Is(err, password.ErrPasswordCommon) {
			return RegisterResult{}, ErrCommonPassword
		}
		return RegisterResult{}, ErrWeakPassword
	}
	passwordHash, err := s.hasher.Hash(input.Password)
	if err != nil {
		return RegisterResult{}, fmt.Errorf("%w: hash password: %w", ErrDependencyUnavailable, err)
	}
	var sessionResult sessions.CreateSessionResult
	var authResult AuthResult
	var verificationRaw string
	var email, displayName string
	err = s.store.WithTx(ctx, func(tx *database.Tx) error {
		q := tx.Queries
		user, err := q.CreateUserWithPassword(ctx, database.CreateUserWithPasswordParams{Email: input.Email, EmailNormalized: normalizedEmail, DisplayName: input.DisplayName, PasswordHash: passwordHash})
		if err != nil {
			return mapPersistenceError(err)
		}
		email, displayName = user.Email, user.DisplayName
		if !s.cfg.RequireVerifiedEmail {
			if _, err := q.VerifyUserEmail(ctx, user.ID); err != nil {
				return fmt.Errorf("verify registration email: %w", err)
			}
		}
		organization, err := q.CreateOrganization(ctx, database.CreateOrganizationParams{Name: input.Organization.Name, Slug: input.Organization.Slug, Timezone: input.Organization.Timezone, Locale: input.Organization.Locale, Currency: input.Organization.Currency, CreatedByUserID: user.ID})
		if err != nil {
			return mapPersistenceError(err)
		}
		store, err := q.CreateStore(ctx, database.CreateStoreParams{OrganizationID: organization.ID, Code: input.Store.Code, Name: input.Store.Name, Timezone: input.Store.Timezone, CreatedByUserID: user.ID})
		if err != nil {
			return mapPersistenceError(err)
		}
		membership, err := q.CreateMembership(ctx, database.CreateMembershipParams{OrganizationID: organization.ID, UserID: user.ID, DefaultStoreID: store.ID, CreatedByUserID: user.ID})
		if err != nil {
			return fmt.Errorf("create owner membership: %w", err)
		}
		ownerRole, err := bootstrapRoles(ctx, q, organization.ID, membership.ID)
		if err != nil {
			return err
		}
		if _, err := q.CreateRoleBinding(ctx, database.CreateRoleBindingParams{OrganizationID: organization.ID, MembershipID: membership.ID, RoleID: ownerRole.ID, CreatedByMembershipID: membership.ID}); err != nil {
			return fmt.Errorf("create owner binding: %w", err)
		}
		if err := bootstrapPaymentMethods(ctx, q, organization.ID, store.ID); err != nil {
			return err
		}
		if s.cfg.RequireVerifiedEmail {
			selector, err := randomUUID()
			if err != nil {
				return fmt.Errorf("generate verification selector: %w", err)
			}
			rawToken, secretHash, err := s.actionTokens.Generate(ActionTokenPurposeEmailVerification, selector)
			if err != nil {
				return fmt.Errorf("generate verification token: %w", err)
			}
			var requestedIP *netip.Addr
			if ip, parseErr := netip.ParseAddr(meta.ClientIP); parseErr == nil {
				requestedIP = &ip
			}
			token, err := q.CreateActionToken(ctx, database.CreateActionTokenParams{ID: selector, UserID: user.ID, Purpose: database.AuthActionTokenPurposeEMAILVERIFICATION, SecretHash: secretHash, ExpiresAt: pgtype.Timestamptz{Time: s.clock.Now().Add(24 * time.Hour), Valid: true}, RequestedIp: requestedIP})
			if err != nil {
				return fmt.Errorf("create verification token: %w", err)
			}
			if token.ID != selector {
				return fmt.Errorf("create verification token: persisted selector mismatch")
			}
			verificationRaw = rawToken
		} else {
			sessionResult, err = s.sessions.CreateSessionInTx(ctx, q, sessions.CreateSessionInput{UserID: user.ID, ClientID: input.ClientID, DeviceName: nullableText(input.DeviceName), UserAgent: nullableText(meta.UserAgent), IPAddress: parseIP(meta.ClientIP), ContextKind: sessions.ContextStore, OrganizationID: organization.ID, MembershipID: membership.ID, StoreID: store.ID})
			if err != nil {
				return fmt.Errorf("create registration session: %w", err)
			}
			selected, err := resolveContext(ctx, q, user.ID, organization.ID, store.ID)
			if err != nil {
				return fmt.Errorf("resolve registration context: %w", err)
			}
			state := State{UserID: user.ID, Email: user.Email, DisplayName: user.DisplayName, EmailVerified: true, PasswordVersion: user.PasswordVersion, SessionID: sessionResult.Session.ID, ClientID: sessionResult.Session.ClientID, DeviceName: sessionResult.Session.DeviceName.String, SessionStatus: database.AuthSessionStatus(sessionResult.Session.Status), CreatedAt: sessionResult.Session.CreatedAt, IdleExpiresAt: sessionResult.Session.IdleExpiresAt, AbsoluteExpires: sessionResult.Session.AbsoluteExpiresAt, Context: selected}
			response, err := s.issue(state, state.CreatedAt)
			if err != nil {
				return err
			}
			authResult = AuthResult{Response: response, RawRefreshToken: sessionResult.RawRefreshToken, RefreshExpires: state.AbsoluteExpires}
		}
		metadata := audit.NewMetadata()
		metadata.Set("client_id", input.ClientID)
		metadata.Set("verification_required", s.cfg.RequireVerifiedEmail)
		if err := s.writeAudit(ctx, q, audit.EventAuthRegistered, database.AuditOutcomeSUCCESS, user.ID, membership.ID, organization.ID, store.ID, sessionResult.Session.ID, meta, metadata); err != nil {
			return fmt.Errorf("write registration audit: %w", err)
		}
		return nil
	})
	if err != nil {
		return RegisterResult{}, err
	}
	if s.cfg.RequireVerifiedEmail {
		if s.mailer == nil {
			slog.Error("verification email not sent", "reason", "mailer unavailable")
		} else {
			mailCtx, cancel := mailDeliveryContext(ctx)
			defer cancel()
			logSecondaryFailure("verification email delivery failed", s.mailer.SendEmailVerification(mailCtx, email, displayName, s.mailLinks.BuildEmailVerification(verificationRaw)))
		}
		return RegisterResult{VerificationRequired: true}, nil
	}
	return RegisterResult{Auth: &authResult}, nil
}

func nullableText(value string) pgtype.Text { return pgtype.Text{String: value, Valid: value != ""} }
func parseIP(value string) *netip.Addr {
	ip, err := netip.ParseAddr(value)
	if err != nil {
		return nil
	}
	return &ip
}
