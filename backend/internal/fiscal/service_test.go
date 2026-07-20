package fiscal_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gabrielalc23/pdv/internal/fiscal"
	"github.com/gabrielalc23/pdv/internal/platform/authn"
	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/gabrielalc23/pdv/internal/platform/tenancy"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

var testActor = authn.StoreActor{
	UserID:         mustUUID("00000000-0000-0000-0000-000000000001"),
	SessionID:      mustUUID("00000000-0000-0000-0000-000000000001"),
	OrganizationID: mustUUID("00000000-0000-0000-0000-000000000001"),
	MembershipID:   mustUUID("00000000-0000-0000-0000-000000000001"),
	StoreID:        mustUUID("00000000-0000-0000-0000-000000000002"),
	ClientID:       "test-client",
}

func TestGetBySaleIDReturnsFiscalDocument(t *testing.T) {
	saleID := mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8a9")
	docID := mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8aa")
	svc := fiscal.NewService(&fiscalFakeStore{
		getSaleByIDFn: func(context.Context, tenancy.StoreScope, pgtype.UUID) (database.Sale, error) {
			return saleFixtureRow(saleID, database.SaleStatusCOMPLETED), nil
		},
		getFiscalDocumentBySaleIDFn: func(context.Context, tenancy.StoreScope, pgtype.UUID) (database.FiscalDocument, error) {
			return fiscalDocumentRowFixture(docID, saleID, database.FiscalDocumentStatusAUTHORIZED), nil
		},
	}, nil)

	resp, err := svc.GetBySaleID(context.Background(), testActor, saleID.String())
	if err != nil {
		t.Fatalf("GetBySaleID returned error: %v", err)
	}
	if resp.Status != string(database.FiscalDocumentStatusAUTHORIZED) {
		t.Fatalf("unexpected fiscal document: %+v", resp)
	}
}

func TestGetBySaleIDDocumentNotFound(t *testing.T) {
	saleID := mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8a9")
	svc := fiscal.NewService(&fiscalFakeStore{
		getSaleByIDFn: func(context.Context, tenancy.StoreScope, pgtype.UUID) (database.Sale, error) {
			return saleFixtureRow(saleID, database.SaleStatusCOMPLETED), nil
		},
		getFiscalDocumentBySaleIDFn: func(context.Context, tenancy.StoreScope, pgtype.UUID) (database.FiscalDocument, error) {
			return database.FiscalDocument{}, pgx.ErrNoRows
		},
	}, nil)

	_, err := svc.GetBySaleID(context.Background(), testActor, saleID.String())
	if !errors.Is(err, fiscal.ErrFiscalDocumentNotFound) {
		t.Fatalf("expected fiscal.ErrFiscalDocumentNotFound, got %v", err)
	}
}

