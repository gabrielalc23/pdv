package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gabrielalc23/pdv/internal/audit"
	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/gabrielalc23/pdv/internal/platform/requestmeta"
	"github.com/gabrielalc23/pdv/internal/sessions"
)

func (s *Service) ResolveRefreshSessionID(ctx context.Context, rawToken string) (pgtype.UUID, error) {
	return s.sessions.ResolveRefreshSessionID(ctx, rawToken)
}

func (s *Service) Refresh(ctx context.Context, rawToken string, meta requestmeta.RequestMetadata) (AuthResult, error) {
	rotated, err := s.sessions.RotateRefreshToken(ctx, sessions.RotateInput{RawRefreshToken: rawToken, RequestMeta: meta})
	if err != nil {
		return AuthResult{}, err
	}
	state, err := s.loadState(ctx, rotated.Session.ID)
	if err != nil && shouldDowngradeContext(err) {
		if downgradeErr := s.setContext(ctx, rotated.Session.UserID, rotated.Session.ID, Context{Kind: database.AuthContextKindIDENTITY}, meta); downgradeErr != nil {
			return AuthResult{}, downgradeErr
		}
		state, err = s.loadState(ctx, rotated.Session.ID)
	}
	if err != nil {
		return AuthResult{}, err
	}
	response, err := s.issue(state, state.CreatedAt)
	if err != nil {
		return AuthResult{}, err
	}
	return AuthResult{Response: response, RawRefreshToken: rotated.RawRefreshToken, RefreshExpires: state.AbsoluteExpires}, nil
}

func shouldDowngradeContext(err error) bool {
	return errors.Is(err, ErrInvalidAuthContext) || errors.Is(err, ErrOrganizationNotFound) || errors.Is(err, ErrOrganizationSuspended) || errors.Is(err, ErrMembershipNotFound) || errors.Is(err, ErrMembershipSuspended) || errors.Is(err, ErrStoreNotFound) || errors.Is(err, ErrStoreInactive)
}

func (s *Service) SwitchContext(ctx context.Context, userID, sessionID pgtype.UUID, input ContextRequest, authTime time.Time, meta requestmeta.RequestMetadata) (AuthResult, error) {
	orgID, err := parseOptionalUUID(input.OrganizationID, "organizationId")
	if err != nil {
		return AuthResult{}, err
	}
	storeID, err := parseOptionalUUID(input.StoreID, "storeId")
	if err != nil {
		return AuthResult{}, err
	}
	if storeID.Valid && !orgID.Valid {
		return AuthResult{}, ErrInvalidAuthContext
	}
	var target Context
	err = s.store.WithTx(ctx, func(tx *database.Tx) error {
		row, err := tx.Queries.GetAuthSessionForUpdate(ctx, sessionID)
		if err != nil {
			return mapPersistenceError(err)
		}
		if row.UserID != userID {
			return ErrSessionNotFound
		}
		if row.Status != database.AuthSessionStatusACTIVE {
			return sessions.ErrSessionRevoked
		}
		now := s.clock.Now()
		if !now.Before(row.IdleExpiresAt.Time) || !now.Before(row.AbsoluteExpiresAt.Time) {
			return sessions.ErrSessionExpired
		}
		target, err = resolveContext(ctx, tx.Queries, userID, orgID, storeID)
		if err != nil {
			return err
		}
		if _, err := tx.Queries.UpdateSessionContext(ctx, database.UpdateSessionContextParams{ContextKind: target.Kind, CurrentOrganizationID: target.OrganizationID, CurrentMembershipID: target.MembershipID, CurrentStoreID: target.StoreID, SessionID: sessionID, UserID: userID}); err != nil {
			return fmt.Errorf("update session context: %w", err)
		}
		metadata := audit.NewMetadata()
		metadata.Set("context_kind", contextKindString(target.Kind))
		if err := s.writeAudit(ctx, tx.Queries, audit.EventAuthContextChanged, database.AuditOutcomeSUCCESS, userID, target.MembershipID, target.OrganizationID, target.StoreID, sessionID, meta, metadata); err != nil {
			return fmt.Errorf("write context audit: %w", err)
		}
		return nil
	})
	if err != nil {
		return AuthResult{}, err
	}
	if s.invalidator != nil {
		s.invalidator.InvalidateSession(ctx, sessionID)
	}
	state, err := s.loadState(ctx, sessionID)
	if err != nil {
		return AuthResult{}, err
	}
	response, err := s.issue(state, authTime)
	if err != nil {
		return AuthResult{}, err
	}
	return AuthResult{Response: response, RefreshExpires: state.AbsoluteExpires}, nil
}

func (s *Service) setContext(ctx context.Context, userID, sessionID pgtype.UUID, target Context, meta requestmeta.RequestMetadata) error {
	err := s.store.WithTx(ctx, func(tx *database.Tx) error {
		row, err := tx.Queries.GetAuthSessionForUpdate(ctx, sessionID)
		if err != nil {
			return mapPersistenceError(err)
		}
		if row.UserID != userID || row.Status != database.AuthSessionStatusACTIVE {
			return sessions.ErrSessionRevoked
		}
		if _, err := tx.Queries.UpdateSessionContext(ctx, database.UpdateSessionContextParams{ContextKind: target.Kind, SessionID: sessionID, UserID: userID}); err != nil {
			return fmt.Errorf("downgrade session context: %w", err)
		}
		metadata := audit.NewMetadata()
		metadata.Set("context_kind", "identity")
		metadata.Set("reason", "context_no_longer_valid")
		return s.writeAudit(ctx, tx.Queries, audit.EventAuthContextChanged, database.AuditOutcomeSUCCESS, userID, pgtype.UUID{}, pgtype.UUID{}, pgtype.UUID{}, sessionID, meta, metadata)
	})
	if err == nil && s.invalidator != nil {
		s.invalidator.InvalidateSession(ctx, sessionID)
	}
	return err
}

func (s *Service) Logout(ctx context.Context, userID, sessionID pgtype.UUID, meta requestmeta.RequestMetadata) error {
	_, err := s.sessions.RevokeCurrentSession(ctx, userID, sessionID, meta)
	if errors.Is(err, sessions.ErrSessionNotFound) {
		return nil
	}
	return err
}

func (s *Service) LogoutAll(ctx context.Context, userID, sessionID pgtype.UUID, meta requestmeta.RequestMetadata) error {
	_, err := s.sessions.RevokeAllUserSessions(ctx, userID, sessionID, meta)
	return err
}
