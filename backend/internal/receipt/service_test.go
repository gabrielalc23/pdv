package receipt

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestGetReceiptReturnsCompletedSaleWithCalculatedSubtotal(t *testing.T) {
	saleID := mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8a9")
	pixID := mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8aa")
	cashID := mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8ab")
	docID := mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8ac")
	svc := NewService(&receiptFakeStore{
		getSaleByIDFn: func(context.Context, pgtype.UUID) (database.GetSaleByIDRow, error) {
			return saleFixtureRow(saleID, database.SaleStatusCOMPLETED), nil
		},
		listSaleItemsBySaleIDFn: func(context.Context, pgtype.UUID) ([]database.SaleItem, error) {
			return []database.SaleItem{
				receiptItemFixture(mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8b0"), saleID, "12.34", "2.000", "4.68", "20.00", "ABC-001"),
				receiptItemFixture(mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8b1"), saleID, "5.00", "1.000", "0.00", "5.00", "XYZ-002"),
			}, nil
		},
		listPaymentsBySaleIDFn: func(context.Context, pgtype.UUID) ([]database.ListPaymentsBySaleIDRow, error) {
			return []database.ListPaymentsBySaleIDRow{
				receiptPaymentRowFixture(mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8c0"), saleID, cashID, "40.00", "40.00", "0.00", 1, "cash", 2),
				receiptPaymentRowFixture(mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8c1"), saleID, pixID, "60.00", "60.00", "0.00", 1, "pix", 1),
			}, nil
		},
		getPaymentMethodByIDFn: func(_ context.Context, id pgtype.UUID) (database.PaymentMethod, error) {
			switch id.String() {
			case pixID.String():
				return receiptPaymentMethodFixture(pixID, "pix", "PIX", database.PaymentMethodKindPIX), nil
			case cashID.String():
				return receiptPaymentMethodFixture(cashID, "cash", "Cash", database.PaymentMethodKindCASH), nil
			default:
				return database.PaymentMethod{}, pgx.ErrNoRows
			}
		},
		getFiscalDocumentBySaleIDFn: func(context.Context, pgtype.UUID) (database.FiscalDocument, error) {
			return receiptFiscalDocumentFixture(docID, saleID, database.FiscalDocumentStatusAUTHORIZED), nil
		},
	})

	resp, err := svc.Get(context.Background(), saleID.String())
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
	saleID := mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8a9")
	tests := []struct {
		name   string
		status database.SaleStatus
	}{
		{name: "open", status: database.SaleStatusOPEN},
		{name: "cancelled", status: database.SaleStatusCANCELLED},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := NewService(&receiptFakeStore{
				getSaleByIDFn: func(context.Context, pgtype.UUID) (database.GetSaleByIDRow, error) {
					return saleFixtureRow(saleID, tc.status), nil
				},
			})

			_, err := svc.Get(context.Background(), saleID.String())
			if !errors.Is(err, ErrReceiptNotAvailable) {
				t.Fatalf("expected ErrReceiptNotAvailable, got %v", err)
			}
		})
	}
}

func TestGetReceiptSaleNotFound(t *testing.T) {
	saleID := mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8a9")
	svc := NewService(&receiptFakeStore{
		getSaleByIDFn: func(context.Context, pgtype.UUID) (database.GetSaleByIDRow, error) {
			return database.GetSaleByIDRow{}, pgx.ErrNoRows
		},
	})

	_, err := svc.Get(context.Background(), saleID.String())
	if !errors.Is(err, ErrSaleNotFound) {
		t.Fatalf("expected ErrSaleNotFound, got %v", err)
	}
}

func TestGetReceiptPropagatesItemAndPaymentErrors(t *testing.T) {
	saleID := mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8a9")
	docID := mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8ac")
	tests := []struct {
		name  string
		store *receiptFakeStore
	}{
		{
			name: "items",
			store: &receiptFakeStore{
				getSaleByIDFn: func(context.Context, pgtype.UUID) (database.GetSaleByIDRow, error) {
					return saleFixtureRow(saleID, database.SaleStatusCOMPLETED), nil
				},
				listSaleItemsBySaleIDFn: func(context.Context, pgtype.UUID) ([]database.SaleItem, error) {
					return nil, errors.New("items failed")
				},
			},
		},
		{
			name: "payments",
			store: &receiptFakeStore{
				getSaleByIDFn: func(context.Context, pgtype.UUID) (database.GetSaleByIDRow, error) {
					return saleFixtureRow(saleID, database.SaleStatusCOMPLETED), nil
				},
				listSaleItemsBySaleIDFn: func(context.Context, pgtype.UUID) ([]database.SaleItem, error) {
					return []database.SaleItem{receiptItemFixture(mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8b0"), saleID, "10.00", "1.000", "0.00", "10.00", "ABC-001")}, nil
				},
				listPaymentsBySaleIDFn: func(context.Context, pgtype.UUID) ([]database.ListPaymentsBySaleIDRow, error) {
					return nil, errors.New("payments failed")
				},
			},
		},
		{
			name: "fiscal",
			store: &receiptFakeStore{
				getSaleByIDFn: func(context.Context, pgtype.UUID) (database.GetSaleByIDRow, error) {
					return saleFixtureRow(saleID, database.SaleStatusCOMPLETED), nil
				},
				listSaleItemsBySaleIDFn: func(context.Context, pgtype.UUID) ([]database.SaleItem, error) {
					return []database.SaleItem{receiptItemFixture(mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8b0"), saleID, "10.00", "1.000", "0.00", "10.00", "ABC-001")}, nil
				},
				listPaymentsBySaleIDFn: func(context.Context, pgtype.UUID) ([]database.ListPaymentsBySaleIDRow, error) {
					return []database.ListPaymentsBySaleIDRow{}, nil
				},
				getFiscalDocumentBySaleIDFn: func(context.Context, pgtype.UUID) (database.FiscalDocument, error) {
					return database.FiscalDocument{}, errors.New("fiscal failed")
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.store.getFiscalDocumentBySaleIDFn == nil {
				tc.store.getFiscalDocumentBySaleIDFn = func(context.Context, pgtype.UUID) (database.FiscalDocument, error) {
					return receiptFiscalDocumentFixture(docID, saleID, database.FiscalDocumentStatusAUTHORIZED), nil
				}
			}
			svc := NewService(tc.store)
			_, err := svc.Get(context.Background(), saleID.String())
			if err == nil {
				t.Fatalf("expected error")
			}
		})
	}
}

type receiptFakeStore struct {
	getSaleByIDFn               func(context.Context, pgtype.UUID) (database.GetSaleByIDRow, error)
	listSaleItemsBySaleIDFn     func(context.Context, pgtype.UUID) ([]database.SaleItem, error)
	listPaymentsBySaleIDFn      func(context.Context, pgtype.UUID) ([]database.ListPaymentsBySaleIDRow, error)
	getPaymentMethodByIDFn      func(context.Context, pgtype.UUID) (database.PaymentMethod, error)
	getFiscalDocumentBySaleIDFn func(context.Context, pgtype.UUID) (database.FiscalDocument, error)
}

func (f *receiptFakeStore) GetSaleByID(ctx context.Context, saleID pgtype.UUID) (database.GetSaleByIDRow, error) {
	if f.getSaleByIDFn != nil {
		return f.getSaleByIDFn(ctx, saleID)
	}
	return database.GetSaleByIDRow{}, pgx.ErrNoRows
}

func (f *receiptFakeStore) ListSaleItemsBySaleID(ctx context.Context, saleID pgtype.UUID) ([]database.SaleItem, error) {
	if f.listSaleItemsBySaleIDFn != nil {
		return f.listSaleItemsBySaleIDFn(ctx, saleID)
	}
	return []database.SaleItem{}, nil
}

func (f *receiptFakeStore) ListPaymentsBySaleID(ctx context.Context, saleID pgtype.UUID) ([]database.ListPaymentsBySaleIDRow, error) {
	if f.listPaymentsBySaleIDFn != nil {
		return f.listPaymentsBySaleIDFn(ctx, saleID)
	}
	return []database.ListPaymentsBySaleIDRow{}, nil
}

func (f *receiptFakeStore) GetPaymentMethodByID(ctx context.Context, id pgtype.UUID) (database.PaymentMethod, error) {
	if f.getPaymentMethodByIDFn != nil {
		return f.getPaymentMethodByIDFn(ctx, id)
	}
	return database.PaymentMethod{}, pgx.ErrNoRows
}

func (f *receiptFakeStore) GetFiscalDocumentBySaleID(ctx context.Context, saleID pgtype.UUID) (database.FiscalDocument, error) {
	if f.getFiscalDocumentBySaleIDFn != nil {
		return f.getFiscalDocumentBySaleIDFn(ctx, saleID)
	}
	return database.FiscalDocument{}, pgx.ErrNoRows
}

func saleFixtureRow(id pgtype.UUID, status database.SaleStatus) database.GetSaleByIDRow {
	now := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	row := database.GetSaleByIDRow{
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

func receiptItemFixture(id, saleID pgtype.UUID, unitPrice, quantity, discount, total, sku string) database.SaleItem {
	return database.SaleItem{
		ID:          id,
		SaleID:      saleID,
		ProductID:   mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8dd"),
		ProductName: "Produto",
		ProductSKU:  sku,
		UnitPrice:   mustNumeric(unitPrice),
		Quantity:    mustNumeric(quantity),
		Discount:    mustNumeric(discount),
		Total:       mustNumeric(total),
		CreatedAt:   timestamptz(time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)),
	}
}

func receiptPaymentRowFixture(id, saleID, methodID pgtype.UUID, amount, received, change string, installments int16, externalReference string, order int) database.ListPaymentsBySaleIDRow {
	return database.ListPaymentsBySaleIDRow{
		ID:                id,
		SaleID:            saleID,
		PaymentMethodID:   methodID,
		Status:            database.PaymentStatusAPPROVED,
		Amount:            mustNumeric(amount),
		ReceivedAmount:    mustNumeric(received),
		ChangeAmount:      mustNumeric(change),
		Installments:      installments,
		ExternalReference: pgtype.Text{String: externalReference, Valid: true},
		PaidAt:            timestamptz(time.Date(2026, 7, 15, 12, 0, order, 0, time.UTC)),
		CreatedAt:         timestamptz(time.Date(2026, 7, 15, 12, 0, order, 0, time.UTC)),
		UpdatedAt:         timestamptz(time.Date(2026, 7, 15, 12, 0, order, 0, time.UTC)),
		IdempotencyKey:    "payment-" + externalReference,
	}
}

func receiptPaymentMethodFixture(id pgtype.UUID, code, name string, kind database.PaymentMethodKind) database.PaymentMethod {
	return database.PaymentMethod{
		ID:            id,
		Code:          code,
		Name:          name,
		Kind:          kind,
		IsActive:      true,
		FeePercentage: mustNumeric("0.00"),
		CreatedAt:     timestamptz(time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)),
		UpdatedAt:     timestamptz(time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)),
	}
}

func receiptFiscalDocumentFixture(id, saleID pgtype.UUID, status database.FiscalDocumentStatus) database.FiscalDocument {
	now := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	row := database.FiscalDocument{
		ID:            id,
		SaleID:        saleID,
		Status:        status,
		Environment:   database.FiscalEnvironmentHOMOLOGATION,
		DocumentModel: 65,
		CreatedAt:     timestamptz(now),
		UpdatedAt:     timestamptz(now),
	}
	if status == database.FiscalDocumentStatusAUTHORIZED {
		row.AccessKey = pgtype.Text{String: "12345678901234567890123456789012345678901234", Valid: true}
		row.Protocol = pgtype.Text{String: "MOCK-77", Valid: true}
		row.Provider = pgtype.Text{String: "mock", Valid: true}
		row.ExternalReference = pgtype.Text{String: "sale-" + saleID.String(), Valid: true}
		row.XML = pgtype.Text{String: "<fiscal />", Valid: true}
		row.IssuedAt = timestamptz(now.Add(time.Minute))
	}
	return row
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
