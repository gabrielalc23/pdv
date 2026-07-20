package payments_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gabrielalc23/pdv/internal/payments"
	"github.com/gabrielalc23/pdv/internal/platform/authn"
	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/gabrielalc23/pdv/internal/platform/tenancy"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	orgID   = mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b801")
	storeID = mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b802")
	actor   = authn.StoreActor{
		UserID:         mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b803"),
		SessionID:      mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b804"),
		OrganizationID: orgID,
		MembershipID:   mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b805"),
		StoreID:        storeID,
		ClientID:       "test-client",
	}
)

func TestListPaymentMethodsReturnsActiveOnlyAndEmptySlice(t *testing.T) {
	methodID := mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8aa")
	svc := payments.NewService(&paymentsFakeStore{
		listActivePaymentMethodsFn: func(context.Context, tenancy.OrganizationScope) ([]database.PaymentMethod, error) {
			return []database.PaymentMethod{paymentMethodFixture(methodID, "pix", "PIX", database.PaymentMethodKindPIX, true)}, nil
		},
	})

	resp, err := svc.ListPaymentMethods(context.Background(), actor)
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
	svc := payments.NewService(&paymentsFakeStore{
		listActivePaymentMethodsFn: func(context.Context, tenancy.OrganizationScope) ([]database.PaymentMethod, error) {
			return nil, errors.New("db down")
		},
	})

	_, err := svc.ListPaymentMethods(context.Background(), actor)
	if err == nil || err.Error() != "list active payment methods: db down" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListSalePaymentsReturnsOrderedPayments(t *testing.T) {
	saleID := mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8a9")
	pixID := mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8aa")
	cashID := mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8ab")

	svc := payments.NewService(&paymentsFakeStore{
		getSaleByIDFn: func(context.Context, tenancy.StoreScope, pgtype.UUID) (database.Sale, error) {
			return saleFixtureRow(saleID, database.SaleStatusCOMPLETED), nil
		},
		listPaymentsBySaleIDFn: func(context.Context, tenancy.StoreScope, pgtype.UUID) ([]database.Payment, error) {
			return []database.Payment{
				paymentFixture(mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8b1"), saleID, cashID, "40.00", "40.00", "0.00", 1, "cash", 2),
				paymentFixture(mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8b2"), saleID, pixID, "60.00", "60.00", "0.00", 1, "pix", 1),
			}, nil
		},
		getPaymentMethodByIDFn: func(_ context.Context, _ tenancy.OrganizationScope, id pgtype.UUID) (database.PaymentMethod, error) {
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

	resp, err := svc.ListSalePayments(context.Background(), actor, saleID.String())
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
	svc := payments.NewService(&paymentsFakeStore{
		getSaleByIDFn: func(context.Context, tenancy.StoreScope, pgtype.UUID) (database.Sale, error) {
			return saleFixtureRow(saleID, database.SaleStatusOPEN), nil
		},
		listPaymentsBySaleIDFn: func(context.Context, tenancy.StoreScope, pgtype.UUID) ([]database.Payment, error) {
			return []database.Payment{}, nil
		},
	})

	resp, err := svc.ListSalePayments(context.Background(), actor, saleID.String())
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
	svc := payments.NewService(&paymentsFakeStore{
		getSaleByIDFn: func(context.Context, tenancy.StoreScope, pgtype.UUID) (database.Sale, error) {
			return database.Sale{}, pgx.ErrNoRows
		},
	})

	_, err := svc.ListSalePayments(context.Background(), actor, saleID.String())
	if !errors.Is(err, payments.ErrSaleNotFound) {
		t.Fatalf("expected payments.ErrSaleNotFound, got %v", err)
	}
}

func TestListSalePaymentsMethodMissingFails(t *testing.T) {
	saleID := mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8a9")
	methodID := mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8aa")
	svc := payments.NewService(&paymentsFakeStore{
		getSaleByIDFn: func(context.Context, tenancy.StoreScope, pgtype.UUID) (database.Sale, error) {
			return saleFixtureRow(saleID, database.SaleStatusCOMPLETED), nil
		},
		listPaymentsBySaleIDFn: func(context.Context, tenancy.StoreScope, pgtype.UUID) ([]database.Payment, error) {
			return []database.Payment{
				paymentFixture(mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8b3"), saleID, methodID, "10.00", "10.00", "0.00", 1, "pix", 1),
			}, nil
		},
		getPaymentMethodByIDFn: func(context.Context, tenancy.OrganizationScope, pgtype.UUID) (database.PaymentMethod, error) {
			return database.PaymentMethod{}, pgx.ErrNoRows
		},
	})

	_, err := svc.ListSalePayments(context.Background(), actor, saleID.String())
	if err == nil {
		t.Fatalf("expected error")
	}
}

type paymentsFakeStore struct {
	listActivePaymentMethodsFn func(context.Context, tenancy.OrganizationScope) ([]database.PaymentMethod, error)
	listPaymentsBySaleIDFn     func(context.Context, tenancy.StoreScope, pgtype.UUID) ([]database.Payment, error)
	getSaleByIDFn              func(context.Context, tenancy.StoreScope, pgtype.UUID) (database.Sale, error)
	getPaymentMethodByIDFn     func(context.Context, tenancy.OrganizationScope, pgtype.UUID) (database.PaymentMethod, error)
}

func (f *paymentsFakeStore) ListActivePaymentMethods(ctx context.Context, scope tenancy.OrganizationScope) ([]database.PaymentMethod, error) {
	if f.listActivePaymentMethodsFn != nil {
		return f.listActivePaymentMethodsFn(ctx, scope)
	}
	return []database.PaymentMethod{}, nil
}

func (f *paymentsFakeStore) ListPaymentsBySaleID(ctx context.Context, scope tenancy.StoreScope, saleID pgtype.UUID) ([]database.Payment, error) {
	if f.listPaymentsBySaleIDFn != nil {
		return f.listPaymentsBySaleIDFn(ctx, scope, saleID)
	}
	return []database.Payment{}, nil
}

func (f *paymentsFakeStore) GetSaleByID(ctx context.Context, scope tenancy.StoreScope, saleID pgtype.UUID) (database.Sale, error) {
	if f.getSaleByIDFn != nil {
		return f.getSaleByIDFn(ctx, scope, saleID)
	}
	return database.Sale{}, pgx.ErrNoRows
}

func (f *paymentsFakeStore) GetPaymentMethodByID(ctx context.Context, scope tenancy.OrganizationScope, id pgtype.UUID) (database.PaymentMethod, error) {
	if f.getPaymentMethodByIDFn != nil {
		return f.getPaymentMethodByIDFn(ctx, scope, id)
	}
	return database.PaymentMethod{}, pgx.ErrNoRows
}

func saleFixtureRow(id pgtype.UUID, status database.SaleStatus) database.Sale {
	now := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	row := database.Sale{
		ID:             id,
		OrganizationID: orgID,
		StoreID:        storeID,
		Number:         77,
		IdempotencyKey: "sale-1",
		Status:         status,
		Subtotal:       mustNumeric("100.00"),
		Discount:       mustNumeric("0.00"),
		Addition:       mustNumeric("0.00"),
		Total:          mustNumeric("100.00"),
		OpenedAt:       timestamptz(now),
		CreatedAt:      timestamptz(now),
		UpdatedAt:      timestamptz(now),
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

func paymentFixture(rowID, saleID, methodID pgtype.UUID, amount, received, change string, installments int16, externalReference string, order int) database.Payment {
	return database.Payment{
		ID:                rowID,
		OrganizationID:    orgID,
		StoreID:           storeID,
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
