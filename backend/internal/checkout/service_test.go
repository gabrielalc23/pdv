package checkout_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gabrielalc23/pdv/internal/checkout"
	"github.com/gabrielalc23/pdv/internal/fiscal"
	"github.com/gabrielalc23/pdv/internal/platform/authn"
	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/gabrielalc23/pdv/internal/platform/tenancy"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

var testPaymentMethodID = mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b899")

func checkoutInputWithTotal(total string) checkout.CheckoutInput {
	return checkout.CheckoutInput{
		Payments: []checkout.CheckoutPaymentInput{
			{PaymentMethodID: testPaymentMethodID.String(), Amount: total},
		},
	}
}

var testScope = authn.StoreActor{
	OrganizationID: mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8a1"),
	StoreID:        mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8a2"),
	MembershipID:   mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8a3"),
	UserID:         mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8a4"),
	SessionID:      mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8a5"),
	ClientID:       "test-client",
}

func TestCheckoutCompletesSaleWithInventoryAndFiscalSuccess(t *testing.T) {
	sale := newCheckoutSaleFixture()
	productID := mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8aa")
	documentID := mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8ac")
	now := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)

	tx := &checkoutFakeTxQueries{
		lockSaleByIDFn: func(_ context.Context, _ tenancy.StoreScope, _ pgtype.UUID) (database.Sale, error) {
			return sale.lock(database.SaleStatusOPEN), nil
		},
		listSaleItemsBySaleIDFn: func(_ context.Context, _ tenancy.StoreScope, _ pgtype.UUID) ([]database.SaleItem, error) {
			return []database.SaleItem{checkoutItemFixture(productID, "50.00", "2.000", "0.00", "100.00")}, nil
		},
		getPaymentMethodByIDFn: func(_ context.Context, _ tenancy.OrganizationScope, _ pgtype.UUID) (database.PaymentMethod, error) {
			return paymentMethodFixture(testPaymentMethodID, "CASH", "Dinheiro", database.PaymentMethodKindCASH, true, false, false, 1, false), nil
		},
		decreaseInventoryFn: func(_ context.Context, _ tenancy.StoreScope, arg database.DecreaseInventoryForStoreParams) (database.DecreaseInventoryForStoreRow, error) {
			if arg.ProductID.String() != productID.String() {
				t.Fatalf("unexpected product id: %s", arg.ProductID.String())
			}
			if got := quantityString(arg.Quantity); got != "2.000" {
				t.Fatalf("unexpected decrement quantity: %s", got)
			}
			return decreaseInventoryRowFixture(productID, "10.000", "8.000"), nil
		},
		createInventoryMovementFn: func(_ context.Context, _ tenancy.ActorScope, arg database.CreateInventoryMovementForStoreParams) (database.CreateInventoryMovementForStoreRow, error) {
			if arg.MovementType != database.InventoryMovementTypeSALE {
				t.Fatalf("unexpected movement type: %s", arg.MovementType)
			}
			if arg.ReferenceType != "sale" {
				t.Fatalf("unexpected reference type: %s", arg.ReferenceType)
			}
			if arg.ReferenceID.String() != sale.id.String() {
				t.Fatalf("unexpected reference id: %s", arg.ReferenceID.String())
			}
			return inventoryMovementFixture(productID, database.InventoryMovementTypeSALE, "2.000", "10.000", "8.000"), nil
		},
		createPaymentFn: func(_ context.Context, _ tenancy.ActorScope, _ database.CreatePaymentForStoreParams) (database.CreatePaymentForStoreRow, error) {
			return createPaymentRowFixture(sale.id, testPaymentMethodID, "100.00", "100.00", "0.00", 1, "checkout:"+sale.id.String()+":payment:0"), nil
		},
		approvePaymentFn: func(_ context.Context, _ tenancy.StoreScope, _ database.ApprovePaymentForStoreParams) (database.Payment, error) {
			return approvePaymentRowFixture(sale.id, testPaymentMethodID, "100.00", "100.00", "0.00", 1, database.PaymentStatusAPPROVED), nil
		},
		completeSaleFn: func(_ context.Context, _ tenancy.ActorScope, _ database.CompleteSaleForStoreParams) (database.Sale, error) {
			return sale.complete(database.SaleStatusCOMPLETED), nil
		},
		createFiscalDocumentFn: func(_ context.Context, _ tenancy.StoreScope, _ database.CreateFiscalDocumentForStoreParams) (database.CreateFiscalDocumentForStoreRow, error) {
			return fiscalDocumentCreateFixture(sale.id, documentID), nil
		},
		markFiscalDocumentAuthorizedFn: func(_ context.Context, _ tenancy.StoreScope, arg database.MarkFiscalDocumentAuthorizedForStoreParams) (database.FiscalDocument, error) {
			if arg.ID.String() != documentID.String() {
				t.Fatalf("unexpected fiscal document id: %s", arg.ID.String())
			}
			return fiscalDocumentFixture(documentID, sale.id, database.FiscalDocumentStatusAUTHORIZED), nil
		},
	}

	var providerCallCount int
	provider := &fiscal.MockProvider{
		Now: func() time.Time {
			providerCallCount++
			return now
		},
	}

	svc := checkout.NewService(&checkoutFakeTxManager{tx: tx}, provider)

	resp, err := svc.Checkout(context.Background(), testScope, sale.id.String(), checkoutInputWithTotal("100.00"))
	if err != nil {
		t.Fatalf("Checkout returned error: %v", err)
	}

	if providerCallCount != 1 {
		t.Fatalf("expected provider call once, got %d", providerCallCount)
	}
	if resp.Sale.Status != string(database.SaleStatusCOMPLETED) {
		t.Fatalf("unexpected sale status: %+v", resp.Sale)
	}
	if resp.FiscalDocument.Status != string(database.FiscalDocumentStatusAUTHORIZED) {
		t.Fatalf("unexpected fiscal document: %+v", resp.FiscalDocument)
	}
	if resp.FiscalDocument.AccessKey == nil || *resp.FiscalDocument.AccessKey == "" {
		t.Fatalf("expected access key: %+v", resp.FiscalDocument)
	}
}

