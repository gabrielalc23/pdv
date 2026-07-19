package memberships

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gabrielalc23/pdv/internal/audit"
	"github.com/gabrielalc23/pdv/internal/platform/authcontext"
	"github.com/gabrielalc23/pdv/internal/platform/authz"
	"github.com/gabrielalc23/pdv/internal/platform/database"
)

func TestUpdateStatusRejectsLastActiveOwner(t *testing.T) {
	tx := &fakeTxStore{
		target:       membershipFixture(database.MembershipStatusACTIVE),
		bindings:     []database.ListMemberRoleBindingsRow{ownerBindingFixture()},
		activeOwners: 1,
	}
	service := NewService(nil, &fakeTxManager{tx: tx}, &fakeSessionRevoker{}, nil)

	_, err := service.UpdateStatus(context.Background(), actorFixture(authz.ScopeMembersStatusUpdate, authz.ScopeOrganizationOwners), tx.target.ID.String(), UpdateStatusInput{Status: "SUSPENDED"})
	if !errors.Is(err, ErrLastOwnerRequired) {
		t.Fatalf("UpdateStatus() error = %v, want LAST_OWNER_REQUIRED", err)
	}
	if !tx.ownerLockCalled {
		t.Fatal("expected active owners to be locked and counted")
	}
	if tx.updateCalled || tx.auditCalled {
		t.Fatal("last-owner rejection must not update or audit")
	}
}

func TestUpdateStatusRequiresOwnerManagementScopeForOwnerTarget(t *testing.T) {
	tx := &fakeTxStore{
		target:       membershipFixture(database.MembershipStatusACTIVE),
		bindings:     []database.ListMemberRoleBindingsRow{ownerBindingFixture()},
		activeOwners: 2,
	}
	service := NewService(nil, &fakeTxManager{tx: tx}, &fakeSessionRevoker{}, nil)

	_, err := service.UpdateStatus(context.Background(), actorFixture(authz.ScopeMembersStatusUpdate), tx.target.ID.String(), UpdateStatusInput{Status: "SUSPENDED"})
	if !errors.Is(err, ErrInsufficientScope) {
		t.Fatalf("UpdateStatus() error = %v, want INSUFFICIENT_SCOPE", err)
	}
	if !tx.ownerLockCalled || tx.updateCalled {
		t.Fatal("owner target must be checked under the owner lock and rejected before update")
	}
}

func TestUpdateStatusRevokesAndInvalidatesAfterCommit(t *testing.T) {
	target := membershipFixture(database.MembershipStatusACTIVE)
	sessionID := testUUID("10000000-0000-4000-8000-000000000009")
	tx := &fakeTxStore{target: target, activeOwners: 2}
	manager := &fakeTxManager{tx: tx}
	revoker := &fakeSessionRevoker{ids: []pgtype.UUID{sessionID}}
	invalidator := &fakeInvalidator{committed: &manager.committed}
	readStore := &fakeStore{detail: detailFixture(target, database.MembershipStatusSUSPENDED, 2)}
	service := NewService(readStore, manager, revoker, invalidator)

	result, err := service.UpdateStatus(context.Background(), actorFixture(authz.ScopeMembersStatusUpdate), target.ID.String(), UpdateStatusInput{Status: "suspended"})
	if err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}
	if result.Status != "SUSPENDED" || result.AuthorizationVersion != 2 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if !revoker.called || !tx.updateCalled || !tx.auditCalled {
		t.Fatal("expected update, session revocation, and transactional audit")
	}
	if len(invalidator.ids) != 1 || invalidator.ids[0] != sessionID {
		t.Fatalf("invalidated sessions = %v", invalidator.ids)
	}
	if invalidator.beforeCommit {
		t.Fatal("session cache was invalidated before transaction commit")
	}
}

