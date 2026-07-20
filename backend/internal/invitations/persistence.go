package invitations

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gabrielalc23/pdv/internal/audit"
	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/gabrielalc23/pdv/internal/sessions"
)

type storeImpl struct{ q *database.Queries }

func NewStore(q *database.Queries) Store { return &storeImpl{q: q} }

func (s *storeImpl) ListInvitations(ctx context.Context, params database.ListInvitationsParams) ([]database.ListInvitationsRow, error) {
	return s.q.ListInvitations(ctx, params)
}
func (s *storeImpl) CountInvitations(ctx context.Context, params database.CountInvitationsParams) (int64, error) {
	return s.q.CountInvitations(ctx, params)
}
func (s *storeImpl) GetInvitation(ctx context.Context, id pgtype.UUID) (database.OrganizationInvitation, error) {
	return s.q.GetInvitationForUpdate(ctx, id)
}
func (s *storeImpl) ListInvitationRoleBindings(ctx context.Context, organizationID, invitationID pgtype.UUID) ([]database.ListInvitationRoleBindingsRow, error) {
	return s.q.ListInvitationRoleBindings(ctx, database.ListInvitationRoleBindingsParams{OrganizationID: organizationID, InvitationID: invitationID})
}
func (s *storeImpl) GetOrganization(ctx context.Context, id pgtype.UUID) (database.Organization, error) {
	return s.q.GetOrganizationForActor(ctx, id)
}
func (s *storeImpl) GetRole(ctx context.Context, organizationID, roleID pgtype.UUID) (database.Role, error) {
	return s.q.GetRoleForOrganization(ctx, database.GetRoleForOrganizationParams{OrganizationID: organizationID, RoleID: roleID})
}
func (s *storeImpl) ListScopeCodesForRole(ctx context.Context, organizationID, roleID pgtype.UUID) ([]string, error) {
	return s.q.ListScopeCodesForRole(ctx, database.ListScopeCodesForRoleParams{OrganizationID: organizationID, RoleID: roleID})
}
func (s *storeImpl) GetUserByNormalizedEmail(ctx context.Context, email string) (database.User, error) {
	return s.q.GetUserIdentityByNormalizedEmail(ctx, email)
}
func (s *storeImpl) GetUserByID(ctx context.Context, id pgtype.UUID) (database.User, error) {
	return s.q.GetUserByID(ctx, id)
}

type txProvider struct {
	store    *database.PostgresStore
	writer   audit.Writer
	sessions *sessions.Service
}

func NewTxProvider(store *database.PostgresStore, writer audit.Writer, sessionService *sessions.Service) TxProvider {
	return &txProvider{store: store, writer: writer, sessions: sessionService}
}

func (p *txProvider) WithTx(ctx context.Context, fn func(TxStore) error) error {
	if p == nil || p.store == nil || p.writer == nil || p.sessions == nil {
		return fmt.Errorf("%w: transaction provider is not configured", ErrDependencyUnavailable)
	}
	return p.store.WithTx(ctx, func(tx *database.Tx) error {
		return fn(&txStore{storeImpl: storeImpl{q: tx.Queries}, writer: p.writer, sessions: p.sessions})
	})
}

type txStore struct {
	storeImpl
	writer   audit.Writer
	sessions *sessions.Service
}

