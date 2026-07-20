package roles

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gabrielalc23/pdv/internal/audit"
	"github.com/gabrielalc23/pdv/internal/platform/authcontext"
	"github.com/gabrielalc23/pdv/internal/platform/authz"
	"github.com/gabrielalc23/pdv/internal/platform/clock"
	"github.com/gabrielalc23/pdv/internal/platform/database"
)

var (
	testOrganizationID = testUUID("00000000-0000-0000-0000-000000000001")
	testActorID        = testUUID("00000000-0000-0000-0000-000000000002")
	testMembershipID   = testUUID("00000000-0000-0000-0000-000000000003")
	testTargetID       = testUUID("00000000-0000-0000-0000-000000000004")
	testSessionID      = testUUID("00000000-0000-0000-0000-000000000005")
	testRoleID         = testUUID("00000000-0000-0000-0000-000000000006")
	testBindingID      = testUUID("00000000-0000-0000-0000-000000000007")
)

func TestValidateRoleScopes(t *testing.T) {
	actor := testActor(authz.ScopeRolesCreate)
	tests := []struct {
		name       string
		assignment database.RoleAssignmentScope
		catalog    []database.PermissionScope
		grants     []string
		wantErr    error
	}{
		{
			name:       "store role rejects organization scope",
			assignment: database.RoleAssignmentScopeSTORE,
			catalog:    []database.PermissionScope{{Code: "roles.read", ScopeLevel: database.PermissionScopeLevelORGANIZATION, IsAssignable: true}},
			grants:     []string{"roles.read"},
			wantErr:    ErrScopeLevelInvalid,
		},
		{
			name:       "custom role rejects non assignable scope",
			assignment: database.RoleAssignmentScopeORGANIZATION,
			catalog:    []database.PermissionScope{{Code: "organization.owners.manage", ScopeLevel: database.PermissionScopeLevelORGANIZATION}},
			grants:     []string{"organization.owners.manage"},
			wantErr:    ErrScopeNotAssignable,
		},
		{
			name:       "organization role accepts store scope",
			assignment: database.RoleAssignmentScopeORGANIZATION,
			catalog:    []database.PermissionScope{{Code: "sales.read", ScopeLevel: database.PermissionScopeLevelSTORE, IsAssignable: true}},
			grants:     []string{"sales.read"},
		},
		{
			name:       "actor cannot grant a scope absent from organization bindings",
			assignment: database.RoleAssignmentScopeORGANIZATION,
			catalog:    []database.PermissionScope{{Code: "sales.read", ScopeLevel: database.PermissionScopeLevelSTORE, IsAssignable: true}},
			wantErr:    ErrAuthorizationEscalation,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tx := &fakeTx{permissionScopes: test.catalog, actorGrantScopes: test.grants}
			err := validateRoleScopes(context.Background(), tx, actor, test.assignment, []string{test.catalog[0].Code})
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("validateRoleScopes() error = %v, want %v", err, test.wantErr)
			}
		})
	}
}

func TestCreateRoleIncrementsOrganizationVersionAndInvalidatesAfterCommit(t *testing.T) {
	now := time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC)
	tx := &fakeTx{
		permissionScopes: []database.PermissionScope{{Code: "sales.read", ScopeLevel: database.PermissionScopeLevelSTORE, IsAssignable: true}},
		actorGrantScopes: []string{"sales.read"},
		createdRole: database.Role{
			ID: testRoleID, Key: "supervisor", Name: "Supervisor", AssignmentScope: database.RoleAssignmentScopeORGANIZATION,
			IsMutable: true, IsActive: true, CreatedAt: pgtype.Timestamptz{Time: now, Valid: true}, UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true},
		},
		sessionIDs: []pgtype.UUID{testSessionID},
	}
	manager := &fakeTxProvider{tx: tx}
	invalidator := &fakeInvalidator{manager: manager}
	service, err := NewService(&fakeStore{}, manager, invalidator, clock.NewFakeClock(now))
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	_, err = service.CreateRole(context.Background(), testActor(authz.ScopeRolesCreate), UpsertRoleInput{
		Key: "supervisor", Name: "Supervisor", AssignmentScope: "ORGANIZATION", Scopes: []string{"sales.read"},
	})
	if err != nil {
		t.Fatalf("CreateRole() error = %v", err)
	}
	if !tx.organizationVersionIncremented || !tx.auditWritten {
		t.Fatal("expected authorization version increment and transactional audit")
	}
	if !manager.committed {
		t.Fatal("expected transaction commit")
	}
	if len(invalidator.ids) != 1 || invalidator.ids[0] != testSessionID {
		t.Fatalf("invalidated sessions = %v, want %v", invalidator.ids, testSessionID)
	}
	if invalidator.calledBeforeCommit {
		t.Fatal("session cache was invalidated before commit")
	}
}