func TestUpdateStatusRemovedIsTerminal(t *testing.T) {
	target := membershipFixture(database.MembershipStatusREMOVED)
	tx := &fakeTxStore{target: target}
	service := NewService(nil, &fakeTxManager{tx: tx}, &fakeSessionRevoker{}, nil)

	_, err := service.UpdateStatus(context.Background(), actorFixture(authz.ScopeMembersStatusUpdate), target.ID.String(), UpdateStatusInput{Status: "ACTIVE"})
	if !errors.Is(err, ErrInvalidStatusTransition) {
		t.Fatalf("UpdateStatus() error = %v, want invalid transition", err)
	}
}

func TestUpdateDefaultStoreIncrementsVersionAndInvalidatesAfterCommit(t *testing.T) {
	target := membershipFixture(database.MembershipStatusACTIVE)
	storeID := testUUID("10000000-0000-4000-8000-000000000007")
	sessionID := testUUID("10000000-0000-4000-8000-000000000009")
	tx := &fakeTxStore{
		target:     target,
		stores:     []database.Store{{ID: storeID, OrganizationID: target.OrganizationID, Status: database.StoreStatusACTIVE}},
		sessionIDs: []pgtype.UUID{sessionID},
	}
	manager := &fakeTxManager{tx: tx}
	invalidator := &fakeInvalidator{committed: &manager.committed}
	detail := detailFixture(target, database.MembershipStatusACTIVE, 2)
	detail.DefaultStoreID = storeID
	detail.DefaultStoreCode = pgtype.Text{String: "MAIN", Valid: true}
	detail.DefaultStoreName = pgtype.Text{String: "Main", Valid: true}
	readStore := &fakeStore{detail: detail}
	service := NewService(readStore, manager, &fakeSessionRevoker{}, invalidator)
	storeIDString := storeID.String()

	result, err := service.UpdateDefaultStore(context.Background(), actorFixture(authz.ScopeMembersStatusUpdate), target.ID.String(), UpdateDefaultStoreInput{DefaultStoreID: &storeIDString})
	if err != nil {
		t.Fatalf("UpdateDefaultStore() error = %v", err)
	}
	if result.DefaultStore == nil || result.DefaultStore.ID != storeID.String() || result.AuthorizationVersion != 2 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if !tx.defaultStoreUpdateCalled || !tx.auditCalled {
		t.Fatal("expected authorization version update and transactional audit")
	}
	if len(invalidator.ids) != 1 || invalidator.ids[0] != sessionID || invalidator.beforeCommit {
		t.Fatalf("post-commit invalidation = %+v", invalidator)
	}
}

type fakeStore struct {
	Store
	detail database.GetMembershipForOrganizationRow
}

func (s *fakeStore) GetMembershipForOrganization(context.Context, database.GetMembershipForOrganizationParams) (database.GetMembershipForOrganizationRow, error) {
	return s.detail, nil
}

func (s *fakeStore) ListMemberRoleBindings(context.Context, database.ListMemberRoleBindingsParams) ([]database.ListMemberRoleBindingsRow, error) {
	return []database.ListMemberRoleBindingsRow{}, nil
}

type fakeTxManager struct {
	tx        TxStore
	committed bool
}

func (m *fakeTxManager) WithTx(ctx context.Context, fn func(TxStore) error) error {
	if err := fn(m.tx); err != nil {
		return err
	}
	m.committed = true
	return nil
}

type fakeTxStore struct {
	TxStore
	target                   database.OrganizationMembership
	bindings                 []database.ListMemberRoleBindingsRow
	stores                   []database.Store
	sessionIDs               []pgtype.UUID
	activeOwners             int64
	ownerLockCalled          bool
	updateCalled             bool
	defaultStoreUpdateCalled bool
	auditCalled              bool
}

func (s *fakeTxStore) GetMembershipForUpdate(context.Context, database.GetMembershipForUpdateParams) (database.OrganizationMembership, error) {
	return s.target, nil
}

func (s *fakeTxStore) ListMemberRoleBindings(context.Context, database.ListMemberRoleBindingsParams) ([]database.ListMemberRoleBindingsRow, error) {
	return s.bindings, nil
}

func (s *fakeTxStore) LockActiveOwnerMembershipsForUpdate(context.Context, pgtype.UUID) (int64, error) {
	s.ownerLockCalled = true
	return s.activeOwners, nil
}