func TestAuthorizeSuccessUpdatesDocument(t *testing.T) {
	saleID := mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8a9")
	docID := mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8aa")
	authorizedAt := time.Date(2026, 7, 15, 12, 10, 0, 0, time.UTC)
	store := &fiscalFakeStore{
		getSaleByIDFn: func(context.Context, tenancy.StoreScope, pgtype.UUID) (database.Sale, error) {
			return saleFixtureRow(saleID, database.SaleStatusCOMPLETED), nil
		},
		getFiscalDocumentBySaleIDFn: func(context.Context, tenancy.StoreScope, pgtype.UUID) (database.FiscalDocument, error) {
			return fiscalDocumentRowFixture(docID, saleID, database.FiscalDocumentStatusPENDING), nil
		},
		markFiscalDocumentAuthorizedFn: func(_ context.Context, _ tenancy.StoreScope, arg database.MarkFiscalDocumentAuthorizedForStoreParams) (database.FiscalDocument, error) {
			if arg.ID.String() != docID.String() {
				t.Fatalf("unexpected id: %s", arg.ID.String())
			}
			return authorizedFiscalDocumentFixture(docID, saleID, authorizedAt), nil
		},
	}

	svc := fiscal.NewService(store, &fiscal.MockProvider{Now: func() time.Time { return authorizedAt }})
	resp, err := svc.Authorize(context.Background(), testActor, saleID.String(), fiscal.AuthorizationInput{SaleID: saleID.String(), SaleNumber: 77, SaleTotal: "100.00"})
	if err != nil {
		t.Fatalf("Authorize returned error: %v", err)
	}
	if resp.Status != string(database.FiscalDocumentStatusAUTHORIZED) {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if store.markAuthorizedCalls != 1 {
		t.Fatalf("expected authorized update, got %d", store.markAuthorizedCalls)
	}
}

func TestAuthorizeFailureUpdatesDocumentError(t *testing.T) {
	saleID := mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8a9")
	docID := mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8aa")
	store := &fiscalFakeStore{
		getSaleByIDFn: func(context.Context, tenancy.StoreScope, pgtype.UUID) (database.Sale, error) {
			return saleFixtureRow(saleID, database.SaleStatusCOMPLETED), nil
		},
		getFiscalDocumentBySaleIDFn: func(context.Context, tenancy.StoreScope, pgtype.UUID) (database.FiscalDocument, error) {
			return fiscalDocumentRowFixture(docID, saleID, database.FiscalDocumentStatusPENDING), nil
		},
		markFiscalDocumentErrorFn: func(_ context.Context, _ tenancy.StoreScope, arg database.MarkFiscalDocumentErrorForStoreParams) (database.FiscalDocument, error) {
			if arg.ID.String() != docID.String() {
				t.Fatalf("unexpected id: %s", arg.ID.String())
			}
			return errorFiscalDocumentFixture(docID, saleID), nil
		},
	}

	svc := fiscal.NewService(store, &fiscal.MockProvider{Fail: true})
	resp, err := svc.Authorize(context.Background(), testActor, saleID.String(), fiscal.AuthorizationInput{SaleID: saleID.String(), SaleNumber: 77, SaleTotal: "100.00"})
	if !errors.Is(err, fiscal.ErrFiscalAuthorizationFailed) {
		t.Fatalf("expected fiscal.ErrFiscalAuthorizationFailed, got %v", err)
	}
	if resp.Status != string(database.FiscalDocumentStatusERROR) {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if store.markErrorCalls != 1 {
		t.Fatalf("expected error update, got %d", store.markErrorCalls)
	}
}

func TestMockProviderAuthorizeFailure(t *testing.T) {
	p := &fiscal.MockProvider{Fail: true}
	_, err := p.Authorize(context.Background(), fiscal.AuthorizationInput{SaleID: "1", SaleNumber: 1, SaleTotal: "10.00"})
	if err == nil {
		t.Fatalf("expected error")
	}
}

type fiscalFakeStore struct {
	getSaleByIDFn                  func(context.Context, tenancy.StoreScope, pgtype.UUID) (database.Sale, error)
	getFiscalDocumentBySaleIDFn    func(context.Context, tenancy.StoreScope, pgtype.UUID) (database.FiscalDocument, error)
	getFiscalDocumentByIDFn        func(context.Context, tenancy.StoreScope, pgtype.UUID) (database.FiscalDocument, error)
	markFiscalDocumentAuthorizedFn func(context.Context, tenancy.StoreScope, database.MarkFiscalDocumentAuthorizedForStoreParams) (database.FiscalDocument, error)
	markFiscalDocumentErrorFn      func(context.Context, tenancy.StoreScope, database.MarkFiscalDocumentErrorForStoreParams) (database.FiscalDocument, error)
	markAuthorizedCalls            int
	markErrorCalls                 int
}

func (f *fiscalFakeStore) GetSaleByID(ctx context.Context, scope tenancy.StoreScope, saleID pgtype.UUID) (database.Sale, error) {
	if f.getSaleByIDFn != nil {
		return f.getSaleByIDFn(ctx, scope, saleID)
	}
	return database.Sale{}, pgx.ErrNoRows
}

func (f *fiscalFakeStore) GetFiscalDocumentBySaleID(ctx context.Context, scope tenancy.StoreScope, saleID pgtype.UUID) (database.FiscalDocument, error) {
	if f.getFiscalDocumentBySaleIDFn != nil {
		return f.getFiscalDocumentBySaleIDFn(ctx, scope, saleID)
	}
	return database.FiscalDocument{}, pgx.ErrNoRows
}

func (f *fiscalFakeStore) GetFiscalDocumentByID(ctx context.Context, scope tenancy.StoreScope, id pgtype.UUID) (database.FiscalDocument, error) {
	if f.getFiscalDocumentByIDFn != nil {
		return f.getFiscalDocumentByIDFn(ctx, scope, id)
	}
	return database.FiscalDocument{}, pgx.ErrNoRows
}

func (f *fiscalFakeStore) MarkFiscalDocumentAuthorized(ctx context.Context, scope tenancy.StoreScope, arg database.MarkFiscalDocumentAuthorizedForStoreParams) (database.FiscalDocument, error) {
	f.markAuthorizedCalls++
	if f.markFiscalDocumentAuthorizedFn != nil {
		return f.markFiscalDocumentAuthorizedFn(ctx, scope, arg)
	}
	return database.FiscalDocument{}, nil
}

func (f *fiscalFakeStore) MarkFiscalDocumentError(ctx context.Context, scope tenancy.StoreScope, arg database.MarkFiscalDocumentErrorForStoreParams) (database.FiscalDocument, error) {
	f.markErrorCalls++
	if f.markFiscalDocumentErrorFn != nil {
		return f.markFiscalDocumentErrorFn(ctx, scope, arg)
	}
	return database.FiscalDocument{}, nil
}

func (f *fiscalFakeStore) LockFiscalDocumentBySaleID(ctx context.Context, scope tenancy.StoreScope, saleID pgtype.UUID) (database.FiscalDocument, error) {
	return f.GetFiscalDocumentBySaleID(ctx, scope, saleID)
}

func authorizedFiscalDocumentFixture(id, saleID pgtype.UUID, at time.Time) database.FiscalDocument {
	row := fiscalDocumentRowFixture(id, saleID, database.FiscalDocumentStatusAUTHORIZED)
	row.AccessKey = pgtype.Text{String: "12345678901234567890123456789012345678901234", Valid: true}
	row.Protocol = pgtype.Text{String: "MOCK-77", Valid: true}
	row.Provider = pgtype.Text{String: "mock", Valid: true}
	row.ExternalReference = pgtype.Text{String: "sale-" + saleID.String(), Valid: true}
	row.XML = pgtype.Text{String: "<fiscal />", Valid: true}
	row.IssuedAt = timestamptz(at)
	return row
}

func errorFiscalDocumentFixture(id, saleID pgtype.UUID) database.FiscalDocument {
	row := fiscalDocumentRowFixture(id, saleID, database.FiscalDocumentStatusERROR)
	row.ErrorCode = pgtype.Text{String: "mock_authorization_failed", Valid: true}
	row.ErrorMessage = pgtype.Text{String: "Fiscal authorization failed", Valid: true}
	return row
}

func saleFixtureRow(id pgtype.UUID, status database.SaleStatus) database.Sale {
	now := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	row := database.Sale{
		ID:             id,
		Number:         77,
		Status:         status,
		Subtotal:       mustNumeric("100.00"),
		Discount:       mustNumeric("0.00"),
		Addition:       mustNumeric("0.00"),
		Total:          mustNumeric("100.00"),
		OpenedAt:       timestamptz(now),
		CreatedAt:      timestamptz(now),
		UpdatedAt:      timestamptz(now),
		IdempotencyKey: "sale-1",
	}
	if status == database.SaleStatusCOMPLETED {
		row.CompletedAt = timestamptz(now.Add(time.Minute))
	}
	if status == database.SaleStatusCANCELLED {
		row.CancelledAt = timestamptz(now.Add(time.Minute))
	}
	return row
}

func fiscalDocumentRowFixture(id, saleID pgtype.UUID, status database.FiscalDocumentStatus) database.FiscalDocument {
	now := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	return database.FiscalDocument{
		ID:            id,
		SaleID:        saleID,
		Status:        status,
		Environment:   database.FiscalEnvironmentHOMOLOGATION,
		DocumentModel: 65,
		CreatedAt:     timestamptz(now),
		UpdatedAt:     timestamptz(now),
	}
}

func mustUUID(raw string) pgtype.UUID {
	var id pgtype.UUID
	if err := id.Scan(raw); err != nil {
		panic(err)
	}
	return id
}

func mustNumeric(raw string) pgtype.Numeric {
	var n pgtype.Numeric
	if err := n.Scan(raw); err != nil {
		panic(err)
	}
	return n
}

func timestamptz(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: value.UTC(), Valid: true}
}
