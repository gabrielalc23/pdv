package roles

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gabrielalc23/pdv/internal/audit"
	"github.com/gabrielalc23/pdv/internal/platform/database"
)

type storeImpl struct {
	q *database.Queries
}

func NewStore(q *database.Queries) Store {
	return &storeImpl{q: q}
}

func (s *storeImpl) ListPermissionScopes(ctx context.Context) ([]database.PermissionScope, error) {
	return s.q.ListPermissionScopes(ctx)
}

func (s *storeImpl) ListRolesWithScopes(ctx context.Context, organizationID pgtype.UUID) ([]database.ListRolesWithScopesRow, error) {
	return s.q.ListRolesWithScopes(ctx, organizationID)
}

func (s *storeImpl) GetRoleWithScopes(ctx context.Context, organizationID, roleID pgtype.UUID) (database.GetRoleWithScopesForOrganizationRow, error) {
	return s.q.GetRoleWithScopesForOrganization(ctx, database.GetRoleWithScopesForOrganizationParams{
		OrganizationID: organizationID,
		RoleID:         roleID,
	})
}

type txProviderImpl struct {
	store  *database.PostgresStore
	writer audit.Writer
}

func NewTxProvider(store *database.PostgresStore, writer audit.Writer) TxProvider {
	return &txProviderImpl{store: store, writer: writer}
}

func (p *txProviderImpl) WithTx(ctx context.Context, fn func(TxStore) error) error {
	return p.store.WithTx(ctx, func(tx *database.Tx) error {
		return fn(&txStore{storeImpl: storeImpl{q: tx.Queries}, writer: p.writer})
	})
}

type txStore struct {
	storeImpl
	writer audit.Writer
}

func (s *txStore) GetPermissionScopesByCodes(ctx context.Context, scopeCodes []string) ([]database.PermissionScope, error) {
	return s.q.GetPermissionScopesByCodes(ctx, scopeCodes)
}

func (s *txStore) ResolveActorGrantScopes(ctx context.Context, organizationID, membershipID pgtype.UUID) ([]string, error) {
	return s.q.ResolveActorGrantScopes(ctx, database.ResolveActorGrantScopesParams{OrganizationID: organizationID, MembershipID: membershipID})
}

func (s *txStore) CreateRole(ctx context.Context, params database.CreateRoleParams) (database.Role, error) {
	return s.q.CreateRole(ctx, params)
}

func (s *txStore) LockRoleForScopeChange(ctx context.Context, organizationID, roleID pgtype.UUID) error {
	_, err := s.q.LockRoleForScopeChange(ctx, database.LockRoleForScopeChangeParams{OrganizationID: organizationID, RoleID: roleID})
	return err
}

func (s *txStore) ReplaceRoleScopes(ctx context.Context, organizationID, roleID pgtype.UUID, scopeCodes []string) error {
	_, err := s.q.ReplaceRoleScopes(ctx, database.ReplaceRoleScopesParams{OrganizationID: organizationID, RoleID: roleID, ScopeCodes: scopeCodes})
	return err
}

func (s *txStore) UpdateRole(ctx context.Context, params database.UpdateRoleParams) (database.Role, error) {
	return s.q.UpdateRole(ctx, params)
}

func (s *txStore) UpdateRoleStatus(ctx context.Context, params database.UpdateRoleStatusParams) (database.Role, error) {
	return s.q.UpdateRoleStatus(ctx, params)
}

func (s *txStore) IncrementOrganizationAuthorizationVersion(ctx context.Context, organizationID pgtype.UUID) error {
	_, err := s.q.IncrementOrganizationAuthorizationVersion(ctx, organizationID)
	return err
}

func (s *txStore) GetMembershipForUpdate(ctx context.Context, organizationID, membershipID pgtype.UUID) (database.OrganizationMembership, error) {
	return s.q.GetMembershipForUpdate(ctx, database.GetMembershipForUpdateParams{OrganizationID: organizationID, MembershipID: membershipID})
}

func (s *txStore) GetStoreForOrganization(ctx context.Context, organizationID, storeID pgtype.UUID) (database.Store, error) {
	return s.q.GetStoreForOrganization(ctx, database.GetStoreForOrganizationParams{OrganizationID: organizationID, StoreID: storeID})
}

func (s *txStore) ListMemberRoleBindings(ctx context.Context, organizationID, membershipID pgtype.UUID) ([]database.ListMemberRoleBindingsRow, error) {
	return s.q.ListMemberRoleBindings(ctx, database.ListMemberRoleBindingsParams{OrganizationID: organizationID, MembershipID: membershipID})
}

func (s *txStore) CreateRoleBinding(ctx context.Context, params database.CreateRoleBindingParams) (database.CreateRoleBindingRow, error) {
	return s.q.CreateRoleBinding(ctx, params)
}

func (s *txStore) GetRoleBindingForUpdate(ctx context.Context, organizationID, membershipID, bindingID pgtype.UUID) (database.GetRoleBindingForOrganizationRow, error) {
	return s.q.GetRoleBindingForOrganization(ctx, database.GetRoleBindingForOrganizationParams{OrganizationID: organizationID, MembershipID: membershipID, BindingID: bindingID})
}

func (s *txStore) DeleteRoleBinding(ctx context.Context, organizationID, membershipID, bindingID pgtype.UUID) (database.MembershipRoleBinding, error) {
	return s.q.DeleteRoleBinding(ctx, database.DeleteRoleBindingParams{OrganizationID: organizationID, MembershipID: membershipID, BindingID: bindingID})
}

func (s *txStore) LockOrganizationForOwnerChange(ctx context.Context, organizationID pgtype.UUID) error {
	_, err := s.q.LockOrganizationForOwnerChange(ctx, organizationID)
	return err
}

func (s *txStore) CountActiveOwnersForUpdate(ctx context.Context, organizationID pgtype.UUID) (int64, error) {
	return s.q.CountActiveOwnersForUpdate(ctx, organizationID)
}

func (s *txStore) IncrementMembershipAuthorizationVersion(ctx context.Context, organizationID, membershipID pgtype.UUID) error {
	_, err := s.q.IncrementMembershipAuthorizationVersion(ctx, database.IncrementMembershipAuthorizationVersionParams{OrganizationID: organizationID, MembershipID: membershipID})
	return err
}

func (s *txStore) ListSessionIDsForOrganization(ctx context.Context, organizationID pgtype.UUID) ([]pgtype.UUID, error) {
	return s.q.ListSessionIDsForOrganization(ctx, organizationID)
}

func (s *txStore) ListSessionIDsForMembership(ctx context.Context, organizationID, membershipID pgtype.UUID) ([]pgtype.UUID, error) {
	return s.q.ListSessionIDsForMembership(ctx, database.ListSessionIDsForMembershipParams{OrganizationID: organizationID, MembershipID: membershipID})
}

func (s *txStore) WriteAudit(ctx context.Context, event audit.Event) error {
	if s.writer == nil {
		return errors.New("roles: audit writer is required")
	}
	return s.writer.Write(ctx, s.q, event)
}
