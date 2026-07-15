package inventory

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestCreateEntryValid(t *testing.T) {
	txQueries := &fakeTxQueries{
		getProductByIDFn: func(context.Context, pgtype.UUID) (database.Product, error) {
			return productFixture(true), nil
		},
		increaseInventoryFn: func(context.Context, database.IncreaseInventoryParams) (database.IncreaseInventoryRow, error) {
			return increaseRowFixture("8.000", "10.500"), nil
		},
		createInventoryMovementFn: func(_ context.Context, arg database.CreateInventoryMovementParams) (database.InventoryMovement, error) {
			if arg.MovementType != database.InventoryMovementTypePURCHASE {
				t.Fatalf("unexpected movement type: %s", arg.MovementType)
			}
			if arg.ReferenceType != "purchase" {
				t.Fatalf("unexpected reference type: %s", arg.ReferenceType)
			}
			if arg.ReferenceID.String() != productFixture(true).ID.String() {
				t.Fatalf("unexpected reference id: %s", arg.ReferenceID.String())
			}
			return movementFixture(database.InventoryMovementTypePURCHASE, "2.500", "8.000", "10.500"), nil
		},
	}

	txManager := &fakeTxManager{tx: txQueries}
	svc := NewService(&fakeReadStore{}, txManager)

	resp, err := svc.CreateEntry(context.Background(), CreateInventoryEntryInput{
		ProductID:     productFixture(true).ID.String(),
		Quantity:      "2.500",
		Reason:        strPtr("Compra de fornecedor"),
		ReferenceType: "purchase",
		ReferenceID:   productFixture(true).ID.String(),
	})
	if err != nil {
		t.Fatalf("CreateEntry returned error: %v", err)
	}

	if !txManager.committed || txManager.rolledBack {
		t.Fatalf("unexpected tx state: committed=%v rolledBack=%v", txManager.committed, txManager.rolledBack)
	}
	if resp.Inventory.ProductID != productFixture(true).ID.String() {
		t.Fatalf("unexpected inventory product id: %s", resp.Inventory.ProductID)
	}
	if resp.Inventory.PreviousQuantity != "8.000" || resp.Inventory.CurrentQuantity != "10.500" {
		t.Fatalf("unexpected inventory summary: %+v", resp.Inventory)
	}
	if resp.Movement.Type != string(database.InventoryMovementTypePURCHASE) {
		t.Fatalf("unexpected movement type: %+v", resp.Movement)
	}
}

func TestCreateAdjustmentValidIn(t *testing.T) {
	txQueries := &fakeTxQueries{
		getProductByIDFn: func(context.Context, pgtype.UUID) (database.Product, error) {
			return productFixture(true), nil
		},
		increaseInventoryFn: func(context.Context, database.IncreaseInventoryParams) (database.IncreaseInventoryRow, error) {
			return increaseRowFixture("10.500", "12.000"), nil
		},
		createInventoryMovementFn: func(_ context.Context, arg database.CreateInventoryMovementParams) (database.InventoryMovement, error) {
			if arg.MovementType != database.InventoryMovementTypeADJUSTMENTIN {
				t.Fatalf("unexpected movement type: %s", arg.MovementType)
			}
			return movementFixture(database.InventoryMovementTypeADJUSTMENTIN, "1.500", "10.500", "12.000"), nil
		},
	}

	txManager := &fakeTxManager{tx: txQueries}
	svc := NewService(&fakeReadStore{}, txManager)

	resp, err := svc.CreateAdjustment(context.Background(), CreateInventoryAdjustmentInput{
		ProductID:     productFixture(true).ID.String(),
		Direction:     "IN",
		Quantity:      "1.500",
		Reason:        "Ajuste de entrada",
		ReferenceType: "manual_adjustment",
		ReferenceID:   productFixture(true).ID.String(),
	})
	if err != nil {
		t.Fatalf("CreateAdjustment returned error: %v", err)
	}

	if !txManager.committed || txManager.rolledBack {
		t.Fatalf("unexpected tx state: committed=%v rolledBack=%v", txManager.committed, txManager.rolledBack)
	}
	if resp.Inventory.PreviousQuantity != "10.500" || resp.Inventory.CurrentQuantity != "12.000" {
		t.Fatalf("unexpected inventory summary: %+v", resp.Inventory)
	}
	if resp.Movement.Type != string(database.InventoryMovementTypeADJUSTMENTIN) {
		t.Fatalf("unexpected movement type: %+v", resp.Movement)
	}
}

