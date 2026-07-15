package sales

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestCreateSale(t *testing.T) {
	t.Run("creates open sale", func(t *testing.T) {
		tx := &fakeTxQueries{
			createSaleFn: func(context.Context, string) (database.CreateSaleRow, error) {
				return saleRowFixture(database.SaleStatusOPEN).create(), nil
			},
			listSaleItemsBySaleIDFn: func(context.Context, pgtype.UUID) ([]database.SaleItem, error) {
				return nil, nil
			},
		}

		txManager := &fakeTxManager{tx: tx}
		svc := NewService(&fakeReadStore{}, txManager)

		resp, err := svc.Create(context.Background(), CreateSaleInput{IdempotencyKey: "sale-1"})
		if err != nil {
			t.Fatalf("Create returned error: %v", err)
		}

		if !txManager.committed || txManager.rolledBack {
			t.Fatalf("unexpected tx state: committed=%v rolledBack=%v", txManager.committed, txManager.rolledBack)
		}
		if resp.Status != string(database.SaleStatusOPEN) || resp.Subtotal != "0.00" || resp.Total != "0.00" {
			t.Fatalf("unexpected sale response: %+v", resp)
		}
		if resp.IdempotencyKey != "sale-1" {
			t.Fatalf("unexpected idempotency key: %s", resp.IdempotencyKey)
		}
		if len(resp.Items) != 0 {
			t.Fatalf("expected empty items, got %+v", resp.Items)
		}
	})

	t.Run("persistence error", func(t *testing.T) {
		tx := &fakeTxQueries{
			createSaleFn: func(context.Context, string) (database.CreateSaleRow, error) {
				return database.CreateSaleRow{}, errors.New("insert failed")
			},
		}
		txManager := &fakeTxManager{tx: tx}
		svc := NewService(&fakeReadStore{}, txManager)

		_, err := svc.Create(context.Background(), CreateSaleInput{IdempotencyKey: "sale-1"})
		if err == nil {
			t.Fatalf("expected error")
		}
		if !txManager.rolledBack || txManager.committed {
			t.Fatalf("expected rollback without commit, got committed=%v rolledBack=%v", txManager.committed, txManager.rolledBack)
		}
	})

	t.Run("mapper invalid", func(t *testing.T) {
		tx := &fakeTxQueries{
			createSaleFn: func(context.Context, string) (database.CreateSaleRow, error) {
				row := saleRowFixture(database.SaleStatusOPEN).create()
				row.Subtotal = pgtype.Numeric{}
				return row, nil
			},
			listSaleItemsBySaleIDFn: func(context.Context, pgtype.UUID) ([]database.SaleItem, error) {
				return nil, nil
			},
		}
		txManager := &fakeTxManager{tx: tx}
		svc := NewService(&fakeReadStore{}, txManager)

		_, err := svc.Create(context.Background(), CreateSaleInput{IdempotencyKey: "sale-1"})
		if err == nil {
			t.Fatalf("expected error")
		}
		if !txManager.rolledBack || txManager.committed {
			t.Fatalf("expected rollback without commit, got committed=%v rolledBack=%v", txManager.committed, txManager.rolledBack)
		}
	})
}