func TestCheckoutAggregatesRepeatedProductsAndProcessesInventoryDeterministically(t *testing.T) {
	sale := newCheckoutSaleFixture()
	sale.subtotal = mustNumeric("35.00")
	sale.total = mustNumeric("35.00")
	productA := mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8aa")
	productB := mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8ab")

	calls := make([]string, 0, 8)
	tx := &checkoutFakeTxQueries{
		lockSaleByIDFn: func(_ context.Context, _ tenancy.StoreScope, _ pgtype.UUID) (database.Sale, error) {
			return sale.lock(database.SaleStatusOPEN), nil
		},
		listSaleItemsBySaleIDFn: func(_ context.Context, _ tenancy.StoreScope, _ pgtype.UUID) ([]database.SaleItem, error) {
			return []database.SaleItem{
				checkoutItemWithPriceFixture(productB, "5.00", "1.000", "0.00", "5.00"),
				checkoutItemWithPriceFixture(productA, "10.00", "2.000", "0.00", "20.00"),
				checkoutItemWithPriceFixture(productB, "5.00", "2.000", "0.00", "10.00"),
			}, nil
		},
		getPaymentMethodByIDFn: func(_ context.Context, _ tenancy.OrganizationScope, _ pgtype.UUID) (database.PaymentMethod, error) {
			return paymentMethodFixture(testPaymentMethodID, "CASH", "Dinheiro", database.PaymentMethodKindCASH, true, false, false, 1, false), nil
		},
		decreaseInventoryFn: func(_ context.Context, _ tenancy.StoreScope, arg database.DecreaseInventoryForStoreParams) (database.DecreaseInventoryForStoreRow, error) {
			calls = append(calls, "decrease:"+arg.ProductID.String()+":"+quantityString(arg.Quantity))
			switch arg.ProductID.String() {
			case productA.String():
				return decreaseInventoryRowFixture(productA, "12.000", "10.000"), nil
			case productB.String():
				return decreaseInventoryRowFixture(productB, "20.000", "17.000"), nil
			default:
				t.Fatalf("unexpected product id: %s", arg.ProductID.String())
				return database.DecreaseInventoryForStoreRow{}, nil
			}
		},
		createInventoryMovementFn: func(_ context.Context, _ tenancy.ActorScope, arg database.CreateInventoryMovementForStoreParams) (database.CreateInventoryMovementForStoreRow, error) {
			calls = append(calls, "movement:"+arg.ProductID.String()+":"+quantityString(arg.Quantity))
			return inventoryMovementFixture(arg.ProductID, arg.MovementType, quantityString(arg.Quantity), quantityString(arg.PreviousQuantity), quantityString(arg.CurrentQuantity)), nil
		},
		createPaymentFn: func(_ context.Context, _ tenancy.ActorScope, _ database.CreatePaymentForStoreParams) (database.CreatePaymentForStoreRow, error) {
			return createPaymentRowFixture(sale.id, testPaymentMethodID, "35.00", "35.00", "0.00", 1, "checkout:"+sale.id.String()+":payment:0"), nil
		},
		approvePaymentFn: func(_ context.Context, _ tenancy.StoreScope, _ database.ApprovePaymentForStoreParams) (database.Payment, error) {
			return approvePaymentRowFixture(sale.id, testPaymentMethodID, "35.00", "35.00", "0.00", 1, database.PaymentStatusAPPROVED), nil
		},
		completeSaleFn: func(_ context.Context, _ tenancy.ActorScope, _ database.CompleteSaleForStoreParams) (database.Sale, error) {
			return sale.complete(database.SaleStatusCOMPLETED), nil
		},
		createFiscalDocumentFn: func(_ context.Context, _ tenancy.StoreScope, _ database.CreateFiscalDocumentForStoreParams) (database.CreateFiscalDocumentForStoreRow, error) {
			return fiscalDocumentCreateFixture(sale.id, mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8ad")), nil
		},
		markFiscalDocumentAuthorizedFn: func(_ context.Context, _ tenancy.StoreScope, _ database.MarkFiscalDocumentAuthorizedForStoreParams) (database.FiscalDocument, error) {
			return fiscalDocumentFixture(mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8ad"), sale.id, database.FiscalDocumentStatusAUTHORIZED), nil
		},
	}

	svc := checkout.NewService(&checkoutFakeTxManager{tx: tx}, &fiscal.MockProvider{Now: func() time.Time { return time.Unix(1, 0) }})

	_, err := svc.Checkout(context.Background(), testScope, sale.id.String(), checkoutInputWithTotal("35.00"))
	if err != nil {
		t.Fatalf("Checkout returned error: %v", err)
	}

	if len(calls) != 4 {
		t.Fatalf("unexpected call count: %#v", calls)
	}
	if calls[0] != "decrease:"+productA.String()+":2.000" {
		t.Fatalf("unexpected first inventory call: %#v", calls)
	}
	if calls[1] != "movement:"+productA.String()+":2.000" {
		t.Fatalf("unexpected first movement call: %#v", calls)
	}
	if calls[2] != "decrease:"+productB.String()+":3.000" {
		t.Fatalf("unexpected second inventory call: %#v", calls)
	}
	if calls[3] != "movement:"+productB.String()+":3.000" {
		t.Fatalf("unexpected second movement call: %#v", calls)
	}
}

func TestCheckoutRollsBackWhenSecondProductFails(t *testing.T) {
	sale := newCheckoutSaleFixture()
	sale.subtotal = mustNumeric("60.00")
	sale.total = mustNumeric("60.00")
	productA := mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8aa")
	productB := mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8ab")
	tx := &checkoutFakeTxQueries{
		lockSaleByIDFn: func(_ context.Context, _ tenancy.StoreScope, _ pgtype.UUID) (database.Sale, error) {
			return sale.lock(database.SaleStatusOPEN), nil
		},
		listSaleItemsBySaleIDFn: func(_ context.Context, _ tenancy.StoreScope, _ pgtype.UUID) ([]database.SaleItem, error) {
			return []database.SaleItem{
				checkoutItemFixture(productA, "10.00", "3.000", "0.00", "30.00"),
				checkoutItemFixture(productB, "10.00", "3.000", "0.00", "30.00"),
			}, nil
		},
		getPaymentMethodByIDFn: func(_ context.Context, _ tenancy.OrganizationScope, _ pgtype.UUID) (database.PaymentMethod, error) {
			return paymentMethodFixture(testPaymentMethodID, "CASH", "Dinheiro", database.PaymentMethodKindCASH, true, false, false, 1, false), nil
		},
		getInventoryByProductIDFn: func(_ context.Context, _ tenancy.StoreScope, id pgtype.UUID) (database.Inventory, error) {
			switch id.String() {
			case productB.String():
				return inventoryFixture(productB, "1.000"), nil
			default:
				return database.Inventory{}, pgx.ErrNoRows
			}
		},
		decreaseInventoryFn: func(_ context.Context, _ tenancy.StoreScope, arg database.DecreaseInventoryForStoreParams) (database.DecreaseInventoryForStoreRow, error) {
			if arg.ProductID.String() == productA.String() {
				return decreaseInventoryRowFixture(productA, "10.000", "7.000"), nil
			}
			return database.DecreaseInventoryForStoreRow{}, pgx.ErrNoRows
		},
		createInventoryMovementFn: func(_ context.Context, _ tenancy.ActorScope, _ database.CreateInventoryMovementForStoreParams) (database.CreateInventoryMovementForStoreRow, error) {
			return inventoryMovementFixture(productA, database.InventoryMovementTypeSALE, "3.000", "10.000", "7.000"), nil
		},
	}

	txManager := &checkoutFakeTxManager{tx: tx}
	svc := checkout.NewService(txManager, nil)

	_, err := svc.Checkout(context.Background(), testScope, sale.id.String(), checkoutInputWithTotal("60.00"))
	if !errors.Is(err, checkout.ErrInsufficientInventory) {
		t.Fatalf("expected checkout.ErrInsufficientInventory, got %v", err)
	}
	if !txManager.rolledBack || txManager.committed {
		t.Fatalf("expected rollback without commit, got committed=%v rolledBack=%v", txManager.committed, txManager.rolledBack)
	}
}

func TestCheckoutBlocksSecondAttemptAfterCompletion(t *testing.T) {
	sale := newCheckoutSaleFixture()
	callCount := 0
	tx := &checkoutFakeTxQueries{
		lockSaleByIDFn: func(_ context.Context, _ tenancy.StoreScope, _ pgtype.UUID) (database.Sale, error) {
			callCount++
			if callCount == 1 {
				return sale.lock(database.SaleStatusOPEN), nil
			}
			return sale.lock(database.SaleStatusCOMPLETED), nil
		},
		listSaleItemsBySaleIDFn: func(_ context.Context, _ tenancy.StoreScope, _ pgtype.UUID) ([]database.SaleItem, error) {
			return []database.SaleItem{checkoutItemFixture(mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8ab"), "100.00", "1.000", "0.00", "100.00")}, nil
		},
		getPaymentMethodByIDFn: func(_ context.Context, _ tenancy.OrganizationScope, _ pgtype.UUID) (database.PaymentMethod, error) {
			return paymentMethodFixture(testPaymentMethodID, "CASH", "Dinheiro", database.PaymentMethodKindCASH, true, false, false, 1, false), nil
		},
		decreaseInventoryFn: func(_ context.Context, _ tenancy.StoreScope, _ database.DecreaseInventoryForStoreParams) (database.DecreaseInventoryForStoreRow, error) {
			return decreaseInventoryRowFixture(mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8ab"), "10.000", "9.000"), nil
		},
		createInventoryMovementFn: func(_ context.Context, _ tenancy.ActorScope, _ database.CreateInventoryMovementForStoreParams) (database.CreateInventoryMovementForStoreRow, error) {
			return inventoryMovementFixture(mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8ab"), database.InventoryMovementTypeSALE, "1.000", "10.000", "9.000"), nil
		},
		createPaymentFn: func(_ context.Context, _ tenancy.ActorScope, _ database.CreatePaymentForStoreParams) (database.CreatePaymentForStoreRow, error) {
			return createPaymentRowFixture(sale.id, testPaymentMethodID, "100.00", "100.00", "0.00", 1, "checkout:"+sale.id.String()+":payment:0"), nil
		},
		approvePaymentFn: func(_ context.Context, _ tenancy.StoreScope, _ database.ApprovePaymentForStoreParams) (database.Payment, error) {
			return approvePaymentRowFixture(sale.id, testPaymentMethodID, "100.00", "100.00", "0.00", 1, database.PaymentStatusAPPROVED), nil
		},
		completeSaleFn: func(_ context.Context, _ tenancy.ActorScope, _ database.CompleteSaleForStoreParams) (database.Sale, error) {
			return sale.complete(database.SaleStatusCOMPLETED), nil
		},
		createFiscalDocumentFn: func(_ context.Context, _ tenancy.StoreScope, _ database.CreateFiscalDocumentForStoreParams) (database.CreateFiscalDocumentForStoreRow, error) {
			return fiscalDocumentCreateFixture(sale.id, mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8ac")), nil
		},
		markFiscalDocumentAuthorizedFn: func(_ context.Context, _ tenancy.StoreScope, _ database.MarkFiscalDocumentAuthorizedForStoreParams) (database.FiscalDocument, error) {
			return fiscalDocumentFixture(mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8ac"), sale.id, database.FiscalDocumentStatusAUTHORIZED), nil
		},
	}

	svc := checkout.NewService(&checkoutFakeTxManager{tx: tx}, &fiscal.MockProvider{Now: func() time.Time { return time.Unix(1, 0) }})

	first, err := svc.Checkout(context.Background(), testScope, sale.id.String(), checkoutInputWithTotal("100.00"))
	if err != nil {
		t.Fatalf("first checkout returned error: %v", err)
	}
	if first.Sale.Status != string(database.SaleStatusCOMPLETED) {
		t.Fatalf("unexpected first checkout result: %+v", first.Sale)
	}

	_, err = svc.Checkout(context.Background(), testScope, sale.id.String(), checkoutInputWithTotal("100.00"))
	if !errors.Is(err, checkout.ErrSaleAlreadyCompleted) {
		t.Fatalf("expected checkout.ErrSaleAlreadyCompleted, got %v", err)
	}
}

func TestCheckoutFiscalFailureKeepsCommercialFlowCompleted(t *testing.T) {
	sale := newCheckoutSaleFixture()
	productID := mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8aa")
	docID := mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8ac")
	tx := &checkoutFakeTxQueries{
		lockSaleByIDFn: func(_ context.Context, _ tenancy.StoreScope, _ pgtype.UUID) (database.Sale, error) {
			return sale.lock(database.SaleStatusOPEN), nil
		},
		listSaleItemsBySaleIDFn: func(_ context.Context, _ tenancy.StoreScope, _ pgtype.UUID) ([]database.SaleItem, error) {
			return []database.SaleItem{checkoutItemFixture(productID, "50.00", "2.000", "0.00", "100.00")}, nil
		},
		getPaymentMethodByIDFn: func(_ context.Context, _ tenancy.OrganizationScope, _ pgtype.UUID) (database.PaymentMethod, error) {
			return paymentMethodFixture(testPaymentMethodID, "CASH", "Dinheiro", database.PaymentMethodKindCASH, true, false, false, 1, false), nil
		},
		decreaseInventoryFn: func(_ context.Context, _ tenancy.StoreScope, _ database.DecreaseInventoryForStoreParams) (database.DecreaseInventoryForStoreRow, error) {
			return decreaseInventoryRowFixture(productID, "10.000", "8.000"), nil
		},
		createInventoryMovementFn: func(_ context.Context, _ tenancy.ActorScope, _ database.CreateInventoryMovementForStoreParams) (database.CreateInventoryMovementForStoreRow, error) {
			return inventoryMovementFixture(productID, database.InventoryMovementTypeSALE, "2.000", "10.000", "8.000"), nil
		},
		createPaymentFn: func(_ context.Context, _ tenancy.ActorScope, _ database.CreatePaymentForStoreParams) (database.CreatePaymentForStoreRow, error) {
			return createPaymentRowFixture(sale.id, testPaymentMethodID, "100.00", "100.00", "0.00", 1, "checkout:"+sale.id.String()+":payment:0"), nil
		},
		approvePaymentFn: func(_ context.Context, _ tenancy.StoreScope, _ database.ApprovePaymentForStoreParams) (database.Payment, error) {
			return approvePaymentRowFixture(sale.id, testPaymentMethodID, "100.00", "100.00", "0.00", 1, database.PaymentStatusAPPROVED), nil
		},
		completeSaleFn: func(_ context.Context, _ tenancy.ActorScope, _ database.CompleteSaleForStoreParams) (database.Sale, error) {
			return sale.complete(database.SaleStatusCOMPLETED), nil
		},
		createFiscalDocumentFn: func(_ context.Context, _ tenancy.StoreScope, _ database.CreateFiscalDocumentForStoreParams) (database.CreateFiscalDocumentForStoreRow, error) {
			return fiscalDocumentCreateFixture(sale.id, docID), nil
		},
		markFiscalDocumentErrorFn: func(_ context.Context, _ tenancy.StoreScope, _ database.MarkFiscalDocumentErrorForStoreParams) (database.FiscalDocument, error) {
			return fiscalDocumentFixture(docID, sale.id, database.FiscalDocumentStatusERROR), nil
		},
	}

	svc := checkout.NewService(&checkoutFakeTxManager{tx: tx}, &fiscal.MockProvider{Fail: true})

	resp, err := svc.Checkout(context.Background(), testScope, sale.id.String(), checkoutInputWithTotal("100.00"))
	if err != nil {
		t.Fatalf("Checkout returned error: %v", err)
	}
	if resp.Sale.Status != string(database.SaleStatusCOMPLETED) {
		t.Fatalf("unexpected sale status: %+v", resp.Sale)
	}
	if resp.FiscalDocument.Status != string(database.FiscalDocumentStatusERROR) {
		t.Fatalf("unexpected fiscal document status: %+v", resp.FiscalDocument)
	}
}

func TestValidatePaymentsRequiresPayments(t *testing.T) {
	svc := checkout.NewService(&checkoutFakeTxManager{tx: &checkoutFakeTxQueries{}}, nil)
	_, err := svc.ValidatePayments(context.Background(), &checkoutFakeTxQueries{}, testScope, mustNumeric("100.00"), nil)
	if !errors.Is(err, checkout.ErrPaymentsRequired) {
		t.Fatalf("expected checkout.ErrPaymentsRequired, got %v", err)
	}
}

func TestPaymentIdempotencyKey(t *testing.T) {
	var id pgtype.UUID
	_ = id.Scan("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8aa")
	got := checkout.PaymentIdempotencyKey(id, 2)
	if got == "" {
		t.Fatalf("expected idempotency key")
	}
}

type checkoutFakeTxManager struct {
	tx         checkout.TxQueries
	committed  bool
	rolledBack bool
}

func (m *checkoutFakeTxManager) WithTx(_ context.Context, _ tenancy.ActorScope, fn func(checkout.TxQueries) error) error {
	if m.tx == nil {
		return errors.New("nil transaction")
	}

	err := fn(m.tx)
	if err != nil {
		m.rolledBack = true
		return err
	}

	m.committed = true
	return nil
}

type checkoutFakeTxQueries struct {
	lockSaleByIDFn                 func(context.Context, tenancy.StoreScope, pgtype.UUID) (database.Sale, error)
	listSaleItemsBySaleIDFn        func(context.Context, tenancy.StoreScope, pgtype.UUID) ([]database.SaleItem, error)
	getPaymentMethodByIDFn         func(context.Context, tenancy.OrganizationScope, pgtype.UUID) (database.PaymentMethod, error)
	getInventoryByProductIDFn      func(context.Context, tenancy.StoreScope, pgtype.UUID) (database.Inventory, error)
	decreaseInventoryFn            func(context.Context, tenancy.StoreScope, database.DecreaseInventoryForStoreParams) (database.DecreaseInventoryForStoreRow, error)
	createInventoryMovementFn      func(context.Context, tenancy.ActorScope, database.CreateInventoryMovementForStoreParams) (database.CreateInventoryMovementForStoreRow, error)
	createPaymentFn                func(context.Context, tenancy.ActorScope, database.CreatePaymentForStoreParams) (database.CreatePaymentForStoreRow, error)
	approvePaymentFn               func(context.Context, tenancy.StoreScope, database.ApprovePaymentForStoreParams) (database.Payment, error)
	completeSaleFn                 func(context.Context, tenancy.ActorScope, database.CompleteSaleForStoreParams) (database.Sale, error)
	createFiscalDocumentFn         func(context.Context, tenancy.StoreScope, database.CreateFiscalDocumentForStoreParams) (database.CreateFiscalDocumentForStoreRow, error)
	markFiscalDocumentAuthorizedFn func(context.Context, tenancy.StoreScope, database.MarkFiscalDocumentAuthorizedForStoreParams) (database.FiscalDocument, error)
	markFiscalDocumentErrorFn      func(context.Context, tenancy.StoreScope, database.MarkFiscalDocumentErrorForStoreParams) (database.FiscalDocument, error)
}

func (f *checkoutFakeTxQueries) LockSaleByID(ctx context.Context, scope tenancy.StoreScope, id pgtype.UUID) (database.Sale, error) {
	if f.lockSaleByIDFn != nil {
		return f.lockSaleByIDFn(ctx, scope, id)
	}
	return database.Sale{}, nil
}

func (f *checkoutFakeTxQueries) ListSaleItemsBySaleID(ctx context.Context, scope tenancy.StoreScope, saleID pgtype.UUID) ([]database.SaleItem, error) {
	if f.listSaleItemsBySaleIDFn != nil {
		return f.listSaleItemsBySaleIDFn(ctx, scope, saleID)
	}
	return nil, nil
}

func (f *checkoutFakeTxQueries) GetPaymentMethodByID(ctx context.Context, scope tenancy.OrganizationScope, id pgtype.UUID) (database.PaymentMethod, error) {
	if f.getPaymentMethodByIDFn != nil {
		return f.getPaymentMethodByIDFn(ctx, scope, id)
	}
	return database.PaymentMethod{}, pgx.ErrNoRows
}

func (f *checkoutFakeTxQueries) GetInventoryByProductID(ctx context.Context, scope tenancy.StoreScope, id pgtype.UUID) (database.Inventory, error) {
	if f.getInventoryByProductIDFn != nil {
		return f.getInventoryByProductIDFn(ctx, scope, id)
	}
	return database.Inventory{}, pgx.ErrNoRows
}

func (f *checkoutFakeTxQueries) DecreaseInventory(ctx context.Context, scope tenancy.StoreScope, arg database.DecreaseInventoryForStoreParams) (database.DecreaseInventoryForStoreRow, error) {
	if f.decreaseInventoryFn != nil {
		return f.decreaseInventoryFn(ctx, scope, arg)
	}
	return database.DecreaseInventoryForStoreRow{}, pgx.ErrNoRows
}

func (f *checkoutFakeTxQueries) CreateInventoryMovement(ctx context.Context, scope tenancy.ActorScope, arg database.CreateInventoryMovementForStoreParams) (database.CreateInventoryMovementForStoreRow, error) {
	if f.createInventoryMovementFn != nil {
		return f.createInventoryMovementFn(ctx, scope, arg)
	}
	return database.CreateInventoryMovementForStoreRow{}, nil
}

func (f *checkoutFakeTxQueries) CreatePayment(ctx context.Context, scope tenancy.ActorScope, arg database.CreatePaymentForStoreParams) (database.CreatePaymentForStoreRow, error) {
	if f.createPaymentFn != nil {
		return f.createPaymentFn(ctx, scope, arg)
	}
	return database.CreatePaymentForStoreRow{}, nil
}

func (f *checkoutFakeTxQueries) ApprovePayment(ctx context.Context, scope tenancy.StoreScope, arg database.ApprovePaymentForStoreParams) (database.Payment, error) {
	if f.approvePaymentFn != nil {
		return f.approvePaymentFn(ctx, scope, arg)
	}
	return database.Payment{}, nil
}

func (f *checkoutFakeTxQueries) CompleteSale(ctx context.Context, scope tenancy.ActorScope, arg database.CompleteSaleForStoreParams) (database.Sale, error) {
	if f.completeSaleFn != nil {
		return f.completeSaleFn(ctx, scope, arg)
	}
	return database.Sale{}, nil
}

func (f *checkoutFakeTxQueries) CreateFiscalDocument(ctx context.Context, scope tenancy.StoreScope, arg database.CreateFiscalDocumentForStoreParams) (database.CreateFiscalDocumentForStoreRow, error) {
	if f.createFiscalDocumentFn != nil {
		return f.createFiscalDocumentFn(ctx, scope, arg)
	}
	return database.CreateFiscalDocumentForStoreRow{}, nil
}

func (f *checkoutFakeTxQueries) MarkFiscalDocumentAuthorized(ctx context.Context, scope tenancy.StoreScope, arg database.MarkFiscalDocumentAuthorizedForStoreParams) (database.FiscalDocument, error) {
	if f.markFiscalDocumentAuthorizedFn != nil {
		return f.markFiscalDocumentAuthorizedFn(ctx, scope, arg)
	}
	return database.FiscalDocument{}, nil
}

func (f *checkoutFakeTxQueries) MarkFiscalDocumentError(ctx context.Context, scope tenancy.StoreScope, arg database.MarkFiscalDocumentErrorForStoreParams) (database.FiscalDocument, error) {
	if f.markFiscalDocumentErrorFn != nil {
		return f.markFiscalDocumentErrorFn(ctx, scope, arg)
	}
	return database.FiscalDocument{}, nil
}

type fakeReadStore struct{}

type checkoutSaleFixture struct {
	id             pgtype.UUID
	number         int64
	subtotal       pgtype.Numeric
	discount       pgtype.Numeric
	addition       pgtype.Numeric
	total          pgtype.Numeric
	openedAt       pgtype.Timestamptz
	completedAt    pgtype.Timestamptz
	cancelledAt    pgtype.Timestamptz
	createdAt      pgtype.Timestamptz
	updatedAt      pgtype.Timestamptz
	idempotencyKey string
}

func newCheckoutSaleFixture() checkoutSaleFixture {
	now := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	return checkoutSaleFixture{
		id:             mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8a9"),
		number:         77,
		subtotal:       mustNumeric("100.00"),
		discount:       mustNumeric("0.00"),
		addition:       mustNumeric("0.00"),
		total:          mustNumeric("100.00"),
		openedAt:       timestamptz(now),
		createdAt:      timestamptz(now),
		updatedAt:      timestamptz(now.Add(1 * time.Minute)),
		idempotencyKey: "checkout-sale-1",
	}
}

func (f checkoutSaleFixture) lock(status database.SaleStatus) database.Sale {
	row := database.Sale{
		ID:             f.id,
		OrganizationID: testScope.OrganizationID,
		StoreID:        testScope.StoreID,
		Number:         f.number,
		Status:         status,
		Subtotal:       f.subtotal,
		Discount:       f.discount,
		Addition:       f.addition,
		Total:          f.total,
		OpenedAt:       f.openedAt,
		CreatedAt:      f.createdAt,
		UpdatedAt:      f.updatedAt,
		IdempotencyKey: f.idempotencyKey,
	}
	if status == database.SaleStatusCOMPLETED {
		row.CompletedAt = timestamptz(f.createdAt.Time.Add(1 * time.Minute))
	}
	if status == database.SaleStatusCANCELLED {
		row.CancelledAt = timestamptz(f.createdAt.Time.Add(1 * time.Minute))
	}
	return row
}

func (f checkoutSaleFixture) complete(status database.SaleStatus) database.Sale {
	row := database.Sale{
		ID:             f.id,
		OrganizationID: testScope.OrganizationID,
		StoreID:        testScope.StoreID,
		Number:         f.number,
		Status:         status,
		Subtotal:       f.subtotal,
		Discount:       f.discount,
		Addition:       f.addition,
		Total:          f.total,
		OpenedAt:       f.openedAt,
		CreatedAt:      f.createdAt,
		UpdatedAt:      f.updatedAt,
		IdempotencyKey: f.idempotencyKey,
	}
	if status == database.SaleStatusCOMPLETED {
		row.CompletedAt = timestamptz(f.createdAt.Time.Add(1 * time.Minute))
	}
	if status == database.SaleStatusCANCELLED {
		row.CancelledAt = timestamptz(f.createdAt.Time.Add(1 * time.Minute))
	}
	return row
}

func checkoutItemFixture(productID pgtype.UUID, unitPrice, quantity, discount, total string) database.SaleItem {
	return database.SaleItem{
		ID:             mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b800"),
		OrganizationID: testScope.OrganizationID,
		StoreID:        testScope.StoreID,
		SaleID:         mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8a9"),
		ProductID:      productID,
		ProductName:    "Produto",
		ProductSKU:     "SKU-001",
		UnitPrice:      mustNumeric(unitPrice),
		Quantity:       mustNumeric(quantity),
		Discount:       mustNumeric(discount),
		Total:          mustNumeric(total),
		CreatedAt:      timestamptz(time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)),
	}
}

func checkoutItemWithPriceFixture(productID pgtype.UUID, unitPrice, quantity, discount, total string) database.SaleItem {
	row := checkoutItemFixture(productID, unitPrice, quantity, discount, total)
	row.ID = mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b801")
	return row
}

func paymentMethodFixture(id pgtype.UUID, code, name string, kind database.PaymentMethodKind, active bool, allowsChange bool, allowsInstallments bool, maxInstallments int16, requiresExternalReference bool) database.PaymentMethod {
	return database.PaymentMethod{
		ID:                        id,
		Code:                      code,
		Name:                      name,
		Kind:                      kind,
		AllowsChange:              allowsChange,
		RequiresExternalReference: requiresExternalReference,
		AllowsInstallments:        allowsInstallments,
		MaxInstallments:           maxInstallments,
		IsActive:                  active,
		FeePercentage:             mustNumeric("0.00"),
		CreatedAt:                 timestamptz(time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)),
		UpdatedAt:                 timestamptz(time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)),
	}
}

func inventoryFixture(productID pgtype.UUID, quantity string) database.Inventory {
	return database.Inventory{
		ProductID: productID,
		Quantity:  mustNumeric(quantity),
		CreatedAt: timestamptz(time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)),
		UpdatedAt: timestamptz(time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)),
	}
}

func decreaseInventoryRowFixture(productID pgtype.UUID, previous, current string) database.DecreaseInventoryForStoreRow {
	return database.DecreaseInventoryForStoreRow{
		OrganizationID:   testScope.OrganizationID,
		StoreID:          testScope.StoreID,
		ProductID:        productID,
		PreviousQuantity: mustNumeric(previous),
		CurrentQuantity:  mustNumeric(current),
		CreatedAt:        timestamptz(time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)),
		UpdatedAt:        timestamptz(time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)),
	}
}

func inventoryMovementFixture(productID pgtype.UUID, movementType database.InventoryMovementType, quantity, previous, current string) database.CreateInventoryMovementForStoreRow {
	return database.CreateInventoryMovementForStoreRow{
		OrganizationID:   testScope.OrganizationID,
		StoreID:          testScope.StoreID,
		ID:               mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8b0"),
		ProductID:        productID,
		MovementType:     movementType,
		Quantity:         mustNumeric(quantity),
		PreviousQuantity: mustNumeric(previous),
		CurrentQuantity:  mustNumeric(current),
		Reason:           pgtype.Text{},
		ReferenceType:    "sale",
		ReferenceID:      mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8a9"),
		CreatedAt:        timestamptz(time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)),
	}
}