func TestCreateBindingIsIdempotent(t *testing.T) {
	now := time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC)
	tx := &fakeTx{
		membership: database.OrganizationMembership{ID: testTargetID, OrganizationID: testOrganizationID, Status: database.MembershipStatusACTIVE},
		role: database.GetRoleWithScopesForOrganizationRow{
			ID: testRoleID, Key: "supervisor", Name: "Supervisor", AssignmentScope: database.RoleAssignmentScopeORGANIZATION,
			IsMutable: true, IsActive: true, ScopeCodes: []string{"sales.read"},
		},
		actorGrantScopes: []string{"sales.read"},
		bindings: []database.ListMemberRoleBindingsRow{{
			ID: testBindingID, OrganizationID: testOrganizationID, MembershipID: testTargetID, RoleID: testRoleID,
			CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
		}},
	}
	manager := &fakeTxProvider{tx: tx}
	invalidator := &fakeInvalidator{manager: manager}
	service, err := NewService(&fakeStore{}, manager, invalidator, clock.NewFakeClock(now))
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	response, created, err := service.CreateBinding(context.Background(), testActor(authz.ScopeRolesAssign), testTargetID.String(), CreateBindingInput{RoleID: testRoleID.String()})
	if err != nil {
		t.Fatalf("CreateBinding() error = %v", err)
	}
	if created || response.ID != testBindingID.String() {
		t.Fatalf("CreateBinding() = (%q, %v), want existing binding", response.ID, created)
	}
	if tx.bindingCreated || tx.membershipVersionIncremented || tx.auditWritten {
		t.Fatal("idempotent binding request mutated authorization state")
	}
	if len(invalidator.ids) != 0 {
		t.Fatalf("idempotent binding invalidated sessions: %v", invalidator.ids)
	}
}

func TestDeleteBindingProtectsLastOwner(t *testing.T) {
	now := time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC)
	tx := &fakeTx{
		membership: database.OrganizationMembership{ID: testTargetID, OrganizationID: testOrganizationID, Status: database.MembershipStatusACTIVE},
		binding: database.GetRoleBindingForOrganizationRow{
			ID: testBindingID, OrganizationID: testOrganizationID, MembershipID: testTargetID, RoleID: testRoleID,
			RoleKey: "owner", IsSystem: true, IsActive: true,
		},
		ownerCount: 1,
	}
	manager := &fakeTxProvider{tx: tx}
	service, err := NewService(&fakeStore{}, manager, nil, clock.NewFakeClock(now))
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	err = service.DeleteBinding(context.Background(), testActor(authz.ScopeRolesAssign, authz.ScopeOrganizationOwners), testTargetID.String(), testBindingID.String())
	if !errors.Is(err, ErrLastOwnerRequired) {
		t.Fatalf("DeleteBinding() error = %v, want %v", err, ErrLastOwnerRequired)
	}
	if !tx.organizationLocked || !tx.ownersCounted {
		t.Fatal("last-owner check did not lock the organization before counting owners")
	}
	if tx.bindingDeleted || manager.committed {
		t.Fatal("last owner binding was deleted or transaction committed")
	}
}

type fakeStore struct{ Store }

type fakeTxProvider struct {
	tx        TxStore
	committed bool
}

func (p *fakeTxProvider) WithTx(ctx context.Context, fn func(TxStore) error) error {
	if err := fn(p.tx); err != nil {
		return err
	}
	p.committed = true
	return nil
}

type fakeInvalidator struct {
	manager            *fakeTxProvider
	ids                []pgtype.UUID
	calledBeforeCommit bool
}

func (i *fakeInvalidator) InvalidateSession(_ context.Context, id pgtype.UUID) {
	if !i.manager.committed {
		i.calledBeforeCommit = true
	}
	i.ids = append(i.ids, id)
}

func (*fakeInvalidator) InvalidateOrganizationAuthorizationVersion(context.Context, pgtype.UUID) {}
func (*fakeInvalidator) InvalidateMembershipAuthorizationVersion(context.Context, pgtype.UUID)   {}