func TestCreateAdjustmentValidOut(t *testing.T) {
	txQueries := &fakeTxQueries{
		getProductByIDFn: func(context.Context, pgtype.UUID) (database.Product, error) {
			return productFixture(true), nil
		},
		decreaseInventoryFn: func(context.Context, database.DecreaseInventoryParams) (database.DecreaseInventoryRow, error) {
			return decreaseRowFixture("12.000", "10.000"), nil
		},
		createInventoryMovementFn: func(_ context.Context, arg database.CreateInventoryMovementParams) (database.InventoryMovement, error) {
			if arg.MovementType != database.InventoryMovementTypeADJUSTMENTOUT {
				t.Fatalf("unexpected movement type: %s", arg.MovementType)
			}
			return movementFixture(database.InventoryMovementTypeADJUSTMENTOUT, "2.000", "12.000", "10.000"), nil
		},
	}

	txManager := &fakeTxManager{tx: txQueries}
	svc := NewService(&fakeReadStore{}, txManager)

	resp, err := svc.CreateAdjustment(context.Background(), CreateInventoryAdjustmentInput{
		ProductID:     productFixture(true).ID.String(),
		Direction:     "OUT",
		Quantity:      "2.000",
		Reason:        "Produto avariado",
		ReferenceType: "manual_adjustment",
		ReferenceID:   productFixture(true).ID.String(),
	})
	if err != nil {
		t.Fatalf("CreateAdjustment returned error: %v", err)
	}

	if !txManager.committed || txManager.rolledBack {
		t.Fatalf("unexpected tx state: committed=%v rolledBack=%v", txManager.committed, txManager.rolledBack)
	}
	if resp.Inventory.PreviousQuantity != "12.000" || resp.Inventory.CurrentQuantity != "10.000" {
		t.Fatalf("unexpected inventory summary: %+v", resp.Inventory)
	}
	if resp.Movement.Type != string(database.InventoryMovementTypeADJUSTMENTOUT) {
		t.Fatalf("unexpected movement type: %+v", resp.Movement)
	}
}

func TestCreateEntryProductNotFound(t *testing.T) {
	txQueries := &fakeTxQueries{
		getProductByIDFn: func(context.Context, pgtype.UUID) (database.Product, error) {
			return database.Product{}, pgx.ErrNoRows
		},
	}

	txManager := &fakeTxManager{tx: txQueries}
	svc := NewService(&fakeReadStore{}, txManager)

	_, err := svc.CreateEntry(context.Background(), CreateInventoryEntryInput{
		ProductID:     productFixture(true).ID.String(),
		Quantity:      "1.000",
		ReferenceType: "purchase",
		ReferenceID:   productFixture(true).ID.String(),
	})
	if !errors.Is(err, ErrProductNotFound) {
		t.Fatalf("expected ErrProductNotFound, got %v", err)
	}
	if !txManager.rolledBack || txManager.committed {
		t.Fatalf("expected rollback without commit, got committed=%v rolledBack=%v", txManager.committed, txManager.rolledBack)
	}
}

func TestCreateAdjustmentInsufficientInventory(t *testing.T) {
	txQueries := &fakeTxQueries{
		getProductByIDFn: func(context.Context, pgtype.UUID) (database.Product, error) {
			return productFixture(true), nil
		},
		decreaseInventoryFn: func(context.Context, database.DecreaseInventoryParams) (database.DecreaseInventoryRow, error) {
			return database.DecreaseInventoryRow{}, pgx.ErrNoRows
		},
		getInventoryByProductIDFn: func(context.Context, pgtype.UUID) (database.Inventory, error) {
			return inventoryFixture("1.000"), nil
		},
	}

	txManager := &fakeTxManager{tx: txQueries}
	svc := NewService(&fakeReadStore{}, txManager)

	_, err := svc.CreateAdjustment(context.Background(), CreateInventoryAdjustmentInput{
		ProductID:     productFixture(true).ID.String(),
		Direction:     "OUT",
		Quantity:      "2.000",
		Reason:        "Produto avariado",
		ReferenceType: "manual_adjustment",
		ReferenceID:   productFixture(true).ID.String(),
	})
	if !errors.Is(err, ErrInsufficientInventory) {
		t.Fatalf("expected ErrInsufficientInventory, got %v", err)
	}
	if !txManager.rolledBack || txManager.committed {
		t.Fatalf("expected rollback without commit, got committed=%v rolledBack=%v", txManager.committed, txManager.rolledBack)
	}
}