func (s *txStore) CreateInvitation(ctx context.Context, params database.CreateInvitationParams) (database.OrganizationInvitation, error) {
	return s.q.CreateInvitation(ctx, params)
}
func (s *txStore) ExpirePendingInvitationsForEmail(ctx context.Context, organizationID pgtype.UUID, email string) error {
	_, err := s.q.ExpirePendingInvitationsForEmail(ctx, database.ExpirePendingInvitationsForEmailParams{OrganizationID: organizationID, EmailNormalized: email})
	return err
}
func (s *txStore) CreateInvitationRoleBindings(ctx context.Context, params database.CreateInvitationRoleBindingsParams) ([]database.InvitationRoleBinding, error) {
	return s.q.CreateInvitationRoleBindings(ctx, params)
}
func (s *txStore) RotateInvitationSecret(ctx context.Context, params database.RotateInvitationSecretParams) (database.OrganizationInvitation, error) {
	return s.q.RotateInvitationSecret(ctx, params)
}
func (s *txStore) RevokeInvitation(ctx context.Context, params database.RevokeInvitationParams) (database.RevokeInvitationRow, error) {
	return s.q.RevokeInvitation(ctx, params)
}
func (s *txStore) GetMembership(ctx context.Context, organizationID, membershipID pgtype.UUID) (database.GetMembershipForOrganizationRow, error) {
	return s.q.GetMembershipForOrganization(ctx, database.GetMembershipForOrganizationParams{OrganizationID: organizationID, MembershipID: membershipID})
}
func (s *txStore) GetStore(ctx context.Context, organizationID, storeID pgtype.UUID) (database.Store, error) {
	return s.q.GetStoreForOrganization(ctx, database.GetStoreForOrganizationParams{OrganizationID: organizationID, StoreID: storeID})
}
func (s *txStore) CreateUser(ctx context.Context, params database.CreateUserWithPasswordParams) (database.CreateUserWithPasswordRow, error) {
	return s.q.CreateUserWithPassword(ctx, params)
}
func (s *txStore) VerifyUserEmail(ctx context.Context, userID pgtype.UUID) error {
	_, err := s.q.VerifyUserEmail(ctx, userID)
	return err
}
func (s *txStore) GetLatestMembership(ctx context.Context, organizationID, userID pgtype.UUID) (database.OrganizationMembership, error) {
	return s.q.GetLatestMembershipForUserInOrganization(ctx, database.GetLatestMembershipForUserInOrganizationParams{OrganizationID: organizationID, UserID: userID})
}
func (s *txStore) LockMembership(ctx context.Context, organizationID, membershipID pgtype.UUID) (database.OrganizationMembership, error) {
	return s.q.GetMembershipForUpdate(ctx, database.GetMembershipForUpdateParams{OrganizationID: organizationID, MembershipID: membershipID})
}
func (s *txStore) CreateMembership(ctx context.Context, params database.CreateMembershipParams) (database.OrganizationMembership, error) {
	return s.q.CreateMembership(ctx, params)
}
func (s *txStore) UpdateMembershipStatus(ctx context.Context, params database.UpdateMembershipStatusParams) (database.OrganizationMembership, error) {
	return s.q.UpdateMembershipStatus(ctx, params)
}
func (s *txStore) CopyBindings(ctx context.Context, params database.CreateMembershipBindingsFromInvitationParams) ([]database.MembershipRoleBinding, error) {
	return s.q.CreateMembershipBindingsFromInvitation(ctx, params)
}
func (s *txStore) DeleteMembershipBindings(ctx context.Context, organizationID, membershipID pgtype.UUID) error {
	_, err := s.q.DeleteMembershipRoleBindings(ctx, database.DeleteMembershipRoleBindingsParams{OrganizationID: organizationID, MembershipID: membershipID})
	return err
}
func (s *txStore) ResolveActorGrantScopes(ctx context.Context, organizationID, membershipID pgtype.UUID) ([]string, error) {
	return s.q.ResolveActorGrantScopes(ctx, database.ResolveActorGrantScopesParams{OrganizationID: organizationID, MembershipID: membershipID})
}
func (s *txStore) IncrementMembershipVersion(ctx context.Context, organizationID, membershipID pgtype.UUID) (int64, error) {
	row, err := s.q.IncrementMembershipAuthorizationVersion(ctx, database.IncrementMembershipAuthorizationVersionParams{OrganizationID: organizationID, MembershipID: membershipID})
	return row.AuthorizationVersion, err
}
func (s *txStore) ListSessionIDs(ctx context.Context, organizationID, membershipID pgtype.UUID) ([]pgtype.UUID, error) {
	return s.q.ListSessionIDsForMembership(ctx, database.ListSessionIDsForMembershipParams{OrganizationID: organizationID, MembershipID: membershipID})
}
func (s *txStore) ResolveScopes(ctx context.Context, params database.ResolveEffectiveScopesParams) (database.ResolveEffectiveScopesRow, error) {
	return s.q.ResolveEffectiveScopes(ctx, params)
}
func (s *txStore) AcceptInvitation(ctx context.Context, params database.AcceptInvitationParams) (database.AcceptInvitationRow, error) {
	return s.q.AcceptInvitation(ctx, params)
}
func (s *txStore) CreateSession(ctx context.Context, input sessions.CreateSessionInput) (sessions.CreateSessionResult, error) {
	return s.sessions.CreateSessionInTx(ctx, s.q, input)
}
func (s *txStore) WriteAudit(ctx context.Context, event audit.Event) error {
	if s.writer == nil {
		return errors.New("invitations: audit writer is required")
	}
	return s.writer.Write(ctx, s.q, event)
}
