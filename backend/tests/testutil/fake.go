package testutil

import (
	"context"
	"time"

	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/gabrielalc23/pdv/internal/platform/tenancy"
	"github.com/gabrielalc23/pdv/internal/receipt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type FakeStore struct {
	GetSaleByIDFn                 func(context.Context, tenancy.StoreScope, pgtype.UUID) (database.Sale, error)
	ListSaleItemsBySaleIDFn       func(context.Context, tenancy.StoreScope, pgtype.UUID) ([]database.SaleItem, error)
	ListReceiptPaymentsBySaleIDFn func(context.Context, tenancy.StoreScope, pgtype.UUID) ([]database.ReceiptPayment, error)
	GetFiscalDocumentBySaleIDFn   func(context.Context, tenancy.StoreScope, pgtype.UUID) (database.FiscalDocument, error)
}

func (f *FakeStore) GetSaleByID(ctx context.Context, scope tenancy.StoreScope, saleID pgtype.UUID) (database.Sale, error) {
	if f.GetSaleByIDFn != nil {
		return f.GetSaleByIDFn(ctx, scope, saleID)
	}
	return database.Sale{}, pgx.ErrNoRows
}

func (f *FakeStore) ListSaleItemsBySaleID(ctx context.Context, scope tenancy.StoreScope, saleID pgtype.UUID) ([]database.SaleItem, error) {
	if f.ListSaleItemsBySaleIDFn != nil {
		return f.ListSaleItemsBySaleIDFn(ctx, scope, saleID)
	}
	return []database.SaleItem{}, nil
}

func (f *FakeStore) ListReceiptPaymentsBySaleID(ctx context.Context, scope tenancy.StoreScope, saleID pgtype.UUID) ([]database.ReceiptPayment, error) {
	if f.ListReceiptPaymentsBySaleIDFn != nil {
		return f.ListReceiptPaymentsBySaleIDFn(ctx, scope, saleID)
	}
	return []database.ReceiptPayment{}, nil
}

func (f *FakeStore) GetFiscalDocumentBySaleID(ctx context.Context, scope tenancy.StoreScope, saleID pgtype.UUID) (database.FiscalDocument, error) {
	if f.GetFiscalDocumentBySaleIDFn != nil {
		return f.GetFiscalDocumentBySaleIDFn(ctx, scope, saleID)
	}
	return database.FiscalDocument{}, pgx.ErrNoRows
}

var _ receipt.Store = (*FakeStore)(nil)

func SaleFixtureRow(id pgtype.UUID, status database.SaleStatus) database.Sale {
	now := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	row := database.Sale{
		ID:             id,
		Number:         77,
		Status:         status,
		Subtotal:       Numeric("100.00"),
		Discount:       Numeric("0.00"),
		Addition:       Numeric("0.00"),
		Total:          Numeric("100.00"),
		OpenedAt:       Timestamptz(now),
		CreatedAt:      Timestamptz(now),
		UpdatedAt:      Timestamptz(now),
		IdempotencyKey: "sale-1",
	}
	if status == database.SaleStatusCOMPLETED {
		row.CompletedAt = Timestamptz(now.Add(time.Minute))
	}
	if status == database.SaleStatusCANCELLED {
		row.CancelledAt = Timestamptz(now.Add(time.Minute))
	}
	return row
}

func SaleItemFixture(id, saleID pgtype.UUID, unitPrice, quantity, discount, total, sku string) database.SaleItem {
	return database.SaleItem{
		ID:          id,
		SaleID:      saleID,
		ProductID:   UUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8dd"),
		ProductName: "Produto",
		ProductSKU:  sku,
		UnitPrice:   Numeric(unitPrice),
		Quantity:    Numeric(quantity),
		Discount:    Numeric(discount),
		Total:       Numeric(total),
		CreatedAt:   Timestamptz(time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)),
	}
}

func ReceiptPaymentFixture(id, saleID pgtype.UUID, amount, received, change string, installments int16, methodName string, order int) database.ReceiptPayment {
	return database.ReceiptPayment{
		ID:                id,
		SaleID:            saleID,
		PaymentMethodName: methodName,
		Status:            database.PaymentStatusAPPROVED,
		Amount:            Numeric(amount),
		ReceivedAmount:    Numeric(received),
		ChangeAmount:      Numeric(change),
		Installments:      installments,
		ExternalReference: pgtype.Text{String: methodName, Valid: true},
		PaidAt:            Timestamptz(time.Date(2026, 7, 15, 12, 0, order, 0, time.UTC)),
		CreatedAt:         Timestamptz(time.Date(2026, 7, 15, 12, 0, order, 0, time.UTC)),
		UpdatedAt:         Timestamptz(time.Date(2026, 7, 15, 12, 0, order, 0, time.UTC)),
	}
}

func FiscalDocumentFixture(id, saleID pgtype.UUID, status database.FiscalDocumentStatus) database.FiscalDocument {
	now := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	row := database.FiscalDocument{
		ID:            id,
		SaleID:        saleID,
		Status:        status,
		Environment:   database.FiscalEnvironmentHOMOLOGATION,
		DocumentModel: 65,
		CreatedAt:     Timestamptz(now),
		UpdatedAt:     Timestamptz(now),
	}
	if status == database.FiscalDocumentStatusAUTHORIZED {
		row.AccessKey = pgtype.Text{String: "12345678901234567890123456789012345678901234", Valid: true}
		row.Protocol = pgtype.Text{String: "MOCK-77", Valid: true}
		row.Provider = pgtype.Text{String: "mock", Valid: true}
		row.ExternalReference = pgtype.Text{String: "sale-" + saleID.String(), Valid: true}
		row.XML = pgtype.Text{String: "<fiscal />", Valid: true}
		row.IssuedAt = Timestamptz(now.Add(time.Minute))
	}
	return row
}

func UUID(raw string) pgtype.UUID {
	var id pgtype.UUID
	if err := id.Scan(raw); err != nil {
		panic(err)
	}
	return id
}

func Numeric(raw string) pgtype.Numeric {
	var n pgtype.Numeric
	if err := n.Scan(raw); err != nil {
		panic(err)
	}
	return n
}

func Timestamptz(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: value.UTC(), Valid: true}
}