func TestCreateEntryQuantityValidation(t *testing.T) {
	svc := NewService(&fakeReadStore{}, &fakeTxManager{})

	cases := []struct {
		name     string
		quantity string
		field    string
	}{
		{name: "zero", quantity: "0", field: "quantity"},
		{name: "negative", quantity: "-1.000", field: "quantity"},
		{name: "invalid", quantity: "abc", field: "quantity"},
		{name: "too many decimals", quantity: "1.2345", field: "quantity"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.CreateEntry(context.Background(), CreateInventoryEntryInput{
				ProductID:     productFixture(true).ID.String(),
				Quantity:      tc.quantity,
				ReferenceType: "purchase",
				ReferenceID:   productFixture(true).ID.String(),
			})
			requireValidationField(t, err, tc.field)
		})
	}
}

func TestCreateAdjustmentValidation(t *testing.T) {
	svc := NewService(&fakeReadStore{}, &fakeTxManager{})

	t.Run("direction invalid", func(t *testing.T) {
		_, err := svc.CreateAdjustment(context.Background(), CreateInventoryAdjustmentInput{
			ProductID:     productFixture(true).ID.String(),
			Direction:     "SIDEWAYS",
			Quantity:      "1.000",
			Reason:        "Produto avariado",
			ReferenceType: "manual_adjustment",
			ReferenceID:   productFixture(true).ID.String(),
		})
		requireValidationField(t, err, "direction")
	})

	t.Run("reason blank", func(t *testing.T) {
		_, err := svc.CreateAdjustment(context.Background(), CreateInventoryAdjustmentInput{
			ProductID:     productFixture(true).ID.String(),
			Direction:     "OUT",
			Quantity:      "1.000",
			Reason:        "   ",
			ReferenceType: "manual_adjustment",
			ReferenceID:   productFixture(true).ID.String(),
		})
		requireValidationField(t, err, "reason")
	})
}

func TestDuplicateOperation(t *testing.T) {
	txQueries := &fakeTxQueries{
		getProductByIDFn: func(context.Context, pgtype.UUID) (database.Product, error) {
			return productFixture(true), nil
		},
		increaseInventoryFn: func(context.Context, database.IncreaseInventoryParams) (database.IncreaseInventoryRow, error) {
			return increaseRowFixture("8.000", "10.000"), nil
		},
		createInventoryMovementFn: func(context.Context, database.CreateInventoryMovementParams) (database.InventoryMovement, error) {
			return database.InventoryMovement{}, &pgconn.PgError{Code: "23505", ConstraintName: "inventory_movements_reference_unique"}
		},
	}

	txManager := &fakeTxManager{tx: txQueries}
	svc := NewService(&fakeReadStore{}, txManager)

	_, err := svc.CreateEntry(context.Background(), CreateInventoryEntryInput{
		ProductID:     productFixture(true).ID.String(),
		Quantity:      "2.000",
		ReferenceType: "purchase",
		ReferenceID:   productFixture(true).ID.String(),
	})
	if !errors.Is(err, ErrInventoryOperationAlreadyProcessed) {
		t.Fatalf("expected ErrInventoryOperationAlreadyProcessed, got %v", err)
	}
	if !txManager.rolledBack || txManager.committed {
		t.Fatalf("expected rollback without commit, got committed=%v rolledBack=%v", txManager.committed, txManager.rolledBack)
	}
}

