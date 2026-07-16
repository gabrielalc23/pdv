package payments

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestListPaymentMethodsReturnsActiveOnlyAndEmptySlice(t *testing.T) {
	methodID := mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8aa")
	svc := NewService(&paymentsFakeStore{
		listActivePaymentMethodsFn: func(context.Context) ([]database.PaymentMethod, error) {
			return []database.PaymentMethod{paymentMethodFixture(methodID, "pix", "PIX", database.PaymentMethodKindPIX, true)}, nil
		},
	})

	resp, err := svc.ListPaymentMethods(context.Background())
	if err != nil {
		t.Fatalf("ListPaymentMethods returned error: %v", err)
	}
	if resp.Data == nil || len(resp.Data) != 1 {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if !resp.Data[0].IsActive {
		t.Fatalf("expected active payment method: %+v", resp.Data[0])
	}
}

func TestListPaymentMethodsPropagatesError(t *testing.T) {
	svc := NewService(&paymentsFakeStore{
		listActivePaymentMethodsFn: func(context.Context) ([]database.PaymentMethod, error) {
			return nil, errors.New("db down")
		},
	})

	_, err := svc.ListPaymentMethods(context.Background())
	if err == nil || err.Error() != "list active payment methods: db down" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListSalePaymentsReturnsOrderedPayments(t *testing.T) {
	saleID := mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8a9")
	pixID := mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8aa")
	cashID := mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8ab")

	svc := NewService(&paymentsFakeStore{
		getSaleByIDFn: func(context.Context, pgtype.UUID) (database.GetSaleByIDRow, error) {
			return saleFixtureRow(saleID, database.SaleStatusCOMPLETED), nil
		},
		listPaymentsBySaleIDFn: func(context.Context, pgtype.UUID) ([]database.ListPaymentsBySaleIDRow, error) {
			return []database.ListPaymentsBySaleIDRow{
				paymentListRowFixture(mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8b1"), saleID, cashID, "40.00", "40.00", "0.00", 1, "cash", 2),
				paymentListRowFixture(mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8b2"), saleID, pixID, "60.00", "60.00", "0.00", 1, "pix", 1),
			}, nil
		},
		getPaymentMethodByIDFn: func(_ context.Context, id pgtype.UUID) (database.PaymentMethod, error) {
			switch id.String() {
			case pixID.String():
				return paymentMethodFixture(pixID, "pix", "PIX", database.PaymentMethodKindPIX, true), nil
			case cashID.String():
				return paymentMethodFixture(cashID, "cash", "Cash", database.PaymentMethodKindCASH, true), nil
			default:
				return database.PaymentMethod{}, pgx.ErrNoRows
			}
		},
	})

	resp, err := svc.ListSalePayments(context.Background(), saleID.String())
	if err != nil {
		t.Fatalf("ListSalePayments returned error: %v", err)
	}
	if resp.Data == nil || len(resp.Data) != 2 {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if resp.Data[0].PaymentMethodCode != "cash" || resp.Data[1].PaymentMethodCode != "pix" {
		t.Fatalf("unexpected ordering: %+v", resp.Data)
	}
	if resp.Data[0].Amount != "40.00" || resp.Data[1].Amount != "60.00" {
		t.Fatalf("unexpected amounts: %+v", resp.Data)
	}
}

func TestListSalePaymentsReturnsEmptySliceWhenNone(t *testing.T) {
	saleID := mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8a9")
	svc := NewService(&paymentsFakeStore{
		getSaleByIDFn: func(context.Context, pgtype.UUID) (database.GetSaleByIDRow, error) {
			return saleFixtureRow(saleID, database.SaleStatusOPEN), nil
		},
		listPaymentsBySaleIDFn: func(context.Context, pgtype.UUID) ([]database.ListPaymentsBySaleIDRow, error) {
			return []database.ListPaymentsBySaleIDRow{}, nil
		},
	})

	resp, err := svc.ListSalePayments(context.Background(), saleID.String())
	if err != nil {
		t.Fatalf("ListSalePayments returned error: %v", err)
	}
	if resp.Data == nil {
		t.Fatalf("expected empty slice, got nil")
	}
	if len(resp.Data) != 0 {
		t.Fatalf("expected empty slice, got %+v", resp.Data)
	}
}

func TestListSalePaymentsSaleNotFound(t *testing.T) {
	saleID := mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8a9")
	svc := NewService(&paymentsFakeStore{
		getSaleByIDFn: func(context.Context, pgtype.UUID) (database.GetSaleByIDRow, error) {
			return database.GetSaleByIDRow{}, pgx.ErrNoRows
		},
	})

	_, err := svc.ListSalePayments(context.Background(), saleID.String())
	if !errors.Is(err, ErrSaleNotFound) {
		t.Fatalf("expected ErrSaleNotFound, got %v", err)
	}
}

func TestListSalePaymentsMethodMissingFails(t *testing.T) {
	saleID := mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8a9")
	methodID := mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8aa")
	svc := NewService(&paymentsFakeStore{
		getSaleByIDFn: func(context.Context, pgtype.UUID) (database.GetSaleByIDRow, error) {
			return saleFixtureRow(saleID, database.SaleStatusCOMPLETED), nil
		},
		listPaymentsBySaleIDFn: func(context.Context, pgtype.UUID) ([]database.ListPaymentsBySaleIDRow, error) {
			return []database.ListPaymentsBySaleIDRow{
				paymentListRowFixture(mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8b3"), saleID, methodID, "10.00", "10.00", "0.00", 1, "pix", 1),
			}, nil
		},
		getPaymentMethodByIDFn: func(context.Context, pgtype.UUID) (database.PaymentMethod, error) {
			return database.PaymentMethod{}, pgx.ErrNoRows
		},
	})

	_, err := svc.ListSalePayments(context.Background(), saleID.String())
	if err == nil {
		t.Fatalf("expected error")
	}
}

type paymentsFakeStore struct {
	listActivePaymentMethodsFn func(context.Context) ([]database.PaymentMethod, error)
	listPaymentsBySaleIDFn     func(context.Context, pgtype.UUID) ([]database.ListPaymentsBySaleIDRow, error)
	getSaleByIDFn              func(context.Context, pgtype.UUID) (database.GetSaleByIDRow, error)
	getPaymentMethodByIDFn     func(context.Context, pgtype.UUID) (database.PaymentMethod, error)
}

func (f *paymentsFakeStore) ListActivePaymentMethods(ctx context.Context) ([]database.PaymentMethod, error) {
	if f.listActivePaymentMethodsFn != nil {
		return f.listActivePaymentMethodsFn(ctx)
	}
	return []database.PaymentMethod{}, nil
}

func (f *paymentsFakeStore) ListPaymentsBySaleID(ctx context.Context, saleID pgtype.UUID) ([]database.ListPaymentsBySaleIDRow, error) {
	if f.listPaymentsBySaleIDFn != nil {
		return f.listPaymentsBySaleIDFn(ctx, saleID)
	}
	return []database.ListPaymentsBySaleIDRow{}, nil
}

func (f *paymentsFakeStore) GetSaleByID(ctx context.Context, saleID pgtype.UUID) (database.GetSaleByIDRow, error) {
	if f.getSaleByIDFn != nil {
		return f.getSaleByIDFn(ctx, saleID)
	}
	return database.GetSaleByIDRow{}, pgx.ErrNoRows
}

func (f *paymentsFakeStore) GetPaymentMethodByID(ctx context.Context, id pgtype.UUID) (database.PaymentMethod, error) {
	if f.getPaymentMethodByIDFn != nil {
		return f.getPaymentMethodByIDFn(ctx, id)
	}
	return database.PaymentMethod{}, pgx.ErrNoRows
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

func paymentMethodFixture(id pgtype.UUID, code, name string, kind database.PaymentMethodKind, active bool) database.PaymentMethod {
	return database.PaymentMethod{
		ID:            id,
		Code:          code,
		Name:          name,
		Kind:          kind,
		IsActive:      active,
		FeePercentage: mustNumeric("0.00"),
		CreatedAt:     timestamptz(time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)),
		UpdatedAt:     timestamptz(time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)),
	}
}

func paymentListRowFixture(rowID, saleID, methodID pgtype.UUID, amount, received, change string, installments int16, externalReference string, order int) database.ListPaymentsBySaleIDRow {
	return database.ListPaymentsBySaleIDRow{
		ID:                rowID,
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