func TestListSales(t *testing.T) {
	t.Run("default pagination", func(t *testing.T) {
		var capturedCount database.NullSaleStatus
		var capturedList database.ListSalesParams

		svc := NewService(&fakeReadStore{
			countSalesFn: func(_ context.Context, arg database.NullSaleStatus) (int64, error) {
				capturedCount = arg
				return 1, nil
			},
			listSalesFn: func(_ context.Context, arg database.ListSalesParams) ([]database.ListSalesRow, error) {
				capturedList = arg
				return []database.ListSalesRow{saleRowFixture(database.SaleStatusOPEN).list()}, nil
			},
		}, &fakeTxManager{})

		resp, err := svc.List(context.Background(), ListSalesInput{})
		if err != nil {
			t.Fatalf("List returned error: %v", err)
		}

		if resp.Pagination.Page != 1 || resp.Pagination.PageSize != 20 || resp.Pagination.Total != 1 || resp.Pagination.TotalPages != 1 {
			t.Fatalf("unexpected pagination: %+v", resp.Pagination)
		}
		if capturedList.PageOffset != 0 || capturedList.PageSize != 20 {
			t.Fatalf("unexpected list params: %+v", capturedList)
		}
		if capturedCount.Valid {
			t.Fatalf("expected empty status filter, got %+v", capturedCount)
		}
		if len(resp.Data) != 1 || resp.Data[0].Status != string(database.SaleStatusOPEN) {
			t.Fatalf("unexpected result: %+v", resp.Data)
		}
	})

	t.Run("rejects page less than 1", func(t *testing.T) {
		page := 0
		_, err := NewService(&fakeReadStore{}, &fakeTxManager{}).List(context.Background(), ListSalesInput{Page: &page})
		requireValidationField(t, err, "page")
	})

	t.Run("rejects page size greater than 100", func(t *testing.T) {
		pageSize := 101
		_, err := NewService(&fakeReadStore{}, &fakeTxManager{}).List(context.Background(), ListSalesInput{PageSize: &pageSize})
		requireValidationField(t, err, "pageSize")
	})

	t.Run("rejects invalid status", func(t *testing.T) {
		_, err := NewService(&fakeReadStore{}, &fakeTxManager{}).List(context.Background(), ListSalesInput{Status: "wrong"})
		requireValidationField(t, err, "status")
	})

	t.Run("propagates list error", func(t *testing.T) {
		svc := NewService(&fakeReadStore{
			countSalesFn: func(context.Context, database.NullSaleStatus) (int64, error) {
				return 0, errors.New("count failed")
			},
		}, &fakeTxManager{})

		_, err := svc.List(context.Background(), ListSalesInput{})
		if err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("propagates mapper error", func(t *testing.T) {
		svc := NewService(&fakeReadStore{
			countSalesFn: func(context.Context, database.NullSaleStatus) (int64, error) {
				return 1, nil
			},
			listSalesFn: func(context.Context, database.ListSalesParams) ([]database.ListSalesRow, error) {
				row := saleRowFixture(database.SaleStatusOPEN).list()
				row.Subtotal = pgtype.Numeric{}
				return []database.ListSalesRow{row}, nil
			},
		}, &fakeTxManager{})

		_, err := svc.List(context.Background(), ListSalesInput{})
		if err == nil {
			t.Fatalf("expected error")
		}
	})
}

func TestGetSale(t *testing.T) {
	t.Run("gets sale with items", func(t *testing.T) {
		item := saleItemFixture(mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8aa"), "1.000", "0.00", "12.90")
		svc := NewService(&fakeReadStore{
			getSaleByIDFn: func(context.Context, pgtype.UUID) (database.GetSaleByIDRow, error) {
				return saleWithItemsFixture(database.SaleStatusOPEN).get(), nil
			},
			listSaleItemsBySaleIDFn: func(context.Context, pgtype.UUID) ([]database.SaleItem, error) {
				return []database.SaleItem{item}, nil
			},
		}, &fakeTxManager{})

		resp, err := svc.Get(context.Background(), saleWithItemsFixture(database.SaleStatusOPEN).ID.String())
		if err != nil {
			t.Fatalf("Get returned error: %v", err)
		}
		if len(resp.Items) != 1 || resp.Items[0].ProductSKU != item.ProductSKU {
			t.Fatalf("unexpected sale response: %+v", resp)
		}
	})

	t.Run("not found", func(t *testing.T) {
		svc := NewService(&fakeReadStore{
			getSaleByIDFn: func(context.Context, pgtype.UUID) (database.GetSaleByIDRow, error) {
				return database.GetSaleByIDRow{}, pgx.ErrNoRows
			},
		}, &fakeTxManager{})

		_, err := svc.Get(context.Background(), saleWithItemsFixture(database.SaleStatusOPEN).ID.String())
		if !errors.Is(err, ErrSaleNotFound) {
			t.Fatalf("expected ErrSaleNotFound, got %v", err)
		}
	})

	t.Run("mapper error", func(t *testing.T) {
		svc := NewService(&fakeReadStore{
			getSaleByIDFn: func(context.Context, pgtype.UUID) (database.GetSaleByIDRow, error) {
				row := saleWithItemsFixture(database.SaleStatusOPEN).get()
				row.Subtotal = pgtype.Numeric{}
				return row, nil
			},
			listSaleItemsBySaleIDFn: func(context.Context, pgtype.UUID) ([]database.SaleItem, error) {
				return nil, nil
			},
		}, &fakeTxManager{})

		_, err := svc.Get(context.Background(), saleWithItemsFixture(database.SaleStatusOPEN).ID.String())
		if err == nil {
			t.Fatalf("expected error")
		}
	})
}

func TestAddItem(t *testing.T) {
	t.Run("adds item valid", func(t *testing.T) {
		var capturedCreate database.CreateSaleItemParams
		var capturedRecalc database.RecalculateSaleTotalsParams
		item := saleItemFixture(mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8aa"), "2.500", "2.25", "30.00")

		tx := &fakeTxQueries{
			lockSaleByIDFn: func(context.Context, pgtype.UUID) (database.LockSaleByIDRow, error) {
				return saleWithItemsFixture(database.SaleStatusOPEN).lock(), nil
			},
			getProductByIDFn: func(context.Context, pgtype.UUID) (database.Product, error) {
				return productFixture(true), nil
			},
			createSaleItemFn: func(_ context.Context, arg database.CreateSaleItemParams) (database.SaleItem, error) {
				capturedCreate = arg
				return item, nil
			},
			listSaleItemsBySaleIDFn: func(context.Context, pgtype.UUID) ([]database.SaleItem, error) {
				return []database.SaleItem{item}, nil
			},
			recalculateSaleTotalsFn: func(_ context.Context, arg database.RecalculateSaleTotalsParams) (database.RecalculateSaleTotalsRow, error) {
				capturedRecalc = arg
				return saleWithItemsFixture(database.SaleStatusOPEN).recalc(), nil
			},
		}

		txManager := &fakeTxManager{tx: tx}
		svc := NewService(&fakeReadStore{}, txManager)

		resp, err := svc.AddItem(context.Background(), saleWithItemsFixture(database.SaleStatusOPEN).ID.String(), AddSaleItemInput{
			ProductID: productFixture(true).ID.String(),
			Quantity:  "2.500",
			Discount:  strPtr("2.25"),
		})
		if err != nil {
			t.Fatalf("AddItem returned error: %v", err)
		}

		if !txManager.committed || txManager.rolledBack {
			t.Fatalf("unexpected tx state: committed=%v rolledBack=%v", txManager.committed, txManager.rolledBack)
		}
		if capturedCreate.ProductName != productFixture(true).Name || capturedCreate.ProductSKU != productFixture(true).SKU {
			t.Fatalf("unexpected item snapshot: %+v", capturedCreate)
		}
		unitPrice, err := numericToMoneyString(capturedCreate.UnitPrice)
		if err != nil {
			t.Fatalf("format unit price: %v", err)
		}
		expectedUnitPrice, err := numericToMoneyString(productFixture(true).Price)
		if err != nil {
			t.Fatalf("format expected unit price: %v", err)
		}
		if unitPrice != expectedUnitPrice {
			t.Fatalf("unexpected unit price: %s", unitPrice)
		}
		subtotal, err := numericToMoneyString(capturedRecalc.Subtotal)
		if err != nil {
			t.Fatalf("format subtotal: %v", err)
		}
		discount, err := numericToMoneyString(capturedRecalc.Discount)
		if err != nil {
			t.Fatalf("format discount: %v", err)
		}
		total, err := numericToMoneyString(capturedRecalc.Total)
		if err != nil {
			t.Fatalf("format total: %v", err)
		}
		if subtotal != "32.25" || discount != "2.25" || total != "30.00" {
			t.Fatalf("unexpected recalc: %+v", capturedRecalc)
		}
		if len(resp.Items) != 1 || resp.Items[0].Total != "30.00" {
			t.Fatalf("unexpected response: %+v", resp)
		}
	})

	t.Run("product not found", func(t *testing.T) {
		tx := &fakeTxQueries{
			lockSaleByIDFn: func(context.Context, pgtype.UUID) (database.LockSaleByIDRow, error) {
				return saleWithItemsFixture(database.SaleStatusOPEN).lock(), nil
			},
			getProductByIDFn: func(context.Context, pgtype.UUID) (database.Product, error) {
				return database.Product{}, pgx.ErrNoRows
			},
		}
		_, err := NewService(&fakeReadStore{}, &fakeTxManager{tx: tx}).AddItem(context.Background(), saleWithItemsFixture(database.SaleStatusOPEN).ID.String(), AddSaleItemInput{
			ProductID: productFixture(true).ID.String(),
			Quantity:  "1.000",
		})
		if !errors.Is(err, ErrProductNotFound) {
			t.Fatalf("expected ErrProductNotFound, got %v", err)
		}
	})

	t.Run("product inactive", func(t *testing.T) {
		tx := &fakeTxQueries{
			lockSaleByIDFn: func(context.Context, pgtype.UUID) (database.LockSaleByIDRow, error) {
				return saleWithItemsFixture(database.SaleStatusOPEN).lock(), nil
			},
			getProductByIDFn: func(context.Context, pgtype.UUID) (database.Product, error) {
				return productFixture(false), nil
			},
		}
		_, err := NewService(&fakeReadStore{}, &fakeTxManager{tx: tx}).AddItem(context.Background(), saleWithItemsFixture(database.SaleStatusOPEN).ID.String(), AddSaleItemInput{
			ProductID: productFixture(true).ID.String(),
			Quantity:  "1.000",
		})
		if !errors.Is(err, ErrProductInactive) {
			t.Fatalf("expected ErrProductInactive, got %v", err)
		}
	})

	t.Run("sale not found", func(t *testing.T) {
		tx := &fakeTxQueries{
			lockSaleByIDFn: func(context.Context, pgtype.UUID) (database.LockSaleByIDRow, error) {
				return database.LockSaleByIDRow{}, pgx.ErrNoRows
			},
		}
		_, err := NewService(&fakeReadStore{}, &fakeTxManager{tx: tx}).AddItem(context.Background(), saleWithItemsFixture(database.SaleStatusOPEN).ID.String(), AddSaleItemInput{
			ProductID: productFixture(true).ID.String(),
			Quantity:  "1.000",
		})
		if !errors.Is(err, ErrSaleNotFound) {
			t.Fatalf("expected ErrSaleNotFound, got %v", err)
		}
	})

	t.Run("sale not open", func(t *testing.T) {
		tx := &fakeTxQueries{
			lockSaleByIDFn: func(context.Context, pgtype.UUID) (database.LockSaleByIDRow, error) {
				return saleWithItemsFixture(database.SaleStatusCANCELLED).lock(), nil
			},
		}
		_, err := NewService(&fakeReadStore{}, &fakeTxManager{tx: tx}).AddItem(context.Background(), saleWithItemsFixture(database.SaleStatusOPEN).ID.String(), AddSaleItemInput{
			ProductID: productFixture(true).ID.String(),
			Quantity:  "1.000",
		})
		if !errors.Is(err, ErrSaleNotOpen) {
			t.Fatalf("expected ErrSaleNotOpen, got %v", err)
		}
	})

	t.Run("quantity invalid", func(t *testing.T) {
		_, err := NewService(&fakeReadStore{}, &fakeTxManager{}).AddItem(context.Background(), saleWithItemsFixture(database.SaleStatusOPEN).ID.String(), AddSaleItemInput{
			ProductID: productFixture(true).ID.String(),
			Quantity:  "abc",
		})
		requireValidationField(t, err, "quantity")
	})

	t.Run("discount invalid", func(t *testing.T) {
		_, err := NewService(&fakeReadStore{}, &fakeTxManager{}).AddItem(context.Background(), saleWithItemsFixture(database.SaleStatusOPEN).ID.String(), AddSaleItemInput{
			ProductID: productFixture(true).ID.String(),
			Quantity:  "1.000",
			Discount:  strPtr("abc"),
		})
		requireValidationField(t, err, "discount")
	})

	t.Run("discount greater than subtotal", func(t *testing.T) {
		tx := &fakeTxQueries{
			lockSaleByIDFn: func(context.Context, pgtype.UUID) (database.LockSaleByIDRow, error) {
				return saleWithItemsFixture(database.SaleStatusOPEN).lock(), nil
			},
			getProductByIDFn: func(context.Context, pgtype.UUID) (database.Product, error) {
				return productFixture(true), nil
			},
		}
		_, err := NewService(&fakeReadStore{}, &fakeTxManager{tx: tx}).AddItem(context.Background(), saleWithItemsFixture(database.SaleStatusOPEN).ID.String(), AddSaleItemInput{
			ProductID: productFixture(true).ID.String(),
			Quantity:  "1.000",
			Discount:  strPtr("100.00"),
		})
		requireValidationField(t, err, "discount")
	})

	t.Run("create item failure rolls back", func(t *testing.T) {
		tx := &fakeTxQueries{
			lockSaleByIDFn: func(context.Context, pgtype.UUID) (database.LockSaleByIDRow, error) {
				return saleWithItemsFixture(database.SaleStatusOPEN).lock(), nil
			},
			getProductByIDFn: func(context.Context, pgtype.UUID) (database.Product, error) {
				return productFixture(true), nil
			},
			createSaleItemFn: func(context.Context, database.CreateSaleItemParams) (database.SaleItem, error) {
				return database.SaleItem{}, errors.New("insert failed")
			},
		}
		txManager := &fakeTxManager{tx: tx}
		_, err := NewService(&fakeReadStore{}, txManager).AddItem(context.Background(), saleWithItemsFixture(database.SaleStatusOPEN).ID.String(), AddSaleItemInput{
			ProductID: productFixture(true).ID.String(),
			Quantity:  "1.000",
		})
		if err == nil {
			t.Fatalf("expected error")
		}
		if !txManager.rolledBack || txManager.committed {
			t.Fatalf("expected rollback")
		}
	})

	t.Run("recalc failure rolls back", func(t *testing.T) {
		item := saleItemFixture(mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8aa"), "1.000", "0.00", "12.90")
		tx := &fakeTxQueries{
			lockSaleByIDFn: func(context.Context, pgtype.UUID) (database.LockSaleByIDRow, error) {
				return saleWithItemsFixture(database.SaleStatusOPEN).lock(), nil
			},
			getProductByIDFn: func(context.Context, pgtype.UUID) (database.Product, error) {
				return productFixture(true), nil
			},
			createSaleItemFn: func(context.Context, database.CreateSaleItemParams) (database.SaleItem, error) {
				return item, nil
			},
			listSaleItemsBySaleIDFn: func(context.Context, pgtype.UUID) ([]database.SaleItem, error) {
				return []database.SaleItem{item}, nil
			},
			recalculateSaleTotalsFn: func(context.Context, database.RecalculateSaleTotalsParams) (database.RecalculateSaleTotalsRow, error) {
				return database.RecalculateSaleTotalsRow{}, errors.New("recalc failed")
			},
		}
		txManager := &fakeTxManager{tx: tx}
		_, err := NewService(&fakeReadStore{}, txManager).AddItem(context.Background(), saleWithItemsFixture(database.SaleStatusOPEN).ID.String(), AddSaleItemInput{
			ProductID: productFixture(true).ID.String(),
			Quantity:  "1.000",
		})
		if err == nil {
			t.Fatalf("expected error")
		}
		if !txManager.rolledBack || txManager.committed {
			t.Fatalf("expected rollback")
		}
	})
}

func TestUpdateItem(t *testing.T) {
	t.Run("updates quantity and discount", func(t *testing.T) {
		var capturedUpdate database.UpdateSaleItemParams
		item := saleItemFixture(mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8aa"), "1.000", "0.00", "12.90")
		updatedItem := saleItemFixture(mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8aa"), "2.000", "1.00", "24.80")

		tx := &fakeTxQueries{
			lockSaleByIDFn: func(context.Context, pgtype.UUID) (database.LockSaleByIDRow, error) {
				return saleWithItemsFixture(database.SaleStatusOPEN).lock(), nil
			},
			getSaleItemByIDFn: func(context.Context, database.GetSaleItemByIDParams) (database.SaleItem, error) {
				return item, nil
			},
			updateSaleItemFn: func(_ context.Context, arg database.UpdateSaleItemParams) (database.SaleItem, error) {
				capturedUpdate = arg
				return updatedItem, nil
			},
			listSaleItemsBySaleIDFn: func(context.Context, pgtype.UUID) ([]database.SaleItem, error) {
				return []database.SaleItem{updatedItem}, nil
			},
			recalculateSaleTotalsFn: func(context.Context, database.RecalculateSaleTotalsParams) (database.RecalculateSaleTotalsRow, error) {
				return saleWithItemsFixture(database.SaleStatusOPEN).recalc(), nil
			},
		}

		txManager := &fakeTxManager{tx: tx}
		resp, err := NewService(&fakeReadStore{}, txManager).UpdateItem(context.Background(), saleWithItemsFixture(database.SaleStatusOPEN).ID.String(), item.ID.String(), UpdateSaleItemInput{
			Quantity: "2.000",
			Discount: strPtr("1.00"),
		})
		if err != nil {
			t.Fatalf("UpdateItem returned error: %v", err)
		}
		quantity, err := numericToQuantityString(capturedUpdate.Quantity)
		if err != nil {
			t.Fatalf("format quantity: %v", err)
		}
		discount, err := numericToMoneyString(capturedUpdate.Discount)
		if err != nil {
			t.Fatalf("format discount: %v", err)
		}
		total, err := numericToMoneyString(capturedUpdate.Total)
		if err != nil {
			t.Fatalf("format total: %v", err)
		}
		if quantity != "2.000" || discount != "1.00" {
			t.Fatalf("unexpected update params: %+v", capturedUpdate)
		}
		if total != "24.80" {
			t.Fatalf("unexpected update total: %+v", capturedUpdate)
		}
		if len(resp.Items) != 1 || resp.Items[0].Total != "24.80" {
			t.Fatalf("unexpected response: %+v", resp)
		}
	})

	t.Run("item not found", func(t *testing.T) {
		tx := &fakeTxQueries{
			lockSaleByIDFn: func(context.Context, pgtype.UUID) (database.LockSaleByIDRow, error) {
				return saleWithItemsFixture(database.SaleStatusOPEN).lock(), nil
			},
			getSaleItemByIDFn: func(context.Context, database.GetSaleItemByIDParams) (database.SaleItem, error) {
				return database.SaleItem{}, pgx.ErrNoRows
			},
		}
		_, err := NewService(&fakeReadStore{}, &fakeTxManager{tx: tx}).UpdateItem(context.Background(), saleWithItemsFixture(database.SaleStatusOPEN).ID.String(), mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8aa").String(), UpdateSaleItemInput{
			Quantity: "2.000",
		})
		if !errors.Is(err, ErrSaleItemNotFound) {
			t.Fatalf("expected ErrSaleItemNotFound, got %v", err)
		}
	})

	t.Run("item on another sale not found", func(t *testing.T) {
		tx := &fakeTxQueries{
			lockSaleByIDFn: func(context.Context, pgtype.UUID) (database.LockSaleByIDRow, error) {
				return saleWithItemsFixture(database.SaleStatusOPEN).lock(), nil
			},
			getSaleItemByIDFn: func(context.Context, database.GetSaleItemByIDParams) (database.SaleItem, error) {
				return database.SaleItem{}, pgx.ErrNoRows
			},
		}
		_, err := NewService(&fakeReadStore{}, &fakeTxManager{tx: tx}).UpdateItem(context.Background(), saleWithItemsFixture(database.SaleStatusOPEN).ID.String(), mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8ab").String(), UpdateSaleItemInput{
			Quantity: "2.000",
		})
		if !errors.Is(err, ErrSaleItemNotFound) {
			t.Fatalf("expected ErrSaleItemNotFound, got %v", err)
		}
	})

	t.Run("sale not open", func(t *testing.T) {
		tx := &fakeTxQueries{
			lockSaleByIDFn: func(context.Context, pgtype.UUID) (database.LockSaleByIDRow, error) {
				return saleWithItemsFixture(database.SaleStatusCOMPLETED).lock(), nil
			},
		}
		_, err := NewService(&fakeReadStore{}, &fakeTxManager{tx: tx}).UpdateItem(context.Background(), saleWithItemsFixture(database.SaleStatusOPEN).ID.String(), mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8aa").String(), UpdateSaleItemInput{
			Quantity: "2.000",
		})
		if !errors.Is(err, ErrSaleNotOpen) {
			t.Fatalf("expected ErrSaleNotOpen, got %v", err)
		}
	})

	t.Run("discount greater than subtotal", func(t *testing.T) {
		item := saleItemFixture(mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8aa"), "1.000", "0.00", "12.90")
		tx := &fakeTxQueries{
			lockSaleByIDFn: func(context.Context, pgtype.UUID) (database.LockSaleByIDRow, error) {
				return saleWithItemsFixture(database.SaleStatusOPEN).lock(), nil
			},
			getSaleItemByIDFn: func(context.Context, database.GetSaleItemByIDParams) (database.SaleItem, error) {
				return item, nil
			},
		}
		_, err := NewService(&fakeReadStore{}, &fakeTxManager{tx: tx}).UpdateItem(context.Background(), saleWithItemsFixture(database.SaleStatusOPEN).ID.String(), item.ID.String(), UpdateSaleItemInput{
			Quantity: "1.000",
			Discount: strPtr("100.00"),
		})
		requireValidationField(t, err, "discount")
	})

	t.Run("rollback on recalc error", func(t *testing.T) {
		item := saleItemFixture(mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8aa"), "1.000", "0.00", "12.90")
		updatedItem := saleItemFixture(mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8aa"), "2.000", "0.00", "25.80")
		tx := &fakeTxQueries{
			lockSaleByIDFn: func(context.Context, pgtype.UUID) (database.LockSaleByIDRow, error) {
				return saleWithItemsFixture(database.SaleStatusOPEN).lock(), nil
			},
			getSaleItemByIDFn: func(context.Context, database.GetSaleItemByIDParams) (database.SaleItem, error) {
				return item, nil
			},
			updateSaleItemFn: func(context.Context, database.UpdateSaleItemParams) (database.SaleItem, error) {
				return updatedItem, nil
			},
			listSaleItemsBySaleIDFn: func(context.Context, pgtype.UUID) ([]database.SaleItem, error) {
				return []database.SaleItem{updatedItem}, nil
			},
			recalculateSaleTotalsFn: func(context.Context, database.RecalculateSaleTotalsParams) (database.RecalculateSaleTotalsRow, error) {
				return database.RecalculateSaleTotalsRow{}, errors.New("recalc failed")
			},
		}
		txManager := &fakeTxManager{tx: tx}
		_, err := NewService(&fakeReadStore{}, txManager).UpdateItem(context.Background(), saleWithItemsFixture(database.SaleStatusOPEN).ID.String(), item.ID.String(), UpdateSaleItemInput{
			Quantity: "2.000",
		})
		if err == nil {
			t.Fatalf("expected error")
		}
		if !txManager.rolledBack || txManager.committed {
			t.Fatalf("expected rollback")
		}
	})
}

func TestRemoveItem(t *testing.T) {
	t.Run("removes item", func(t *testing.T) {
		var capturedDelete database.DeleteSaleItemParams
		item := saleItemFixture(mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8aa"), "1.000", "0.00", "12.90")

		tx := &fakeTxQueries{
			lockSaleByIDFn: func(context.Context, pgtype.UUID) (database.LockSaleByIDRow, error) {
				return saleWithItemsFixture(database.SaleStatusOPEN).lock(), nil
			},
			getSaleItemByIDFn: func(context.Context, database.GetSaleItemByIDParams) (database.SaleItem, error) {
				return item, nil
			},
			deleteSaleItemFn: func(_ context.Context, arg database.DeleteSaleItemParams) (database.SaleItem, error) {
				capturedDelete = arg
				return item, nil
			},
			listSaleItemsBySaleIDFn: func(context.Context, pgtype.UUID) ([]database.SaleItem, error) {
				return nil, nil
			},
			recalculateSaleTotalsFn: func(context.Context, database.RecalculateSaleTotalsParams) (database.RecalculateSaleTotalsRow, error) {
				return saleWithItemsFixture(database.SaleStatusOPEN).recalc(), nil
			},
		}

		txManager := &fakeTxManager{tx: tx}
		resp, err := NewService(&fakeReadStore{}, txManager).RemoveItem(context.Background(), saleWithItemsFixture(database.SaleStatusOPEN).ID.String(), item.ID.String())
		if err != nil {
			t.Fatalf("RemoveItem returned error: %v", err)
		}
		if capturedDelete.ID.String() != item.ID.String() {
			t.Fatalf("unexpected delete params: %+v", capturedDelete)
		}
		if len(resp.Items) != 0 {
			t.Fatalf("expected empty items, got %+v", resp.Items)
		}
	})

	t.Run("item not found", func(t *testing.T) {
		tx := &fakeTxQueries{
			lockSaleByIDFn: func(context.Context, pgtype.UUID) (database.LockSaleByIDRow, error) {
				return saleWithItemsFixture(database.SaleStatusOPEN).lock(), nil
			},
			getSaleItemByIDFn: func(context.Context, database.GetSaleItemByIDParams) (database.SaleItem, error) {
				return database.SaleItem{}, pgx.ErrNoRows
			},
		}
		_, err := NewService(&fakeReadStore{}, &fakeTxManager{tx: tx}).RemoveItem(context.Background(), saleWithItemsFixture(database.SaleStatusOPEN).ID.String(), mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8aa").String())
		if !errors.Is(err, ErrSaleItemNotFound) {
			t.Fatalf("expected ErrSaleItemNotFound, got %v", err)
		}
	})

	t.Run("sale not open", func(t *testing.T) {
		tx := &fakeTxQueries{
			lockSaleByIDFn: func(context.Context, pgtype.UUID) (database.LockSaleByIDRow, error) {
				return saleWithItemsFixture(database.SaleStatusCANCELLED).lock(), nil
			},
		}
		_, err := NewService(&fakeReadStore{}, &fakeTxManager{tx: tx}).RemoveItem(context.Background(), saleWithItemsFixture(database.SaleStatusOPEN).ID.String(), mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8aa").String())
		if !errors.Is(err, ErrSaleNotOpen) {
			t.Fatalf("expected ErrSaleNotOpen, got %v", err)
		}
	})

	t.Run("rollback on delete error", func(t *testing.T) {
		item := saleItemFixture(mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8aa"), "1.000", "0.00", "12.90")
		tx := &fakeTxQueries{
			lockSaleByIDFn: func(context.Context, pgtype.UUID) (database.LockSaleByIDRow, error) {
				return saleWithItemsFixture(database.SaleStatusOPEN).lock(), nil
			},
			getSaleItemByIDFn: func(context.Context, database.GetSaleItemByIDParams) (database.SaleItem, error) {
				return item, nil
			},
			deleteSaleItemFn: func(context.Context, database.DeleteSaleItemParams) (database.SaleItem, error) {
				return database.SaleItem{}, errors.New("delete failed")
			},
		}
		txManager := &fakeTxManager{tx: tx}
		_, err := NewService(&fakeReadStore{}, txManager).RemoveItem(context.Background(), saleWithItemsFixture(database.SaleStatusOPEN).ID.String(), item.ID.String())
		if err == nil {
			t.Fatalf("expected error")
		}
		if !txManager.rolledBack || txManager.committed {
			t.Fatalf("expected rollback")
		}
	})
}

func TestCancelSale(t *testing.T) {
	t.Run("cancels open sale", func(t *testing.T) {
		tx := &fakeTxQueries{
			lockSaleByIDFn: func(context.Context, pgtype.UUID) (database.LockSaleByIDRow, error) {
				return saleWithItemsFixture(database.SaleStatusOPEN).lock(), nil
			},
			cancelSaleFn: func(context.Context, pgtype.UUID) (database.CancelSaleRow, error) {
				return saleWithItemsFixture(database.SaleStatusCANCELLED).cancel(), nil
			},
			listSaleItemsBySaleIDFn: func(context.Context, pgtype.UUID) ([]database.SaleItem, error) {
				return nil, nil
			},
		}
		txManager := &fakeTxManager{tx: tx}
		resp, err := NewService(&fakeReadStore{}, txManager).Cancel(context.Background(), saleWithItemsFixture(database.SaleStatusOPEN).ID.String())
		if err != nil {
			t.Fatalf("Cancel returned error: %v", err)
		}
		if resp.Status != string(database.SaleStatusCANCELLED) {
			t.Fatalf("unexpected status: %+v", resp)
		}
		if !txManager.committed || txManager.rolledBack {
			t.Fatalf("unexpected tx state: committed=%v rolledBack=%v", txManager.committed, txManager.rolledBack)
		}
	})

	t.Run("sale not found", func(t *testing.T) {
		tx := &fakeTxQueries{
			lockSaleByIDFn: func(context.Context, pgtype.UUID) (database.LockSaleByIDRow, error) {
				return database.LockSaleByIDRow{}, pgx.ErrNoRows
			},
		}
		_, err := NewService(&fakeReadStore{}, &fakeTxManager{tx: tx}).Cancel(context.Background(), saleWithItemsFixture(database.SaleStatusOPEN).ID.String())
		if !errors.Is(err, ErrSaleNotFound) {
			t.Fatalf("expected ErrSaleNotFound, got %v", err)
		}
	})

	t.Run("already cancelled", func(t *testing.T) {
		tx := &fakeTxQueries{
			lockSaleByIDFn: func(context.Context, pgtype.UUID) (database.LockSaleByIDRow, error) {
				return saleWithItemsFixture(database.SaleStatusCANCELLED).lock(), nil
			},
		}
		_, err := NewService(&fakeReadStore{}, &fakeTxManager{tx: tx}).Cancel(context.Background(), saleWithItemsFixture(database.SaleStatusOPEN).ID.String())
		if !errors.Is(err, ErrSaleAlreadyCancelled) {
			t.Fatalf("expected ErrSaleAlreadyCancelled, got %v", err)
		}
	})

	t.Run("completed sale not open", func(t *testing.T) {
		tx := &fakeTxQueries{
			lockSaleByIDFn: func(context.Context, pgtype.UUID) (database.LockSaleByIDRow, error) {
				return saleWithItemsFixture(database.SaleStatusCOMPLETED).lock(), nil
			},
		}
		_, err := NewService(&fakeReadStore{}, &fakeTxManager{tx: tx}).Cancel(context.Background(), saleWithItemsFixture(database.SaleStatusOPEN).ID.String())
		if !errors.Is(err, ErrSaleNotOpen) {
			t.Fatalf("expected ErrSaleNotOpen, got %v", err)
		}
	})

	t.Run("rollback on persistence error", func(t *testing.T) {
		tx := &fakeTxQueries{
			lockSaleByIDFn: func(context.Context, pgtype.UUID) (database.LockSaleByIDRow, error) {
				return saleWithItemsFixture(database.SaleStatusOPEN).lock(), nil
			},
			cancelSaleFn: func(context.Context, pgtype.UUID) (database.CancelSaleRow, error) {
				return database.CancelSaleRow{}, errors.New("cancel failed")
			},
		}
		txManager := &fakeTxManager{tx: tx}
		_, err := NewService(&fakeReadStore{}, txManager).Cancel(context.Background(), saleWithItemsFixture(database.SaleStatusOPEN).ID.String())
		if err == nil {
			t.Fatalf("expected error")
		}
		if !txManager.rolledBack || txManager.committed {
			t.Fatalf("expected rollback")
		}
	})
}

func TestQuantityAndMoneyValidation(t *testing.T) {
	svc := NewService(&fakeReadStore{}, &fakeTxManager{})

	_, err := svc.AddItem(context.Background(), saleWithItemsFixture(database.SaleStatusOPEN).ID.String(), AddSaleItemInput{
		ProductID: productFixture(true).ID.String(),
		Quantity:  "0",
	})
	requireValidationField(t, err, "quantity")

	_, err = svc.AddItem(context.Background(), saleWithItemsFixture(database.SaleStatusOPEN).ID.String(), AddSaleItemInput{
		ProductID: productFixture(true).ID.String(),
		Quantity:  "1.000",
		Discount:  strPtr("-1.00"),
	})
	requireValidationField(t, err, "discount")
}

type fakeReadStore struct {
	getSaleByIDFn           func(context.Context, pgtype.UUID) (database.GetSaleByIDRow, error)
	listSalesFn             func(context.Context, database.ListSalesParams) ([]database.ListSalesRow, error)
	countSalesFn            func(context.Context, database.NullSaleStatus) (int64, error)
	listSaleItemsBySaleIDFn func(context.Context, pgtype.UUID) ([]database.SaleItem, error)
}

func (f *fakeReadStore) GetSaleByID(ctx context.Context, id pgtype.UUID) (database.GetSaleByIDRow, error) {
	if f.getSaleByIDFn == nil {
		panic("unexpected GetSaleByID call")
	}
	return f.getSaleByIDFn(ctx, id)
}

func (f *fakeReadStore) ListSales(ctx context.Context, arg database.ListSalesParams) ([]database.ListSalesRow, error) {
	if f.listSalesFn == nil {
		panic("unexpected ListSales call")
	}
	return f.listSalesFn(ctx, arg)
}

func (f *fakeReadStore) CountSales(ctx context.Context, arg database.NullSaleStatus) (int64, error) {
	if f.countSalesFn == nil {
		panic("unexpected CountSales call")
	}
	return f.countSalesFn(ctx, arg)
}

func (f *fakeReadStore) ListSaleItemsBySaleID(ctx context.Context, saleID pgtype.UUID) ([]database.SaleItem, error) {
	if f.listSaleItemsBySaleIDFn == nil {
		panic("unexpected ListSaleItemsBySaleID call")
	}
	return f.listSaleItemsBySaleIDFn(ctx, saleID)
}

type fakeTxQueries struct {
	getSaleByIDFn           func(context.Context, pgtype.UUID) (database.GetSaleByIDRow, error)
	lockSaleByIDFn          func(context.Context, pgtype.UUID) (database.LockSaleByIDRow, error)
	getProductByIDFn        func(context.Context, pgtype.UUID) (database.Product, error)
	getSaleItemByIDFn       func(context.Context, database.GetSaleItemByIDParams) (database.SaleItem, error)
	createSaleFn            func(context.Context, string) (database.CreateSaleRow, error)
	createSaleItemFn        func(context.Context, database.CreateSaleItemParams) (database.SaleItem, error)
	updateSaleItemFn        func(context.Context, database.UpdateSaleItemParams) (database.SaleItem, error)
	deleteSaleItemFn        func(context.Context, database.DeleteSaleItemParams) (database.SaleItem, error)
	recalculateSaleTotalsFn func(context.Context, database.RecalculateSaleTotalsParams) (database.RecalculateSaleTotalsRow, error)
	cancelSaleFn            func(context.Context, pgtype.UUID) (database.CancelSaleRow, error)
	listSaleItemsBySaleIDFn func(context.Context, pgtype.UUID) ([]database.SaleItem, error)
}

func (f *fakeTxQueries) GetSaleByID(ctx context.Context, id pgtype.UUID) (database.GetSaleByIDRow, error) {
	if f.getSaleByIDFn == nil {
		panic("unexpected GetSaleByID call")
	}
	return f.getSaleByIDFn(ctx, id)
}

func (f *fakeTxQueries) LockSaleByID(ctx context.Context, id pgtype.UUID) (database.LockSaleByIDRow, error) {
	if f.lockSaleByIDFn == nil {
		panic("unexpected LockSaleByID call")
	}
	return f.lockSaleByIDFn(ctx, id)
}

func (f *fakeTxQueries) GetProductByID(ctx context.Context, id pgtype.UUID) (database.Product, error) {
	if f.getProductByIDFn == nil {
		panic("unexpected GetProductByID call")
	}
	return f.getProductByIDFn(ctx, id)
}

func (f *fakeTxQueries) GetSaleItemByID(ctx context.Context, arg database.GetSaleItemByIDParams) (database.SaleItem, error) {
	if f.getSaleItemByIDFn == nil {
		panic("unexpected GetSaleItemByID call")
	}
	return f.getSaleItemByIDFn(ctx, arg)
}

func (f *fakeTxQueries) CreateSale(ctx context.Context, idempotencyKey string) (database.CreateSaleRow, error) {
	if f.createSaleFn == nil {
		panic("unexpected CreateSale call")
	}
	return f.createSaleFn(ctx, idempotencyKey)
}

func (f *fakeTxQueries) CreateSaleItem(ctx context.Context, arg database.CreateSaleItemParams) (database.SaleItem, error) {
	if f.createSaleItemFn == nil {
		panic("unexpected CreateSaleItem call")
	}
	return f.createSaleItemFn(ctx, arg)
}

func (f *fakeTxQueries) UpdateSaleItem(ctx context.Context, arg database.UpdateSaleItemParams) (database.SaleItem, error) {
	if f.updateSaleItemFn == nil {
		panic("unexpected UpdateSaleItem call")
	}
	return f.updateSaleItemFn(ctx, arg)
}

func (f *fakeTxQueries) DeleteSaleItem(ctx context.Context, arg database.DeleteSaleItemParams) (database.SaleItem, error) {
	if f.deleteSaleItemFn == nil {
		panic("unexpected DeleteSaleItem call")
	}
	return f.deleteSaleItemFn(ctx, arg)
}

func (f *fakeTxQueries) RecalculateSaleTotals(ctx context.Context, arg database.RecalculateSaleTotalsParams) (database.RecalculateSaleTotalsRow, error) {
	if f.recalculateSaleTotalsFn == nil {
		panic("unexpected RecalculateSaleTotals call")
	}
	return f.recalculateSaleTotalsFn(ctx, arg)
}

func (f *fakeTxQueries) CancelSale(ctx context.Context, id pgtype.UUID) (database.CancelSaleRow, error) {
	if f.cancelSaleFn == nil {
		panic("unexpected CancelSale call")
	}
	return f.cancelSaleFn(ctx, id)
}

func (f *fakeTxQueries) ListSaleItemsBySaleID(ctx context.Context, saleID pgtype.UUID) ([]database.SaleItem, error) {
	if f.listSaleItemsBySaleIDFn == nil {
		panic("unexpected ListSaleItemsBySaleID call")
	}
	return f.listSaleItemsBySaleIDFn(ctx, saleID)
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

type saleFixtureValues struct {
	ID             pgtype.UUID
	Number         int64
	Status         database.SaleStatus
	Subtotal       pgtype.Numeric
	Discount       pgtype.Numeric
	Addition       pgtype.Numeric
	Total          pgtype.Numeric
	OpenedAt       pgtype.Timestamptz
	CompletedAt    pgtype.Timestamptz
	CancelledAt    pgtype.Timestamptz
	CreatedAt      pgtype.Timestamptz
	UpdatedAt      pgtype.Timestamptz
	IdempotencyKey string
}

func saleRowFixture(status database.SaleStatus) saleFixtureValues {
	f := saleFixtureValues{
		ID:             mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8a1"),
		Number:         42,
		Status:         status,
		Subtotal:       mustNumeric("0.00"),
		Discount:       mustNumeric("0.00"),
		Addition:       mustNumeric("0.00"),
		Total:          mustNumeric("0.00"),
		OpenedAt:       mustTime("2026-07-15T10:00:00Z"),
		CreatedAt:      mustTime("2026-07-15T10:00:00Z"),
		UpdatedAt:      mustTime("2026-07-15T10:00:00Z"),
		IdempotencyKey: "sale-1",
	}

	switch status {
	case database.SaleStatusCOMPLETED:
		f.CompletedAt = mustTime("2026-07-15T11:00:00Z")
	case database.SaleStatusCANCELLED:
		f.CancelledAt = mustTime("2026-07-15T11:00:00Z")
	}

	return f
}

func saleWithItemsFixture(status database.SaleStatus) saleFixtureValues {
	f := saleRowFixture(status)
	f.Subtotal = mustNumeric("32.25")
	f.Discount = mustNumeric("2.25")
	f.Total = mustNumeric("30.00")
	return f
}

func (f saleFixtureValues) get() database.GetSaleByIDRow {
	return database.GetSaleByIDRow{
		ID:             f.ID,
		Number:         f.Number,
		Status:         f.Status,
		Subtotal:       f.Subtotal,
		Discount:       f.Discount,
		Addition:       f.Addition,
		Total:          f.Total,
		OpenedAt:       f.OpenedAt,
		CompletedAt:    f.CompletedAt,
		CancelledAt:    f.CancelledAt,
		CreatedAt:      f.CreatedAt,
		UpdatedAt:      f.UpdatedAt,
		IdempotencyKey: f.IdempotencyKey,
	}
}

func (f saleFixtureValues) list() database.ListSalesRow {
	return database.ListSalesRow{
		ID:             f.ID,
		Number:         f.Number,
		Status:         f.Status,
		Subtotal:       f.Subtotal,
		Discount:       f.Discount,
		Addition:       f.Addition,
		Total:          f.Total,
		OpenedAt:       f.OpenedAt,
		CompletedAt:    f.CompletedAt,
		CancelledAt:    f.CancelledAt,
		CreatedAt:      f.CreatedAt,
		UpdatedAt:      f.UpdatedAt,
		IdempotencyKey: f.IdempotencyKey,
	}
}

func (f saleFixtureValues) create() database.CreateSaleRow {
	return database.CreateSaleRow{
		ID:             f.ID,
		Number:         f.Number,
		Status:         f.Status,
		Subtotal:       f.Subtotal,
		Discount:       f.Discount,
		Addition:       f.Addition,
		Total:          f.Total,
		OpenedAt:       f.OpenedAt,
		CompletedAt:    f.CompletedAt,
		CancelledAt:    f.CancelledAt,
		CreatedAt:      f.CreatedAt,
		UpdatedAt:      f.UpdatedAt,
		IdempotencyKey: f.IdempotencyKey,
	}
}

func (f saleFixtureValues) lock() database.LockSaleByIDRow {
	return database.LockSaleByIDRow{
		ID:             f.ID,
		Number:         f.Number,
		Status:         f.Status,
		Subtotal:       f.Subtotal,
		Discount:       f.Discount,
		Addition:       f.Addition,
		Total:          f.Total,
		OpenedAt:       f.OpenedAt,
		CompletedAt:    f.CompletedAt,
		CancelledAt:    f.CancelledAt,
		CreatedAt:      f.CreatedAt,
		UpdatedAt:      f.UpdatedAt,
		IdempotencyKey: f.IdempotencyKey,
	}
}

func (f saleFixtureValues) recalc() database.RecalculateSaleTotalsRow {
	return database.RecalculateSaleTotalsRow{
		ID:             f.ID,
		Number:         f.Number,
		Status:         f.Status,
		Subtotal:       f.Subtotal,
		Discount:       f.Discount,
		Addition:       f.Addition,
		Total:          f.Total,
		OpenedAt:       f.OpenedAt,
		CompletedAt:    f.CompletedAt,
		CancelledAt:    f.CancelledAt,
		CreatedAt:      f.CreatedAt,
		UpdatedAt:      f.UpdatedAt,
		IdempotencyKey: f.IdempotencyKey,
	}
}

func (f saleFixtureValues) cancel() database.CancelSaleRow {
	return database.CancelSaleRow{
		ID:             f.ID,
		Number:         f.Number,
		Status:         f.Status,
		Subtotal:       f.Subtotal,
		Discount:       f.Discount,
		Addition:       f.Addition,
		Total:          f.Total,
		OpenedAt:       f.OpenedAt,
		CompletedAt:    f.CompletedAt,
		CancelledAt:    f.CancelledAt,
		CreatedAt:      f.CreatedAt,
		UpdatedAt:      f.UpdatedAt,
		IdempotencyKey: f.IdempotencyKey,
	}
}

func saleItemFixture(id pgtype.UUID, quantity, discount, total string) database.SaleItem {
	return database.SaleItem{
		ID:          id,
		SaleID:      saleWithItemsFixture(database.SaleStatusOPEN).ID,
		ProductID:   productFixture(true).ID,
		ProductName: productFixture(true).Name,
		ProductSKU:  productFixture(true).SKU,
		UnitPrice:   productFixture(true).Price,
		Quantity:    mustNumeric(quantity),
		Discount:    mustNumeric(discount),
		Total:       mustNumeric(total),
		CreatedAt:   mustTime("2026-07-15T10:00:00Z"),
	}
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