func TestListInventoryDefaultPagination(t *testing.T) {
	var capturedCount database.CountInventoryParams
	var capturedList database.ListInventoryParams

	svc := NewService(&fakeReadStore{
		countInventoryFn: func(_ context.Context, arg database.CountInventoryParams) (int64, error) {
			capturedCount = arg
			return 1, nil
		},
		listInventoryFn: func(_ context.Context, arg database.ListInventoryParams) ([]database.ListInventoryRow, error) {
			capturedList = arg
			return []database.ListInventoryRow{inventoryListRowFixture()}, nil
		},
	}, &fakeTxManager{})

	resp, err := svc.ListInventory(context.Background(), ListInventoryInput{})
	if err != nil {
		t.Fatalf("ListInventory returned error: %v", err)
	}

	if resp.Pagination.Page != 1 || resp.Pagination.PageSize != 20 || resp.Pagination.Total != 1 || resp.Pagination.TotalPages != 1 {
		t.Fatalf("unexpected pagination: %+v", resp.Pagination)
	}
	if capturedList.PageSize != 20 || capturedList.PageOffset != 0 {
		t.Fatalf("unexpected list params: %+v", capturedList)
	}
	activeOnly, ok := capturedCount.ActiveOnly.(bool)
	if !ok || activeOnly {
		t.Fatalf("unexpected activeOnly: %#v", capturedCount.ActiveOnly)
	}
	if capturedCount.Search.Valid {
		t.Fatalf("expected empty search, got %#v", capturedCount.Search)
	}
}

func TestListInventoryPageSizeMaximum(t *testing.T) {
	svc := NewService(&fakeReadStore{}, &fakeTxManager{})

	pageSize := 101
	_, err := svc.ListInventory(context.Background(), ListInventoryInput{PageSize: &pageSize})
	requireValidationField(t, err, "pageSize")
}

func TestCreateEntryRollbackOnMovementError(t *testing.T) {
	txQueries := &fakeTxQueries{
		getProductByIDFn: func(context.Context, pgtype.UUID) (database.Product, error) {
			return productFixture(true), nil
		},
		increaseInventoryFn: func(context.Context, database.IncreaseInventoryParams) (database.IncreaseInventoryRow, error) {
			return increaseRowFixture("8.000", "10.000"), nil
		},
		createInventoryMovementFn: func(context.Context, database.CreateInventoryMovementParams) (database.InventoryMovement, error) {
			return database.InventoryMovement{}, errors.New("insert failed")
		},
	}

	txManager := &fakeTxManager{tx: txQueries}
	svc := NewService(&fakeReadStore{}, txManager)

	_, err := svc.CreateEntry(context.Background(), CreateInventoryEntryInput{
		ProductID:     productFixture(true).ID.String(),
		Quantity:      "2.000",
		ReferenceType: "purchase",
		ReferenceID:   productFixture(true).ID.String(),
	})
	if err == nil {
		t.Fatalf("expected error")
	}
	if !txManager.rolledBack || txManager.committed {
		t.Fatalf("expected rollback without commit, got committed=%v rolledBack=%v", txManager.committed, txManager.rolledBack)
	}
}

func TestCreateAdjustmentRollbackOnBalanceError(t *testing.T) {
	txQueries := &fakeTxQueries{
		getProductByIDFn: func(context.Context, pgtype.UUID) (database.Product, error) {
			return productFixture(true), nil
		},
		decreaseInventoryFn: func(context.Context, database.DecreaseInventoryParams) (database.DecreaseInventoryRow, error) {
			return database.DecreaseInventoryRow{}, errors.New("update failed")
		},
	}

	txManager := &fakeTxManager{tx: txQueries}
	svc := NewService(&fakeReadStore{}, txManager)

	_, err := svc.CreateAdjustment(context.Background(), CreateInventoryAdjustmentInput{
		ProductID:     productFixture(true).ID.String(),
		Direction:     "OUT",
		Quantity:      "2.000",
		Reason:        "Produto avariado",
		ReferenceType: "manual_adjustment",
		ReferenceID:   productFixture(true).ID.String(),
	})
	if err == nil {
		t.Fatalf("expected error")
	}
	if !txManager.rolledBack || txManager.committed {
		t.Fatalf("expected rollback without commit, got committed=%v rolledBack=%v", txManager.committed, txManager.rolledBack)
	}
}

type fakeReadStore struct {
	getProductByIDFn                     func(context.Context, pgtype.UUID) (database.Product, error)
	getInventoryByProductIDFn            func(context.Context, pgtype.UUID) (database.Inventory, error)
	listInventoryFn                      func(context.Context, database.ListInventoryParams) ([]database.ListInventoryRow, error)
	countInventoryFn                     func(context.Context, database.CountInventoryParams) (int64, error)
	listInventoryMovementsByProductIDFn  func(context.Context, database.ListInventoryMovementsByProductIDParams) ([]database.InventoryMovement, error)
	countInventoryMovementsByProductIDFn func(context.Context, database.CountInventoryMovementsByProductIDParams) (int64, error)
}