func createPaymentRowFixture(saleID, methodID pgtype.UUID, amount, received, change string, installments int16, idempotencyKey string) database.CreatePaymentForStoreRow {
	return database.CreatePaymentForStoreRow{
		ID:                mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8c0"),
		OrganizationID:    testScope.OrganizationID,
		StoreID:           testScope.StoreID,
		SaleID:            saleID,
		PaymentMethodID:   methodID,
		Status:            database.PaymentStatusPENDING,
		Amount:            mustNumeric(amount),
		ReceivedAmount:    mustNumeric(received),
		ChangeAmount:      mustNumeric(change),
		Installments:      installments,
		ExternalReference: pgtype.Text{},
		PaidAt:            pgtype.Timestamptz{},
		CreatedAt:         timestamptz(time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)),
		UpdatedAt:         timestamptz(time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)),
		IdempotencyKey:    idempotencyKey,
	}
}

func approvePaymentRowFixture(saleID, methodID pgtype.UUID, amount, received, change string, installments int16, status database.PaymentStatus) database.Payment {
	return database.Payment{
		ID:                mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8c1"),
		OrganizationID:    testScope.OrganizationID,
		StoreID:           testScope.StoreID,
		SaleID:            saleID,
		PaymentMethodID:   methodID,
		Status:            status,
		Amount:            mustNumeric(amount),
		ReceivedAmount:    mustNumeric(received),
		ChangeAmount:      mustNumeric(change),
		Installments:      installments,
		ExternalReference: pgtype.Text{},
		PaidAt:            timestamptz(time.Date(2026, 7, 15, 12, 0, 1, 0, time.UTC)),
		CreatedAt:         timestamptz(time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)),
		UpdatedAt:         timestamptz(time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)),
		IdempotencyKey:    "payment",
	}
}

