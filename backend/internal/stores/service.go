package stores

import (
	"context"
	"fmt"
	"net/netip"

	"github.com/gabrielalc23/pdv/internal/audit"
	"github.com/gabrielalc23/pdv/internal/platform/authcontext"
	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/gabrielalc23/pdv/internal/platform/requestmeta"
	"github.com/jackc/pgx/v5/pgtype"
)

type Repository interface {
	GetStoreForOrganization(context.Context, database.GetStoreForOrganizationParams) (database.Store, error)
	ListStoresForOrganization(context.Context, database.ListStoresForOrganizationParams) ([]database.Store, error)
	CountStoresForOrganization(context.Context, database.CountStoresForOrganizationParams) (int64, error)
	ListPaymentMethodsForOrganization(context.Context, pgtype.UUID) ([]database.PaymentMethod, error)
	ListStorePaymentMethods(context.Context, database.ListStorePaymentMethodsParams) ([]database.ListStorePaymentMethodsRow, error)
}

type TxManager interface {
	WithTx(context.Context, func(database.Querier) error) error
}

type postgresTxManager struct{ store *database.PostgresStore }

func (m postgresTxManager) WithTx(ctx context.Context, fn func(database.Querier) error) error {
	return m.store.WithTx(ctx, func(tx *database.Tx) error { return fn(tx.Queries) })
}

type Service struct {
	repository  Repository
	txManager   TxManager
	audit       audit.Writer
	invalidator CacheInvalidator
}

type CacheInvalidator interface {
	InvalidateSession(context.Context, pgtype.UUID)
}

func NewService(store *database.PostgresStore, writer audit.Writer, invalidators ...CacheInvalidator) (*Service, error) {
	if store == nil || store.Queries == nil {
		return nil, ErrInvalidServiceDependencies
	}
	return NewServiceWithDependencies(store.Queries, postgresTxManager{store: store}, writer, invalidators...)
}

func NewServiceWithDependencies(repository Repository, txManager TxManager, writer audit.Writer, invalidators ...CacheInvalidator) (*Service, error) {
	if repository == nil || txManager == nil || writer == nil {
		return nil, ErrInvalidServiceDependencies
	}
	var invalidator CacheInvalidator
	if len(invalidators) > 0 {
		invalidator = invalidators[0]
	}
	return &Service{repository: repository, txManager: txManager, audit: writer, invalidator: invalidator}, nil
}

func organizationID(principal authcontext.Principal) (pgtype.UUID, error) {
	if !principal.HasOrganizationScope() || !principal.OrganizationID.Valid || !principal.UserID.Valid || !principal.MembershipID.Valid {
		return pgtype.UUID{}, ErrOrganizationContextRequired
	}
	return principal.OrganizationID, nil
}

func (s *Service) writeAudit(ctx context.Context, q database.Querier, principal authcontext.Principal, eventType, entityType string, entityID, storeID pgtype.UUID, metadata audit.Metadata) error {
	data, err := metadata.Marshal()
	if err != nil {
		return fmt.Errorf("marshal audit metadata: %w", err)
	}
	event := audit.Event{
		OrganizationID: principal.OrganizationID, StoreID: storeID, ActorUserID: principal.UserID,
		ActorMembershipID: principal.MembershipID, SessionID: principal.SessionID,
		EventType: eventType, EntityType: pgtype.Text{String: entityType, Valid: true}, EntityID: entityID,
		Outcome: database.AuditOutcomeSUCCESS, Metadata: data,
	}
	meta := requestmeta.MustFromContext(ctx)
	if meta.RequestID != "" {
		event.RequestID = pgtype.Text{String: meta.RequestID, Valid: true}
	}
	if meta.UserAgent != "" {
		event.UserAgent = pgtype.Text{String: meta.UserAgent, Valid: true}
	}
	if ip, parseErr := netip.ParseAddr(meta.ClientIP); parseErr == nil {
		event.IPAddress = &ip
	}
	if err := s.audit.Write(ctx, q, event); err != nil {
		return fmt.Errorf("write audit event: %w", err)
	}
	return nil
}

func lockOrganization(ctx context.Context, q database.Querier, organizationID pgtype.UUID) error {
	if _, err := q.LockOrganizationForOwnerChange(ctx, organizationID); err != nil {
		return fmt.Errorf("lock organization: %w", err)
	}
	return nil
}