func (f *fakeReadStore) GetProductByID(ctx context.Context, id pgtype.UUID) (database.Product, error) {
	if f.getProductByIDFn == nil {
		panic("unexpected GetProductByID call")
	}
	return f.getProductByIDFn(ctx, id)
}

func (f *fakeReadStore) GetInventoryByProductID(ctx context.Context, id pgtype.UUID) (database.Inventory, error) {
	if f.getInventoryByProductIDFn == nil {
		panic("unexpected GetInventoryByProductID call")
	}
	return f.getInventoryByProductIDFn(ctx, id)
}

func (f *fakeReadStore) ListInventory(ctx context.Context, arg database.ListInventoryParams) ([]database.ListInventoryRow, error) {
	if f.listInventoryFn == nil {
		panic("unexpected ListInventory call")
	}
	return f.listInventoryFn(ctx, arg)
}

func (f *fakeReadStore) CountInventory(ctx context.Context, arg database.CountInventoryParams) (int64, error) {
	if f.countInventoryFn == nil {
		panic("unexpected CountInventory call")
	}
	return f.countInventoryFn(ctx, arg)
}

func (f *fakeReadStore) ListInventoryMovementsByProductID(ctx context.Context, arg database.ListInventoryMovementsByProductIDParams) ([]database.InventoryMovement, error) {
	if f.listInventoryMovementsByProductIDFn == nil {
		panic("unexpected ListInventoryMovementsByProductID call")
	}
	return f.listInventoryMovementsByProductIDFn(ctx, arg)
}

func (f *fakeReadStore) CountInventoryMovementsByProductID(ctx context.Context, arg database.CountInventoryMovementsByProductIDParams) (int64, error) {
	if f.countInventoryMovementsByProductIDFn == nil {
		panic("unexpected CountInventoryMovementsByProductID call")
	}
	return f.countInventoryMovementsByProductIDFn(ctx, arg)
}

type fakeTxQueries struct {
	getProductByIDFn          func(context.Context, pgtype.UUID) (database.Product, error)
	getInventoryByProductIDFn func(context.Context, pgtype.UUID) (database.Inventory, error)
	increaseInventoryFn       func(context.Context, database.IncreaseInventoryParams) (database.IncreaseInventoryRow, error)
	decreaseInventoryFn       func(context.Context, database.DecreaseInventoryParams) (database.DecreaseInventoryRow, error)
	createInventoryMovementFn func(context.Context, database.CreateInventoryMovementParams) (database.InventoryMovement, error)
}

func (f *fakeTxQueries) GetProductByID(ctx context.Context, id pgtype.UUID) (database.Product, error) {
	if f.getProductByIDFn == nil {
		panic("unexpected GetProductByID call")
	}
	return f.getProductByIDFn(ctx, id)
}

func (f *fakeTxQueries) GetInventoryByProductID(ctx context.Context, id pgtype.UUID) (database.Inventory, error) {
	if f.getInventoryByProductIDFn == nil {
		panic("unexpected GetInventoryByProductID call")
	}
	return f.getInventoryByProductIDFn(ctx, id)
}

func (f *fakeTxQueries) IncreaseInventory(ctx context.Context, arg database.IncreaseInventoryParams) (database.IncreaseInventoryRow, error) {
	if f.increaseInventoryFn == nil {
		panic("unexpected IncreaseInventory call")
	}
	return f.increaseInventoryFn(ctx, arg)
}

func (f *fakeTxQueries) DecreaseInventory(ctx context.Context, arg database.DecreaseInventoryParams) (database.DecreaseInventoryRow, error) {
	if f.decreaseInventoryFn == nil {
		panic("unexpected DecreaseInventory call")
	}
	return f.decreaseInventoryFn(ctx, arg)
}

func (f *fakeTxQueries) CreateInventoryMovement(ctx context.Context, arg database.CreateInventoryMovementParams) (database.InventoryMovement, error) {
	if f.createInventoryMovementFn == nil {
		panic("unexpected CreateInventoryMovement call")
	}
	return f.createInventoryMovementFn(ctx, arg)
}