func fiscalDocumentCreateFixture(saleID, id pgtype.UUID) database.CreateFiscalDocumentForStoreRow {
	return database.CreateFiscalDocumentForStoreRow{
		ID:             id,
		OrganizationID: testScope.OrganizationID,
		StoreID:        testScope.StoreID,
		SaleID:         saleID,
		Status:         database.FiscalDocumentStatusPENDING,
		Environment:    database.FiscalEnvironmentHOMOLOGATION,
		DocumentModel:  65,
		CreatedAt:      timestamptz(time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)),
		UpdatedAt:      timestamptz(time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)),
	}
}

func fiscalDocumentFixture(id, saleID pgtype.UUID, status database.FiscalDocumentStatus) database.FiscalDocument {
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
		row.IssuedAt = timestamptz(now.Add(1 * time.Minute))
	}
	if status == database.FiscalDocumentStatusERROR {
		row.ErrorCode = pgtype.Text{String: "mock_authorization_failed", Valid: true}
		row.ErrorMessage = pgtype.Text{String: "Fiscal authorization failed", Valid: true}
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

func numericString(value pgtype.Numeric) string {
	got, err := checkout.MoneyToString(value)
	if err != nil {
		panic(err)
	}
	return got
}

func quantityString(value pgtype.Numeric) string {
	got, err := checkout.QuantityToString(value)
	if err != nil {
		panic(err)
	}
	return got
}

func strPtr(value string) *string {
	return &value
}

func intPtr(value int) *int {
	return &value
}
