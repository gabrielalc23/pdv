package receipt_test

import (
	"context"
	"errors"
	"testing"

	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/gabrielalc23/pdv/internal/platform/tenancy"
	"github.com/gabrielalc23/pdv/internal/receipt"
	"github.com/gabrielalc23/pdv/tests/testutil"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

var testScope = tenancy.StoreScope{
	OrganizationID: testutil.UUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b800"),
	StoreID:        testutil.UUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b801"),
}

func TestGetReceiptReturnsCompletedSaleWithCalculatedSubtotal(t *testing.T) {
	saleID := testutil.UUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8a9")
	docID := testutil.UUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8ac")
	svc := receipt.NewService(&testutil.FakeStore{
		GetSaleByIDFn: func(_ context.Context, _ tenancy.StoreScope, _ pgtype.UUID) (database.Sale, error) {
			return testutil.SaleFixtureRow(saleID, database.SaleStatusCOMPLETED), nil
		},
		ListSaleItemsBySaleIDFn: func(_ context.Context, _ tenancy.StoreScope, _ pgtype.UUID) ([]database.SaleItem, error) {
			return []database.SaleItem{
				testutil.SaleItemFixture(testutil.UUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8b0"), saleID, "12.34", "2.000", "4.68", "20.00", "ABC-001"),
				testutil.SaleItemFixture(testutil.UUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8b1"), saleID, "5.00", "1.000", "0.00", "5.00", "XYZ-002"),
			}, nil
		},
		ListReceiptPaymentsBySaleIDFn: func(_ context.Context, _ tenancy.StoreScope, _ pgtype.UUID) ([]database.ReceiptPayment, error) {
			return []database.ReceiptPayment{
				testutil.ReceiptPaymentFixture(testutil.UUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8c0"), saleID, "40.00", "40.00", "0.00", 1, "Cash", 2),
				testutil.ReceiptPaymentFixture(testutil.UUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8c1"), saleID, "60.00", "60.00", "0.00", 1, "PIX", 1),
			}, nil
		},
		GetFiscalDocumentBySaleIDFn: func(_ context.Context, _ tenancy.StoreScope, _ pgtype.UUID) (database.FiscalDocument, error) {
			return testutil.FiscalDocumentFixture(docID, saleID, database.FiscalDocumentStatusAUTHORIZED), nil
		},
	})

	resp, err := svc.Get(context.Background(), testScope, saleID.String())
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if resp.Sale.Status != string(database.SaleStatusCOMPLETED) {
		t.Fatalf("unexpected sale: %+v", resp.Sale)
	}
	if len(resp.Items) != 2 {
		t.Fatalf("unexpected items: %+v", resp.Items)
	}
	if resp.Items[0].Subtotal != "24.68" {
		t.Fatalf("subtotal should be unit price * quantity, got %+v", resp.Items[0])
	}
	if resp.Items[0].Total != "20.00" {
		t.Fatalf("unexpected total: %+v", resp.Items[0])
	}
	if len(resp.Payments) != 2 || resp.Payments[0].Method != "Cash" || resp.Payments[1].Method != "PIX" {
		t.Fatalf("unexpected payments ordering: %+v", resp.Payments)
	}
	if resp.FiscalDocument == nil || resp.FiscalDocument.Status != string(database.FiscalDocumentStatusAUTHORIZED) {
		t.Fatalf("unexpected fiscal document: %+v", resp.FiscalDocument)
	}
}

func TestGetReceiptUnavailableForOpenOrCancelled(t *testing.T) {
	saleID := testutil.UUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8a9")
	tests := []struct {
		name   string
		status database.SaleStatus
	}{
		{name: "open", status: database.SaleStatusOPEN},
		{name: "cancelled", status: database.SaleStatusCANCELLED},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := receipt.NewService(&testutil.FakeStore{
				GetSaleByIDFn: func(_ context.Context, _ tenancy.StoreScope, _ pgtype.UUID) (database.Sale, error) {
					return testutil.SaleFixtureRow(saleID, tc.status), nil
				},
			})

			_, err := svc.Get(context.Background(), testScope, saleID.String())
			if !errors.Is(err, receipt.ErrReceiptNotAvailable) {
				t.Fatalf("expected ErrReceiptNotAvailable, got %v", err)
			}
		})
	}
}

func TestGetReceiptSaleNotFound(t *testing.T) {
	saleID := testutil.UUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8a9")
	svc := receipt.NewService(&testutil.FakeStore{
		GetSaleByIDFn: func(_ context.Context, _ tenancy.StoreScope, _ pgtype.UUID) (database.Sale, error) {
			return database.Sale{}, pgx.ErrNoRows
		},
	})

	_, err := svc.Get(context.Background(), testScope, saleID.String())
	if !errors.Is(err, receipt.ErrSaleNotFound) {
		t.Fatalf("expected ErrSaleNotFound, got %v", err)
	}
}

func TestGetReceiptPropagatesItemAndPaymentErrors(t *testing.T) {
	saleID := testutil.UUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8a9")
	docID := testutil.UUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8ac")
	tests := []struct {
		name  string
		store *testutil.FakeStore
	}{
		{
			name: "items",
			store: &testutil.FakeStore{
				GetSaleByIDFn: func(_ context.Context, _ tenancy.StoreScope, _ pgtype.UUID) (database.Sale, error) {
					return testutil.SaleFixtureRow(saleID, database.SaleStatusCOMPLETED), nil
				},
				ListSaleItemsBySaleIDFn: func(_ context.Context, _ tenancy.StoreScope, _ pgtype.UUID) ([]database.SaleItem, error) {
					return nil, errors.New("items failed")
				},
			},
		},
		{
			name: "payments",
			store: &testutil.FakeStore{
				GetSaleByIDFn: func(_ context.Context, _ tenancy.StoreScope, _ pgtype.UUID) (database.Sale, error) {
					return testutil.SaleFixtureRow(saleID, database.SaleStatusCOMPLETED), nil
				},
				ListSaleItemsBySaleIDFn: func(_ context.Context, _ tenancy.StoreScope, _ pgtype.UUID) ([]database.SaleItem, error) {
					return []database.SaleItem{testutil.SaleItemFixture(testutil.UUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8b0"), saleID, "10.00", "1.000", "0.00", "10.00", "ABC-001")}, nil
				},
				ListReceiptPaymentsBySaleIDFn: func(_ context.Context, _ tenancy.StoreScope, _ pgtype.UUID) ([]database.ReceiptPayment, error) {
					return nil, errors.New("payments failed")
				},
			},
		},
		{
			name: "fiscal",
			store: &testutil.FakeStore{
				GetSaleByIDFn: func(_ context.Context, _ tenancy.StoreScope, _ pgtype.UUID) (database.Sale, error) {
					return testutil.SaleFixtureRow(saleID, database.SaleStatusCOMPLETED), nil
				},
				ListSaleItemsBySaleIDFn: func(_ context.Context, _ tenancy.StoreScope, _ pgtype.UUID) ([]database.SaleItem, error) {
					return []database.SaleItem{testutil.SaleItemFixture(testutil.UUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8b0"), saleID, "10.00", "1.000", "0.00", "10.00", "ABC-001")}, nil
				},
				ListReceiptPaymentsBySaleIDFn: func(_ context.Context, _ tenancy.StoreScope, _ pgtype.UUID) ([]database.ReceiptPayment, error) {
					return []database.ReceiptPayment{}, nil
				},
				GetFiscalDocumentBySaleIDFn: func(_ context.Context, _ tenancy.StoreScope, _ pgtype.UUID) (database.FiscalDocument, error) {
					return database.FiscalDocument{}, errors.New("fiscal failed")
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.store.GetFiscalDocumentBySaleIDFn == nil {
				tc.store.GetFiscalDocumentBySaleIDFn = func(_ context.Context, _ tenancy.StoreScope, _ pgtype.UUID) (database.FiscalDocument, error) {
					return testutil.FiscalDocumentFixture(docID, saleID, database.FiscalDocumentStatusAUTHORIZED), nil
				}
			}
			svc := receipt.NewService(tc.store)
			_, err := svc.Get(context.Background(), testScope, saleID.String())
			if err == nil {
				t.Fatalf("expected error")
			}
		})
	}
}