type fakeTx struct {
	TxStore
	permissionScopes               []database.PermissionScope
	actorGrantScopes               []string
	createdRole                    database.Role
	membership                     database.OrganizationMembership
	role                           database.GetRoleWithScopesForOrganizationRow
	bindings                       []database.ListMemberRoleBindingsRow
	binding                        database.GetRoleBindingForOrganizationRow
	createdBinding                 database.CreateRoleBindingRow
	sessionIDs                     []pgtype.UUID
	ownerCount                     int64
	organizationVersionIncremented bool
	membershipVersionIncremented   bool
	auditWritten                   bool
	bindingCreated                 bool
	bindingDeleted                 bool
	organizationLocked             bool
	ownersCounted                  bool
}

func (f *fakeTx) GetPermissionScopesByCodes(context.Context, []string) ([]database.PermissionScope, error) {
	return f.permissionScopes, nil
}

func (f *fakeTx) ResolveActorGrantScopes(context.Context, pgtype.UUID, pgtype.UUID) ([]string, error) {
	return f.actorGrantScopes, nil
}

func (f *fakeTx) CreateRole(context.Context, database.CreateRoleParams) (database.Role, error) {
	return f.createdRole, nil
}

func (f *fakeTx) ReplaceRoleScopes(context.Context, pgtype.UUID, pgtype.UUID, []string) error {
	return nil
}

func (f *fakeTx) LockRoleForScopeChange(context.Context, pgtype.UUID, pgtype.UUID) error {
	return nil
}

func (f *fakeTx) IncrementOrganizationAuthorizationVersion(context.Context, pgtype.UUID) error {
	f.organizationVersionIncremented = true
	return nil
}

func (f *fakeTx) GetMembershipForUpdate(context.Context, pgtype.UUID, pgtype.UUID) (database.OrganizationMembership, error) {
	return f.membership, nil
}

func (f *fakeTx) GetRoleWithScopes(context.Context, pgtype.UUID, pgtype.UUID) (database.GetRoleWithScopesForOrganizationRow, error) {
	return f.role, nil
}

func (f *fakeTx) ListMemberRoleBindings(context.Context, pgtype.UUID, pgtype.UUID) ([]database.ListMemberRoleBindingsRow, error) {
	return f.bindings, nil
}

func (f *fakeTx) CreateRoleBinding(context.Context, database.CreateRoleBindingParams) (database.CreateRoleBindingRow, error) {
	f.bindingCreated = true
	return f.createdBinding, nil
}

func (f *fakeTx) GetRoleBindingForUpdate(context.Context, pgtype.UUID, pgtype.UUID, pgtype.UUID) (database.GetRoleBindingForOrganizationRow, error) {
	return f.binding, nil
}

func (f *fakeTx) DeleteRoleBinding(context.Context, pgtype.UUID, pgtype.UUID, pgtype.UUID) (database.MembershipRoleBinding, error) {
	f.bindingDeleted = true
	return database.MembershipRoleBinding{ID: testBindingID, RoleID: testRoleID}, nil
}

func (f *fakeTx) LockOrganizationForOwnerChange(context.Context, pgtype.UUID) error {
	f.organizationLocked = true
	return nil
}

func (f *fakeTx) CountActiveOwnersForUpdate(context.Context, pgtype.UUID) (int64, error) {
	f.ownersCounted = true
	return f.ownerCount, nil
}

func (f *fakeTx) IncrementMembershipAuthorizationVersion(context.Context, pgtype.UUID, pgtype.UUID) error {
	f.membershipVersionIncremented = true
	return nil
}

func (f *fakeTx) ListSessionIDsForOrganization(context.Context, pgtype.UUID) ([]pgtype.UUID, error) {
	return f.sessionIDs, nil
}

func (f *fakeTx) ListSessionIDsForMembership(context.Context, pgtype.UUID, pgtype.UUID) ([]pgtype.UUID, error) {
	return f.sessionIDs, nil
}

func (f *fakeTx) WriteAudit(context.Context, audit.Event) error {
	f.auditWritten = true
	return nil
}

func testActor(scopes ...authcontext.Scope) authcontext.Principal {
	version := int64(1)
	return authcontext.Principal{
		UserID: testActorID, SessionID: testSessionID, ContextKind: authcontext.ContextOrganization,
		OrganizationID: testOrganizationID, MembershipID: testMembershipID,
		Scopes: authcontext.NewScopeSet(scopes...), OrgAuthzVersion: &version, MemberAuthzVersion: &version,
	}
}

func testUUID(value string) pgtype.UUID {
	var id pgtype.UUID
	if err := id.Scan(value); err != nil {
		panic(err)
	}
	return id
}
