package memberships

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gabrielalc23/pdv/internal/audit"
	"github.com/gabrielalc23/pdv/internal/platform/database"
)

type Store interface {
	ListMemberships(context.Context, database.ListMembershipsParams) ([]database.ListMembershipsRow, error)
	CountMemberships(context.Context, database.CountMembershipsParams) (int64, error)
	GetMembershipForOrganization(context.Context, database.GetMembershipForOrganizationParams) (database.GetMembershipForOrganizationRow, error)
	ListMemberRoleBindings(context.Context, database.ListMemberRoleBindingsParams) ([]database.ListMemberRoleBindingsRow, error)
}

type TxStore interface {
	GetMembershipForUpdate(context.Context, database.GetMembershipForUpdateParams) (database.OrganizationMembership, error)
	ListMemberRoleBindings(context.Context, database.ListMemberRoleBindingsParams) ([]database.ListMemberRoleBindingsRow, error)
	ListStoresForMembership(context.Context, database.ListStoresForMembershipParams) ([]database.Store, error)
	ListSessionIDsForMembership(context.Context, database.ListSessionIDsForMembershipParams) ([]pgtype.UUID, error)
	RevokeSession(context.Context, database.RevokeSessionParams) (database.RevokeSessionRow, error)
	RevokeSessionRefreshTokens(context.Context, pgtype.UUID) ([]database.RevokeSessionRefreshTokensRow, error)
	LockActiveOwnerMembershipsForUpdate(context.Context, pgtype.UUID) (int64, error)
	UpdateMembershipDefaultStore(context.Context, pgtype.UUID, pgtype.UUID, pgtype.UUID) (database.OrganizationMembership, error)
	UpdateMembershipStatus(context.Context, database.UpdateMembershipStatusParams) (database.OrganizationMembership, error)
	WriteAudit(context.Context, audit.Event) error
}

type TxManager interface {
	WithTx(context.Context, func(TxStore) error) error
}

type store struct{ q *database.Queries }

func NewStore(q *database.Queries) Store { return &store{q: q} }

func (s *store) ListMemberships(ctx context.Context, params database.ListMembershipsParams) ([]database.ListMembershipsRow, error) {
	return s.q.ListMemberships(ctx, params)
}

func (s *store) CountMemberships(ctx context.Context, params database.CountMembershipsParams) (int64, error) {
	return s.q.CountMemberships(ctx, params)
}

func (s *store) GetMembershipForOrganization(ctx context.Context, params database.GetMembershipForOrganizationParams) (database.GetMembershipForOrganizationRow, error) {
	return s.q.GetMembershipForOrganization(ctx, params)
}

func (s *store) ListMemberRoleBindings(ctx context.Context, params database.ListMemberRoleBindingsParams) ([]database.ListMemberRoleBindingsRow, error) {
	return s.q.ListMemberRoleBindings(ctx, params)
}

type postgresTxManager struct {
	store *database.PostgresStore
	audit audit.Writer
}

func NewTxManager(store *database.PostgresStore, writer audit.Writer) TxManager {
	return &postgresTxManager{store: store, audit: writer}
}

func (m *postgresTxManager) WithTx(ctx context.Context, fn func(TxStore) error) (err error) {
	if m == nil || m.store == nil || m.store.Pool == nil || m.audit == nil {
		return fmt.Errorf("%w: transaction manager is not configured", ErrDependencyUnavailable)
	}
	tx, err := m.store.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
				err = errors.Join(err, rollbackErr)
			}
		}
	}()

	q := database.New(tx)
	if err = fn(&postgresTxStore{tx: tx, q: q, audit: m.audit}); err != nil {
		return err
	}
	err = tx.Commit(ctx)
	return err
}

type postgresTxStore struct {
	tx    pgx.Tx
	q     *database.Queries
	audit audit.Writer
}

func (s *postgresTxStore) GetMembershipForUpdate(ctx context.Context, params database.GetMembershipForUpdateParams) (database.OrganizationMembership, error) {
	return s.q.GetMembershipForUpdate(ctx, params)
}

func (s *postgresTxStore) ListMemberRoleBindings(ctx context.Context, params database.ListMemberRoleBindingsParams) ([]database.ListMemberRoleBindingsRow, error) {
	return s.q.ListMemberRoleBindings(ctx, params)
}

func (s *postgresTxStore) ListStoresForMembership(ctx context.Context, params database.ListStoresForMembershipParams) ([]database.Store, error) {
	return s.q.ListStoresForMembership(ctx, params)
}

func (s *postgresTxStore) ListSessionIDsForMembership(ctx context.Context, params database.ListSessionIDsForMembershipParams) ([]pgtype.UUID, error) {
	return s.q.ListSessionIDsForMembership(ctx, params)
}

func (s *postgresTxStore) RevokeSession(ctx context.Context, params database.RevokeSessionParams) (database.RevokeSessionRow, error) {
	return s.q.RevokeSession(ctx, params)
}

func (s *postgresTxStore) RevokeSessionRefreshTokens(ctx context.Context, sessionID pgtype.UUID) ([]database.RevokeSessionRefreshTokensRow, error) {
	return s.q.RevokeSessionRefreshTokens(ctx, sessionID)
}

// LockActiveOwnerMembershipsForUpdate serializes all owner-changing operations
// on the organization row before locking and counting active owner memberships.
func (s *postgresTxStore) LockActiveOwnerMembershipsForUpdate(ctx context.Context, organizationID pgtype.UUID) (int64, error) {
	if _, err := s.q.LockOrganizationForOwnerChange(ctx, organizationID); err != nil {
		return 0, err
	}
	return s.q.CountActiveOwnersForUpdate(ctx, organizationID)
}

func (s *postgresTxStore) UpdateMembershipDefaultStore(ctx context.Context, organizationID, membershipID, defaultStoreID pgtype.UUID) (database.OrganizationMembership, error) {
	const query = `UPDATE organization_memberships
SET default_store_id = $1, authorization_version = authorization_version + 1
WHERE organization_id = $2 AND id = $3
RETURNING id, organization_id, user_id, status, default_store_id,
          authorization_version, joined_at, suspended_at, removed_at,
          created_by_user_id, created_at, updated_at`
	var row database.OrganizationMembership
	err := s.tx.QueryRow(ctx, query, defaultStoreID, organizationID, membershipID).Scan(
		&row.ID, &row.OrganizationID, &row.UserID, &row.Status, &row.DefaultStoreID,
		&row.AuthorizationVersion, &row.JoinedAt, &row.SuspendedAt, &row.RemovedAt,
		&row.CreatedByUserID, &row.CreatedAt, &row.UpdatedAt,
	)
	return row, err
}

func (s *postgresTxStore) UpdateMembershipStatus(ctx context.Context, params database.UpdateMembershipStatusParams) (database.OrganizationMembership, error) {
	return s.q.UpdateMembershipStatus(ctx, params)
}

func (s *postgresTxStore) WriteAudit(ctx context.Context, event audit.Event) error {
	return s.audit.Write(ctx, s.q, event)
}