type fakeTxManager struct {
	tx         TxQueries
	committed  bool
	rolledBack bool
}

func (f *fakeTxManager) WithTx(ctx context.Context, fn func(TxQueries) error) error {
	if f.tx == nil {
		panic("unexpected transaction")
	}

	err := fn(f.tx)
	if err != nil {
		f.rolledBack = true
		return err
	}

	f.committed = true
	return nil
}

func productFixture(active bool) database.Product {
	return database.Product{
		ID:        mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8a9"),
		SKU:       "COCA-2L",
		Barcode:   pgtype.Text{String: "7890000000000", Valid: true},
		Name:      "Coca-Cola 2L",
		Price:     mustNumeric("12.90"),
		Cost:      mustNumeric("8.50"),
		IsActive:  active,
		CreatedAt: mustTime("2026-07-15T10:00:00Z"),
		UpdatedAt: mustTime("2026-07-15T10:00:00Z"),
	}
}

func inventoryFixture(quantity string) database.Inventory {
	return database.Inventory{
		ProductID: productFixture(true).ID,
		Quantity:  mustNumeric(quantity),
		CreatedAt: mustTime("2026-07-15T10:00:00Z"),
		UpdatedAt: mustTime("2026-07-15T10:00:00Z"),
	}
}

func inventoryListRowFixture() database.ListInventoryRow {
	return database.ListInventoryRow{
		ProductID: productFixture(true).ID,
		SKU:       "COCA-2L",
		Barcode:   pgtype.Text{String: "7890000000000", Valid: true},
		Name:      "Coca-Cola 2L",
		IsActive:  true,
		Quantity:  mustNumeric("8.000"),
		CreatedAt: mustTime("2026-07-15T10:00:00Z"),
		UpdatedAt: mustTime("2026-07-15T10:00:00Z"),
	}
}

func increaseRowFixture(previous, current string) database.IncreaseInventoryRow {
	return database.IncreaseInventoryRow{
		ProductID:        productFixture(true).ID,
		PreviousQuantity: mustNumeric(previous),
		CurrentQuantity:  mustNumeric(current),
		CreatedAt:        mustTime("2026-07-15T10:00:00Z"),
		UpdatedAt:        mustTime("2026-07-15T10:00:00Z"),
	}
}

func decreaseRowFixture(previous, current string) database.DecreaseInventoryRow {
	return database.DecreaseInventoryRow{
		ProductID:        productFixture(true).ID,
		PreviousQuantity: mustNumeric(previous),
		CurrentQuantity:  mustNumeric(current),
		CreatedAt:        mustTime("2026-07-15T10:00:00Z"),
		UpdatedAt:        mustTime("2026-07-15T10:00:00Z"),
	}
}

func movementFixture(movementType database.InventoryMovementType, quantity, previous, current string) database.InventoryMovement {
	return database.InventoryMovement{
		ID:               mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8aa"),
		ProductID:        productFixture(true).ID,
		MovementType:     movementType,
		Quantity:         mustNumeric(quantity),
		PreviousQuantity: mustNumeric(previous),
		CurrentQuantity:  mustNumeric(current),
		Reason:           pgtype.Text{String: "Compra de fornecedor", Valid: true},
		ReferenceType:    "purchase",
		ReferenceID:      productFixture(true).ID,
		CreatedAt:        mustTime("2026-07-15T10:00:00Z"),
	}
}

func requireValidationField(t *testing.T, err error, field string) {
	t.Helper()

	var validationErr *ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
	if validationErr.Field != field {
		t.Fatalf("expected field %q, got %q", field, validationErr.Field)
	}
}

func strPtr(value string) *string {
	return &value
}

func mustUUID(value string) pgtype.UUID {
	var id pgtype.UUID
	if err := id.Scan(value); err != nil {
		panic(err)
	}

	return id
}

func mustNumeric(value string) pgtype.Numeric {
	var n pgtype.Numeric
	if err := n.ScanScientific(value); err != nil {
		panic(err)
	}

	return n
}

func mustTime(value string) pgtype.Timestamptz {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		panic(err)
	}

	return pgtype.Timestamptz{Time: parsed, Valid: true}
}
