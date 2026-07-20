package roles

import (
	"context"
	"errors"
	"fmt"
	"net/netip"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gabrielalc23/pdv/internal/audit"
	"github.com/gabrielalc23/pdv/internal/platform/authcontext"
	"github.com/gabrielalc23/pdv/internal/platform/clock"
	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/gabrielalc23/pdv/internal/platform/requestmeta"
)

type Service struct {
	store       Store
	txProvider  TxProvider
	invalidator SessionCacheInvalidator
	clock       clock.Clock
}

func NewService(store Store, txProvider TxProvider, invalidator SessionCacheInvalidator, clk clock.Clock) (*Service, error) {
	if store == nil {
		return nil, errors.New("roles: store is required")
	}
	if txProvider == nil {
		return nil, errors.New("roles: transaction provider is required")
	}
	if clk == nil {
		clk = clock.RealClock{}
	}
	return &Service{store: store, txProvider: txProvider, invalidator: invalidator, clock: clk}, nil
}

func requireActorScope(actor authcontext.Principal, scope authcontext.Scope) error {
	if !actor.HasOrganizationScope() || !actor.OrganizationID.Valid || !actor.MembershipID.Valid || !actor.UserID.Valid || !actor.SessionID.Valid {
		return ErrOrganizationContext
	}
	if !actor.Scopes.HasAll(scope) {
		return ErrInsufficientScope
	}
	return nil
}

func (s *Service) invalidateSessions(ctx context.Context, sessionIDs []pgtype.UUID) {
	if s.invalidator == nil {
		return
	}
	ctx = context.WithoutCancel(ctx)
	for _, sessionID := range sessionIDs {
		s.invalidator.InvalidateSession(ctx, sessionID)
	}
}

func (s *Service) invalidateOrganizationVersion(ctx context.Context, organizationID pgtype.UUID) {
	if s.invalidator != nil {
		s.invalidator.InvalidateOrganizationAuthorizationVersion(context.WithoutCancel(ctx), organizationID)
	}
}

func (s *Service) invalidateMembershipVersion(ctx context.Context, membershipID pgtype.UUID) {
	if s.invalidator != nil {
		s.invalidator.InvalidateMembershipAuthorizationVersion(context.WithoutCancel(ctx), membershipID)
	}
}

func writeAudit(ctx context.Context, tx TxStore, actor authcontext.Principal, eventType, entityType string, entityID pgtype.UUID, metadata audit.Metadata) error {
	data, err := metadata.Marshal()
	if err != nil {
		return err
	}
	meta := requestmeta.MustFromContext(ctx)
	event := audit.Event{
		OrganizationID:    actor.OrganizationID,
		StoreID:           actor.StoreID,
		ActorUserID:       actor.UserID,
		ActorMembershipID: actor.MembershipID,
		SessionID:         actor.SessionID,
		EventType:         eventType,
		EntityType:        pgtype.Text{String: entityType, Valid: true},
		EntityID:          entityID,
		Outcome:           database.AuditOutcomeSUCCESS,
		Metadata:          data,
	}
	if meta.RequestID != "" {
		event.RequestID = pgtype.Text{String: meta.RequestID, Valid: true}
	}
	if meta.UserAgent != "" {
		event.UserAgent = pgtype.Text{String: meta.UserAgent, Valid: true}
	}
	if ip, parseErr := netip.ParseAddr(meta.ClientIP); parseErr == nil {
		event.IPAddress = &ip
	}
	if err := tx.WriteAudit(ctx, event); err != nil {
		return fmt.Errorf("%w: write audit event: %w", ErrDependencyUnavailable, err)
	}
	return nil
}

func ensureGrantSubset(ctx context.Context, tx TxStore, actor authcontext.Principal, requested []string) error {
	granted, err := tx.ResolveActorGrantScopes(ctx, actor.OrganizationID, actor.MembershipID)
	if err != nil {
		return fmt.Errorf("%w: resolve actor grant scopes: %w", ErrDependencyUnavailable, err)
	}
	available := make(map[string]struct{}, len(granted))
	for _, code := range granted {
		available[code] = struct{}{}
	}
	for _, code := range requested {
		if _, ok := available[code]; !ok {
			return fmt.Errorf("%w: actor cannot grant %s", ErrAuthorizationEscalation, code)
		}
	}
	return nil
}

func validateRoleScopes(ctx context.Context, tx TxStore, actor authcontext.Principal, assignmentScope database.RoleAssignmentScope, scopeCodes []string) error {
	rows, err := tx.GetPermissionScopesByCodes(ctx, scopeCodes)
	if err != nil {
		return fmt.Errorf("%w: get permission scopes: %w", ErrDependencyUnavailable, err)
	}
	if len(rows) != len(scopeCodes) {
		return validationError("scopes", "contains an unknown scope code")
	}
	for _, row := range rows {
		if !row.IsAssignable {
			return fmt.Errorf("%w: %s", ErrScopeNotAssignable, row.Code)
		}
		if assignmentScope == database.RoleAssignmentScopeSTORE && row.ScopeLevel != database.PermissionScopeLevelSTORE {
			return fmt.Errorf("%w: store role cannot contain %s", ErrScopeLevelInvalid, row.Code)
		}
	}
	return ensureGrantSubset(ctx, tx, actor, scopeCodes)
}
