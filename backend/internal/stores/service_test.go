package stores

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gabrielalc23/pdv/internal/audit"
	"github.com/gabrielalc23/pdv/internal/platform/authcontext"
	"github.com/gabrielalc23/pdv/internal/platform/authz"
	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type fakeRepository struct {
	getStore func(context.Context, database.GetStoreForOrganizationParams) (database.Store, error)
}

func (f *fakeRepository) GetStoreForOrganization(ctx context.Context, params database.GetStoreForOrganizationParams) (database.Store, error) {
	if f.getStore != nil {
		return f.getStore(ctx, params)
	}
	return database.Store{}, nil
}

func (f *fakeRepository) ListStoresForOrganization(context.Context, database.ListStoresForOrganizationParams) ([]database.Store, error) {
	return nil, nil
}

func (f *fakeRepository) CountStoresForOrganization(context.Context, database.CountStoresForOrganizationParams) (int64, error) {
	return 0, nil
}

func (f *fakeRepository) ListPaymentMethodsForOrganization(context.Context, pgtype.UUID) ([]database.PaymentMethod, error) {
	return nil, nil
}

func (f *fakeRepository) ListStorePaymentMethods(context.Context, database.ListStorePaymentMethodsParams) ([]database.ListStorePaymentMethodsRow, error) {
	return nil, nil
}

type fakeQuerier struct {
	database.Querier
	createStore       func(database.CreateStoreParams) (database.Store, error)
	copyMethods       func(database.CopyActivePaymentMethodsToStoreParams) error
	lockOrganization  func(pgtype.UUID) error
	lockStore         func(database.LockStoreForStatusChangeParams) (database.Store, error)
	hasOpenSales      func(database.HasOpenSalesForStoreParams) (bool, error)
	countActiveStores func(pgtype.UUID) (int64, error)
	updateStatus      func(database.UpdateStoreStatusParams) (database.Store, error)
}

func (f *fakeQuerier) CreateStore(_ context.Context, params database.CreateStoreParams) (database.Store, error) {
	return f.createStore(params)
}

func (f *fakeQuerier) CopyActivePaymentMethodsToStore(_ context.Context, params database.CopyActivePaymentMethodsToStoreParams) ([]database.StorePaymentMethod, error) {
	return nil, f.copyMethods(params)
}

func (f *fakeQuerier) LockOrganizationForOwnerChange(_ context.Context, id pgtype.UUID) (database.LockOrganizationForOwnerChangeRow, error) {
	return database.LockOrganizationForOwnerChangeRow{}, f.lockOrganization(id)
}

func (f *fakeQuerier) LockStoreForStatusChange(_ context.Context, params database.LockStoreForStatusChangeParams) (database.Store, error) {
	return f.lockStore(params)
}

func (f *fakeQuerier) HasOpenSalesForStore(_ context.Context, params database.HasOpenSalesForStoreParams) (bool, error) {
	return f.hasOpenSales(params)
}

func (f *fakeQuerier) CountActiveStores(_ context.Context, id pgtype.UUID) (int64, error) {
	return f.countActiveStores(id)
}

func (f *fakeQuerier) UpdateStoreStatus(_ context.Context, params database.UpdateStoreStatusParams) (database.Store, error) {
	return f.updateStatus(params)
}

type fakeTxManager struct {
	q    database.Querier
	inTx bool
}

func (f *fakeTxManager) WithTx(ctx context.Context, fn func(database.Querier) error) error {
	f.inTx = true
	defer func() { f.inTx = false }()
	return fn(f.q)
}

type fakeAuditWriter struct {
	txManager *fakeTxManager
	events    []audit.Event
	queries   []database.Querier
}

func (f *fakeAuditWriter) Write(_ context.Context, q database.Querier, event audit.Event) error {
	if f.txManager != nil && !f.txManager.inTx {
		return errors.New("audit was written outside transaction")
	}
	f.queries = append(f.queries, q)
	f.events = append(f.events, event)
	return nil
}

func TestCreateStoreUsesPrincipalTenantAndAuditsInTransaction(t *testing.T) {
	principal := testPrincipal()
	storeID := testUUID("20000000-0000-0000-0000-000000000001")
	q := &fakeQuerier{}
	q.createStore = func(params database.CreateStoreParams) (database.Store, error) {
		if params.OrganizationID != principal.OrganizationID {
			t.Fatalf("organization ID = %s, want principal organization %s", params.OrganizationID.String(), principal.OrganizationID.String())
		}
		if params.CreatedByUserID != principal.UserID {
			t.Fatalf("creator ID = %s, want principal user %s", params.CreatedByUserID.String(), principal.UserID.String())
		}
		return database.Store{ID: storeID, OrganizationID: params.OrganizationID, Code: params.Code, Name: params.Name, Timezone: params.Timezone, Status: database.StoreStatusACTIVE}, nil
	}
	q.copyMethods = func(params database.CopyActivePaymentMethodsToStoreParams) error {
		if params.OrganizationID != principal.OrganizationID || params.StoreID != storeID {
			t.Fatalf("copy methods received wrong tenant keys")
		}
		return nil
	}
	tx := &fakeTxManager{q: q}
	writer := &fakeAuditWriter{txManager: tx}
	service, err := NewServiceWithDependencies(&fakeRepository{}, tx, writer)
	if err != nil {
		t.Fatalf("NewServiceWithDependencies() error = %v", err)
	}

	result, err := service.CreateStore(context.Background(), principal, CreateStoreInput{Code: "sp-01", Name: "Centro", Timezone: "America/Sao_Paulo"})
	if err != nil {
		t.Fatalf("CreateStore() error = %v", err)
	}
	if result.Code != "SP-01" {
		t.Fatalf("Code = %q, want SP-01", result.Code)
	}
	if len(writer.events) != 1 || writer.events[0].EventType != audit.EventStoreCreated {
		t.Fatalf("audit events = %#v", writer.events)
	}
	if len(writer.queries) != 1 || writer.queries[0] != q {
		t.Fatal("audit did not receive the transaction querier")
	}
}