func (s *fakeTxStore) ListStoresForMembership(context.Context, database.ListStoresForMembershipParams) ([]database.Store, error) {
	return s.stores, nil
}

func (s *fakeTxStore) ListSessionIDsForMembership(context.Context, database.ListSessionIDsForMembershipParams) ([]pgtype.UUID, error) {
	return s.sessionIDs, nil
}

func (s *fakeTxStore) UpdateMembershipDefaultStore(_ context.Context, _, _ pgtype.UUID, storeID pgtype.UUID) (database.OrganizationMembership, error) {
	s.defaultStoreUpdateCalled = true
	s.target.DefaultStoreID = storeID
	s.target.AuthorizationVersion++
	return s.target, nil
}

func (s *fakeTxStore) UpdateMembershipStatus(_ context.Context, params database.UpdateMembershipStatusParams) (database.OrganizationMembership, error) {
	s.updateCalled = true
	s.target.Status = params.Status
	s.target.AuthorizationVersion++
	return s.target, nil
}

func (s *fakeTxStore) WriteAudit(context.Context, audit.Event) error {
	s.auditCalled = true
	return nil
}

type fakeSessionRevoker struct {
	ids    []pgtype.UUID
	called bool
}

func (r *fakeSessionRevoker) RevokeMembershipSessions(context.Context, TxStore, pgtype.UUID, pgtype.UUID, pgtype.UUID, string) ([]pgtype.UUID, error) {
	r.called = true
	return r.ids, nil
}

type fakeInvalidator struct {
	committed    *bool
	beforeCommit bool
	ids          []pgtype.UUID
}

func (i *fakeInvalidator) InvalidateSession(_ context.Context, id pgtype.UUID) {
	if i.committed != nil && !*i.committed {
		i.beforeCommit = true
	}
	i.ids = append(i.ids, id)
}

func (i *fakeInvalidator) InvalidateMembershipAuthorizationVersion(context.Context, pgtype.UUID) {}

func actorFixture(scopes ...authcontext.Scope) Actor {
	return Actor{
		UserID:         testUUID("10000000-0000-4000-8000-000000000001"),
		SessionID:      testUUID("10000000-0000-4000-8000-000000000002"),
		OrganizationID: testUUID("10000000-0000-4000-8000-000000000003"),
		MembershipID:   testUUID("10000000-0000-4000-8000-000000000004"),
		Scopes:         authcontext.NewScopeSet(scopes...),
	}
}

func membershipFixture(status database.MembershipStatus) database.OrganizationMembership {
	now := pgtype.Timestamptz{Time: time.Now(), Valid: true}
	return database.OrganizationMembership{
		ID:                   testUUID("10000000-0000-4000-8000-000000000005"),
		OrganizationID:       testUUID("10000000-0000-4000-8000-000000000003"),
		UserID:               testUUID("10000000-0000-4000-8000-000000000006"),
		Status:               status,
		AuthorizationVersion: 1,
		JoinedAt:             now,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
}

func ownerBindingFixture() database.ListMemberRoleBindingsRow {
	return database.ListMemberRoleBindingsRow{RoleKey: "owner", IsSystem: true, IsActive: true}
}

func detailFixture(target database.OrganizationMembership, status database.MembershipStatus, version int64) database.GetMembershipForOrganizationRow {
	return database.GetMembershipForOrganizationRow{
		ID:                   target.ID,
		OrganizationID:       target.OrganizationID,
		UserID:               target.UserID,
		Status:               status,
		AuthorizationVersion: version,
		JoinedAt:             target.JoinedAt,
		CreatedAt:            target.CreatedAt,
		UpdatedAt:            target.UpdatedAt,
		Email:                "member@example.com",
		DisplayName:          "Member",
		UserStatus:           database.UserStatusACTIVE,
	}
}

func testUUID(value string) pgtype.UUID {
	var id pgtype.UUID
	if err := id.Scan(value); err != nil {
		panic(err)
	}
	return id
}