func TestDeactivateStoreRejectsOpenSalesBeforeMutation(t *testing.T) {
	service, q, writer := statusTestService(t, true, 2)

	_, err := service.DeactivateStore(context.Background(), testPrincipal(), testUUID("20000000-0000-0000-0000-000000000001").String())
	if !errors.Is(err, ErrStoreHasOpenSales) {
		t.Fatalf("DeactivateStore() error = %v, want ErrStoreHasOpenSales", err)
	}
	if q.updateStatus != nil {
		t.Fatal("status update was configured unexpectedly")
	}
	if len(writer.events) != 0 {
		t.Fatal("failed mutation must not write a success audit event")
	}
}

func TestDeactivateStoreRejectsLastActiveStore(t *testing.T) {
	service, _, writer := statusTestService(t, false, 1)

	_, err := service.DeactivateStore(context.Background(), testPrincipal(), testUUID("20000000-0000-0000-0000-000000000001").String())
	if !errors.Is(err, ErrLastActiveStore) {
		t.Fatalf("DeactivateStore() error = %v, want ErrLastActiveStore", err)
	}
	if len(writer.events) != 0 {
		t.Fatal("failed mutation must not write a success audit event")
	}
}

func TestRoutesRequireOrganizationContextAndScope(t *testing.T) {
	router := chi.NewRouter()
	RegisterRoutes(router, NewHandler(nil), authz.NewGuard())

	request := httptest.NewRequest(http.MethodGet, "/stores", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("without principal status = %d, want 401", response.Code)
	}

	principal := testPrincipal()
	request = httptest.NewRequest(http.MethodGet, "/stores", nil)
	request = request.WithContext(authcontext.SetPrincipal(request.Context(), principal))
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("without scope status = %d, want 403", response.Code)
	}
}

func TestNormalizePaymentMethodEnforcesDatabaseRules(t *testing.T) {
	tests := []struct {
		name  string
		input UpsertPaymentMethodInput
		field string
	}{
		{"change only for cash", UpsertPaymentMethodInput{Code: "CARD", Name: "Card", Kind: "DEBIT_CARD", AllowsChange: true, MaxInstallments: 1}, "allowsChange"},
		{"provider required", UpsertPaymentMethodInput{Code: "PIX", Name: "PIX", Kind: "PIX", RequiresExternalReference: true, MaxInstallments: 1}, "provider"},
		{"credit installments", UpsertPaymentMethodInput{Code: "CREDIT", Name: "Credit", Kind: "CREDIT_CARD", AllowsInstallments: true, MaxInstallments: 1}, "maxInstallments"},
		{"fee scale", UpsertPaymentMethodInput{Code: "CASH", Name: "Cash", Kind: "CASH", MaxInstallments: 1, FeePercentage: "1.00001"}, "feePercentage"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := normalizePaymentMethodInput(test.input, true)
			var validation *ValidationError
			if !errors.As(err, &validation) || validation.Field != test.field {
				t.Fatalf("error = %v, want validation field %q", err, test.field)
			}
		})
	}
}

func statusTestService(t *testing.T, openSales bool, activeCount int64) (*Service, *fakeQuerier, *fakeAuditWriter) {
	t.Helper()
	principal := testPrincipal()
	storeID := testUUID("20000000-0000-0000-0000-000000000001")
	q := &fakeQuerier{}
	q.lockOrganization = func(id pgtype.UUID) error {
		if id != principal.OrganizationID {
			t.Fatal("organization lock did not use principal tenant")
		}
		return nil
	}
	q.lockStore = func(params database.LockStoreForStatusChangeParams) (database.Store, error) {
		if params.OrganizationID != principal.OrganizationID || params.StoreID != storeID {
			t.Fatal("store lock did not use principal tenant keys")
		}
		return database.Store{ID: storeID, OrganizationID: principal.OrganizationID, Status: database.StoreStatusACTIVE}, nil
	}
	q.hasOpenSales = func(params database.HasOpenSalesForStoreParams) (bool, error) { return openSales, nil }
	q.countActiveStores = func(id pgtype.UUID) (int64, error) { return activeCount, nil }
	tx := &fakeTxManager{q: q}
	writer := &fakeAuditWriter{txManager: tx}
	service, err := NewServiceWithDependencies(&fakeRepository{}, tx, writer)
	if err != nil {
		t.Fatalf("NewServiceWithDependencies() error = %v", err)
	}
	return service, q, writer
}

func testPrincipal() authcontext.Principal {
	return authcontext.Principal{
		UserID: testUUID("10000000-0000-0000-0000-000000000001"), SessionID: testUUID("10000000-0000-0000-0000-000000000002"),
		OrganizationID: testUUID("10000000-0000-0000-0000-000000000003"), MembershipID: testUUID("10000000-0000-0000-0000-000000000004"),
		ContextKind: authcontext.ContextOrganization, Scopes: authcontext.NewScopeSet(),
	}
}

func testUUID(value string) pgtype.UUID {
	var id pgtype.UUID
	if err := id.Scan(value); err != nil {
		panic(err)
	}
	return id
}
