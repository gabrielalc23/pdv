package sales_test

import (
	"context"
	"errors"
	"math/big"
	"testing"
	"time"

	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/gabrielalc23/pdv/internal/platform/tenancy"
	"github.com/gabrielalc23/pdv/internal/sales"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	testOrgID        = mustUUID("00000000-0000-0000-0000-000000000001")
	testStoreID      = mustUUID("00000000-0000-0000-0000-000000000002")
	testMembershipID = mustUUID("00000000-0000-0000-0000-000000000003")
	testScope        = tenancy.ActorScope{
		OrganizationID:    testOrgID,
		StoreID:           testStoreID,
		ActorMembershipID: testMembershipID,
	}
	testStoreScope = tenancy.StoreScope{
		OrganizationID: testOrgID,
		StoreID:        testStoreID,
	}
)

func TestCreateSale(t *testing.T) {
	t.Run("creates open sale", func(t *testing.T) {
		tx := &fakeTxQueries{
			createSaleFn: func(_ context.Context, scope tenancy.ActorScope, params database.CreateSaleForStoreParams) (database.CreateSaleForStoreRow, error) {
				return saleRowFixture(database.SaleStatusOPEN).create(), nil
			},
			listSaleItemsBySaleIDFn: func(_ context.Context, scope tenancy.StoreScope, saleID pgtype.UUID) ([]database.SaleItem, error) {
				return nil, nil
			},
		}

		txManager := &fakeTxManager{tx: tx}
		svc := sales.NewService(&fakeReadStore{}, txManager)

		resp, err := svc.Create(context.Background(), testScope, sales.CreateSaleInput{IdempotencyKey: "sale-1"})
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
			createSaleFn: func(_ context.Context, scope tenancy.ActorScope, params database.CreateSaleForStoreParams) (database.CreateSaleForStoreRow, error) {
				return database.CreateSaleForStoreRow{}, errors.New("insert failed")
			},
		}
		txManager := &fakeTxManager{tx: tx}
		svc := sales.NewService(&fakeReadStore{}, txManager)

		_, err := svc.Create(context.Background(), testScope, sales.CreateSaleInput{IdempotencyKey: "sale-1"})
		if err == nil {
			t.Fatalf("expected error")
		}
		if !txManager.rolledBack || txManager.committed {
			t.Fatalf("expected rollback without commit, got committed=%v rolledBack=%v", txManager.committed, txManager.rolledBack)
		}
	})

	t.Run("mapper invalid", func(t *testing.T) {
		tx := &fakeTxQueries{
			createSaleFn: func(_ context.Context, scope tenancy.ActorScope, params database.CreateSaleForStoreParams) (database.CreateSaleForStoreRow, error) {
				row := saleRowFixture(database.SaleStatusOPEN).create()
				row.Subtotal = pgtype.Numeric{}
				return row, nil
			},
			listSaleItemsBySaleIDFn: func(_ context.Context, scope tenancy.StoreScope, saleID pgtype.UUID) ([]database.SaleItem, error) {
				return nil, nil
			},
		}
		txManager := &fakeTxManager{tx: tx}
		svc := sales.NewService(&fakeReadStore{}, txManager)

		_, err := svc.Create(context.Background(), testScope, sales.CreateSaleInput{IdempotencyKey: "sale-1"})
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
		var capturedCount database.CountSalesForStoreParams
		var capturedList database.ListSalesForStoreParams

		svc := sales.NewService(&fakeReadStore{
			countSalesFn: func(_ context.Context, scope tenancy.StoreScope, arg database.CountSalesForStoreParams) (int64, error) {
				capturedCount = arg
				return 1, nil
			},
			listSalesFn: func(_ context.Context, scope tenancy.StoreScope, arg database.ListSalesForStoreParams) ([]database.Sale, error) {
				capturedList = arg
				return []database.Sale{saleRowFixture(database.SaleStatusOPEN).toSale()}, nil
			},
		}, &fakeTxManager{})

		resp, err := svc.List(context.Background(), testScope, sales.ListSalesInput{})
		if err != nil {
			t.Fatalf("List returned error: %v", err)
		}

		if resp.Pagination.Page != 1 || resp.Pagination.PageSize != 20 || resp.Pagination.Total != 1 || resp.Pagination.TotalPages != 1 {
			t.Fatalf("unexpected pagination: %+v", resp.Pagination)
		}
		if capturedList.PageOffset != 0 || capturedList.PageSize != 20 {
			t.Fatalf("unexpected list params: %+v", capturedList)
		}
		if capturedCount.Status.Valid {
			t.Fatalf("expected empty status filter, got %+v", capturedCount)
		}
		if len(resp.Data) != 1 || resp.Data[0].Status != string(database.SaleStatusOPEN) {
			t.Fatalf("unexpected result: %+v", resp.Data)
		}
	})

	t.Run("rejects page less than 1", func(t *testing.T) {
		page := 0
		_, err := sales.NewService(&fakeReadStore{}, &fakeTxManager{}).List(context.Background(), testScope, sales.ListSalesInput{Page: &page})
		requireValidationField(t, err, "page")
	})

	t.Run("rejects page size greater than 100", func(t *testing.T) {
		pageSize := 101
		_, err := sales.NewService(&fakeReadStore{}, &fakeTxManager{}).List(context.Background(), testScope, sales.ListSalesInput{PageSize: &pageSize})
		requireValidationField(t, err, "pageSize")
	})

	t.Run("rejects invalid status", func(t *testing.T) {
		_, err := sales.NewService(&fakeReadStore{}, &fakeTxManager{}).List(context.Background(), testScope, sales.ListSalesInput{Status: "wrong"})
		requireValidationField(t, err, "status")
	})

	t.Run("propagates list error", func(t *testing.T) {
		svc := sales.NewService(&fakeReadStore{
			countSalesFn: func(_ context.Context, scope tenancy.StoreScope, arg database.CountSalesForStoreParams) (int64, error) {
				return 0, errors.New("count failed")
			},
		}, &fakeTxManager{})

		_, err := svc.List(context.Background(), testScope, sales.ListSalesInput{})
		if err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("propagates mapper error", func(t *testing.T) {
		svc := sales.NewService(&fakeReadStore{
			countSalesFn: func(_ context.Context, scope tenancy.StoreScope, arg database.CountSalesForStoreParams) (int64, error) {
				return 1, nil
			},
			listSalesFn: func(_ context.Context, scope tenancy.StoreScope, arg database.ListSalesForStoreParams) ([]database.Sale, error) {
				row := saleRowFixture(database.SaleStatusOPEN).toSale()
				row.Subtotal = pgtype.Numeric{}
				return []database.Sale{row}, nil
			},
		}, &fakeTxManager{})

		_, err := svc.List(context.Background(), testScope, sales.ListSalesInput{})
		if err == nil {
			t.Fatalf("expected error")
		}
	})
}

func TestGetSale(t *testing.T) {
	t.Run("gets sale with items", func(t *testing.T) {
		item := saleItemFixture(mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8aa"), "1.000", "0.00", "12.90")
		svc := sales.NewService(&fakeReadStore{
			getSaleByIDFn: func(_ context.Context, scope tenancy.StoreScope, id pgtype.UUID) (database.Sale, error) {
				return saleWithItemsFixture(database.SaleStatusOPEN).toSale(), nil
			},
			listSaleItemsBySaleIDFn: func(_ context.Context, scope tenancy.StoreScope, saleID pgtype.UUID) ([]database.SaleItem, error) {
				return []database.SaleItem{item}, nil
			},
		}, &fakeTxManager{})

		resp, err := svc.Get(context.Background(), testScope, saleWithItemsFixture(database.SaleStatusOPEN).ID.String())
		if err != nil {
			t.Fatalf("Get returned error: %v", err)
		}
		if len(resp.Items) != 1 || resp.Items[0].ProductSKU != item.ProductSKU {
			t.Fatalf("unexpected sale response: %+v", resp)
		}
	})

	t.Run("not found", func(t *testing.T) {
		svc := sales.NewService(&fakeReadStore{
			getSaleByIDFn: func(_ context.Context, scope tenancy.StoreScope, id pgtype.UUID) (database.Sale, error) {
				return database.Sale{}, pgx.ErrNoRows
			},
		}, &fakeTxManager{})

		_, err := svc.Get(context.Background(), testScope, saleWithItemsFixture(database.SaleStatusOPEN).ID.String())
		if !errors.Is(err, sales.ErrSaleNotFound) {
			t.Fatalf("expected sales.ErrSaleNotFound, got %v", err)
		}
	})

	t.Run("mapper error", func(t *testing.T) {
		svc := sales.NewService(&fakeReadStore{
			getSaleByIDFn: func(_ context.Context, scope tenancy.StoreScope, id pgtype.UUID) (database.Sale, error) {
				row := saleWithItemsFixture(database.SaleStatusOPEN).toSale()
				row.Subtotal = pgtype.Numeric{}
				return row, nil
			},
			listSaleItemsBySaleIDFn: func(_ context.Context, scope tenancy.StoreScope, saleID pgtype.UUID) ([]database.SaleItem, error) {
				return nil, nil
			},
		}, &fakeTxManager{})

		_, err := svc.Get(context.Background(), testScope, saleWithItemsFixture(database.SaleStatusOPEN).ID.String())
		if err == nil {
			t.Fatalf("expected error")
		}
	})
}

func TestAddItem(t *testing.T) {
	t.Run("adds item valid", func(t *testing.T) {
		var capturedCreate database.CreateSaleItemForStoreParams
		var capturedRecalc database.RecalculateSaleTotalsForStoreParams
		item := saleItemFixture(mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8aa"), "2.500", "2.25", "30.00")

		tx := &fakeTxQueries{
			lockSaleByIDFn: func(_ context.Context, scope tenancy.StoreScope, id pgtype.UUID) (database.Sale, error) {
				return saleWithItemsFixture(database.SaleStatusOPEN).toSale(), nil
			},
			getProductByIDFn: func(_ context.Context, scope tenancy.StoreScope, id pgtype.UUID) (database.GetProductByIDForStoreRow, error) {
				return productFixture(true), nil
			},
			createSaleItemFn: func(_ context.Context, scope tenancy.StoreScope, arg database.CreateSaleItemForStoreParams) (database.SaleItem, error) {
				capturedCreate = arg
				return item, nil
			},
			listSaleItemsBySaleIDFn: func(_ context.Context, scope tenancy.StoreScope, saleID pgtype.UUID) ([]database.SaleItem, error) {
				return []database.SaleItem{item}, nil
			},
			recalculateSaleTotalsFn: func(_ context.Context, scope tenancy.StoreScope, arg database.RecalculateSaleTotalsForStoreParams) (database.Sale, error) {
				capturedRecalc = arg
				return saleWithItemsFixture(database.SaleStatusOPEN).toSale(), nil
			},
		}

		txManager := &fakeTxManager{tx: tx}
		svc := sales.NewService(&fakeReadStore{}, txManager)

		resp, err := svc.AddItem(context.Background(), testScope, saleWithItemsFixture(database.SaleStatusOPEN).ID.String(), sales.AddSaleItemInput{
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
			lockSaleByIDFn: func(_ context.Context, scope tenancy.StoreScope, id pgtype.UUID) (database.Sale, error) {
				return saleWithItemsFixture(database.SaleStatusOPEN).toSale(), nil
			},
			getProductByIDFn: func(_ context.Context, scope tenancy.StoreScope, id pgtype.UUID) (database.GetProductByIDForStoreRow, error) {
				return database.GetProductByIDForStoreRow{}, pgx.ErrNoRows
			},
		}
		_, err := sales.NewService(&fakeReadStore{}, &fakeTxManager{tx: tx}).AddItem(context.Background(), testScope, saleWithItemsFixture(database.SaleStatusOPEN).ID.String(), sales.AddSaleItemInput{
			ProductID: productFixture(true).ID.String(),
			Quantity:  "1.000",
		})
		if !errors.Is(err, sales.ErrProductNotFound) {
			t.Fatalf("expected sales.ErrProductNotFound, got %v", err)
		}
	})

	t.Run("product inactive", func(t *testing.T) {
		tx := &fakeTxQueries{
			lockSaleByIDFn: func(_ context.Context, scope tenancy.StoreScope, id pgtype.UUID) (database.Sale, error) {
				return saleWithItemsFixture(database.SaleStatusOPEN).toSale(), nil
			},
			getProductByIDFn: func(_ context.Context, scope tenancy.StoreScope, id pgtype.UUID) (database.GetProductByIDForStoreRow, error) {
				return productFixture(false), nil
			},
		}
		_, err := sales.NewService(&fakeReadStore{}, &fakeTxManager{tx: tx}).AddItem(context.Background(), testScope, saleWithItemsFixture(database.SaleStatusOPEN).ID.String(), sales.AddSaleItemInput{
			ProductID: productFixture(true).ID.String(),
			Quantity:  "1.000",
		})
		if !errors.Is(err, sales.ErrProductInactive) {
			t.Fatalf("expected sales.ErrProductInactive, got %v", err)
		}
	})

	t.Run("sale not found", func(t *testing.T) {
		tx := &fakeTxQueries{
			lockSaleByIDFn: func(_ context.Context, scope tenancy.StoreScope, id pgtype.UUID) (database.Sale, error) {
				return database.Sale{}, pgx.ErrNoRows
			},
		}
		_, err := sales.NewService(&fakeReadStore{}, &fakeTxManager{tx: tx}).AddItem(context.Background(), testScope, saleWithItemsFixture(database.SaleStatusOPEN).ID.String(), sales.AddSaleItemInput{
			ProductID: productFixture(true).ID.String(),
			Quantity:  "1.000",
		})
		if !errors.Is(err, sales.ErrSaleNotFound) {
			t.Fatalf("expected sales.ErrSaleNotFound, got %v", err)
		}
	})

	t.Run("sale not open", func(t *testing.T) {
		tx := &fakeTxQueries{
			lockSaleByIDFn: func(_ context.Context, scope tenancy.StoreScope, id pgtype.UUID) (database.Sale, error) {
				return saleWithItemsFixture(database.SaleStatusCANCELLED).toSale(), nil
			},
		}
		_, err := sales.NewService(&fakeReadStore{}, &fakeTxManager{tx: tx}).AddItem(context.Background(), testScope, saleWithItemsFixture(database.SaleStatusOPEN).ID.String(), sales.AddSaleItemInput{
			ProductID: productFixture(true).ID.String(),
			Quantity:  "1.000",
		})
		if !errors.Is(err, sales.ErrSaleNotOpen) {
			t.Fatalf("expected sales.ErrSaleNotOpen, got %v", err)
		}
	})

	t.Run("quantity invalid", func(t *testing.T) {
		_, err := sales.NewService(&fakeReadStore{}, &fakeTxManager{}).AddItem(context.Background(), testScope, saleWithItemsFixture(database.SaleStatusOPEN).ID.String(), sales.AddSaleItemInput{
			ProductID: productFixture(true).ID.String(),
			Quantity:  "abc",
		})
		requireValidationField(t, err, "quantity")
	})

	t.Run("discount invalid", func(t *testing.T) {
		_, err := sales.NewService(&fakeReadStore{}, &fakeTxManager{}).AddItem(context.Background(), testScope, saleWithItemsFixture(database.SaleStatusOPEN).ID.String(), sales.AddSaleItemInput{
			ProductID: productFixture(true).ID.String(),
			Quantity:  "1.000",
			Discount:  strPtr("abc"),
		})
		requireValidationField(t, err, "discount")
	})

	t.Run("discount greater than subtotal", func(t *testing.T) {
		tx := &fakeTxQueries{
			lockSaleByIDFn: func(_ context.Context, scope tenancy.StoreScope, id pgtype.UUID) (database.Sale, error) {
				return saleWithItemsFixture(database.SaleStatusOPEN).toSale(), nil
			},
			getProductByIDFn: func(_ context.Context, scope tenancy.StoreScope, id pgtype.UUID) (database.GetProductByIDForStoreRow, error) {
				return productFixture(true), nil
			},
		}
		_, err := sales.NewService(&fakeReadStore{}, &fakeTxManager{tx: tx}).AddItem(context.Background(), testScope, saleWithItemsFixture(database.SaleStatusOPEN).ID.String(), sales.AddSaleItemInput{
			ProductID: productFixture(true).ID.String(),
			Quantity:  "1.000",
			Discount:  strPtr("100.00"),
		})
		requireValidationField(t, err, "discount")
	})

	t.Run("create item failure rolls back", func(t *testing.T) {
		tx := &fakeTxQueries{
			lockSaleByIDFn: func(_ context.Context, scope tenancy.StoreScope, id pgtype.UUID) (database.Sale, error) {
				return saleWithItemsFixture(database.SaleStatusOPEN).toSale(), nil
			},
			getProductByIDFn: func(_ context.Context, scope tenancy.StoreScope, id pgtype.UUID) (database.GetProductByIDForStoreRow, error) {
				return productFixture(true), nil
			},
			createSaleItemFn: func(_ context.Context, scope tenancy.StoreScope, arg database.CreateSaleItemForStoreParams) (database.SaleItem, error) {
				return database.SaleItem{}, errors.New("insert failed")
			},
		}
		txManager := &fakeTxManager{tx: tx}
		_, err := sales.NewService(&fakeReadStore{}, txManager).AddItem(context.Background(), testScope, saleWithItemsFixture(database.SaleStatusOPEN).ID.String(), sales.AddSaleItemInput{
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
			lockSaleByIDFn: func(_ context.Context, scope tenancy.StoreScope, id pgtype.UUID) (database.Sale, error) {
				return saleWithItemsFixture(database.SaleStatusOPEN).toSale(), nil
			},
			getProductByIDFn: func(_ context.Context, scope tenancy.StoreScope, id pgtype.UUID) (database.GetProductByIDForStoreRow, error) {
				return productFixture(true), nil
			},
			createSaleItemFn: func(_ context.Context, scope tenancy.StoreScope, arg database.CreateSaleItemForStoreParams) (database.SaleItem, error) {
				return item, nil
			},
			listSaleItemsBySaleIDFn: func(_ context.Context, scope tenancy.StoreScope, saleID pgtype.UUID) ([]database.SaleItem, error) {
				return []database.SaleItem{item}, nil
			},
			recalculateSaleTotalsFn: func(_ context.Context, scope tenancy.StoreScope, arg database.RecalculateSaleTotalsForStoreParams) (database.Sale, error) {
				return database.Sale{}, errors.New("recalc failed")
			},
		}
		txManager := &fakeTxManager{tx: tx}
		_, err := sales.NewService(&fakeReadStore{}, txManager).AddItem(context.Background(), testScope, saleWithItemsFixture(database.SaleStatusOPEN).ID.String(), sales.AddSaleItemInput{
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
		var capturedUpdate database.UpdateSaleItemForStoreParams
		item := saleItemFixture(mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8aa"), "1.000", "0.00", "12.90")
		updatedItem := saleItemFixture(mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8aa"), "2.000", "1.00", "24.80")

		tx := &fakeTxQueries{
			lockSaleByIDFn: func(_ context.Context, scope tenancy.StoreScope, id pgtype.UUID) (database.Sale, error) {
				return saleWithItemsFixture(database.SaleStatusOPEN).toSale(), nil
			},
			getSaleItemByIDFn: func(_ context.Context, scope tenancy.StoreScope, arg database.GetSaleItemByIDForStoreParams) (database.SaleItem, error) {
				return item, nil
			},
			updateSaleItemFn: func(_ context.Context, scope tenancy.StoreScope, arg database.UpdateSaleItemForStoreParams) (database.SaleItem, error) {
				capturedUpdate = arg
				return updatedItem, nil
			},
			listSaleItemsBySaleIDFn: func(_ context.Context, scope tenancy.StoreScope, saleID pgtype.UUID) ([]database.SaleItem, error) {
				return []database.SaleItem{updatedItem}, nil
			},
			recalculateSaleTotalsFn: func(_ context.Context, scope tenancy.StoreScope, arg database.RecalculateSaleTotalsForStoreParams) (database.Sale, error) {
				return saleWithItemsFixture(database.SaleStatusOPEN).toSale(), nil
			},
		}

		txManager := &fakeTxManager{tx: tx}
		resp, err := sales.NewService(&fakeReadStore{}, txManager).UpdateItem(context.Background(), testScope, saleWithItemsFixture(database.SaleStatusOPEN).ID.String(), item.ID.String(), sales.UpdateSaleItemInput{
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
			lockSaleByIDFn: func(_ context.Context, scope tenancy.StoreScope, id pgtype.UUID) (database.Sale, error) {
				return saleWithItemsFixture(database.SaleStatusOPEN).toSale(), nil
			},
			getSaleItemByIDFn: func(_ context.Context, scope tenancy.StoreScope, arg database.GetSaleItemByIDForStoreParams) (database.SaleItem, error) {
				return database.SaleItem{}, pgx.ErrNoRows
			},
		}
		_, err := sales.NewService(&fakeReadStore{}, &fakeTxManager{tx: tx}).UpdateItem(context.Background(), testScope, saleWithItemsFixture(database.SaleStatusOPEN).ID.String(), mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8aa").String(), sales.UpdateSaleItemInput{
			Quantity: "2.000",
		})
		if !errors.Is(err, sales.ErrSaleItemNotFound) {
			t.Fatalf("expected sales.ErrSaleItemNotFound, got %v", err)
		}
	})

	t.Run("item on another sale not found", func(t *testing.T) {
		tx := &fakeTxQueries{
			lockSaleByIDFn: func(_ context.Context, scope tenancy.StoreScope, id pgtype.UUID) (database.Sale, error) {
				return saleWithItemsFixture(database.SaleStatusOPEN).toSale(), nil
			},
			getSaleItemByIDFn: func(_ context.Context, scope tenancy.StoreScope, arg database.GetSaleItemByIDForStoreParams) (database.SaleItem, error) {
				return database.SaleItem{}, pgx.ErrNoRows
			},
		}
		_, err := sales.NewService(&fakeReadStore{}, &fakeTxManager{tx: tx}).UpdateItem(context.Background(), testScope, saleWithItemsFixture(database.SaleStatusOPEN).ID.String(), mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8ab").String(), sales.UpdateSaleItemInput{
			Quantity: "2.000",
		})
		if !errors.Is(err, sales.ErrSaleItemNotFound) {
			t.Fatalf("expected sales.ErrSaleItemNotFound, got %v", err)
		}
	})

	t.Run("sale not open", func(t *testing.T) {
		tx := &fakeTxQueries{
			lockSaleByIDFn: func(_ context.Context, scope tenancy.StoreScope, id pgtype.UUID) (database.Sale, error) {
				return saleWithItemsFixture(database.SaleStatusCOMPLETED).toSale(), nil
			},
		}
		_, err := sales.NewService(&fakeReadStore{}, &fakeTxManager{tx: tx}).UpdateItem(context.Background(), testScope, saleWithItemsFixture(database.SaleStatusOPEN).ID.String(), mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8aa").String(), sales.UpdateSaleItemInput{
			Quantity: "2.000",
		})
		if !errors.Is(err, sales.ErrSaleNotOpen) {
			t.Fatalf("expected sales.ErrSaleNotOpen, got %v", err)
		}
	})

	t.Run("discount greater than subtotal", func(t *testing.T) {
		item := saleItemFixture(mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8aa"), "1.000", "0.00", "12.90")
		tx := &fakeTxQueries{
			lockSaleByIDFn: func(_ context.Context, scope tenancy.StoreScope, id pgtype.UUID) (database.Sale, error) {
				return saleWithItemsFixture(database.SaleStatusOPEN).toSale(), nil
			},
			getSaleItemByIDFn: func(_ context.Context, scope tenancy.StoreScope, arg database.GetSaleItemByIDForStoreParams) (database.SaleItem, error) {
				return item, nil
			},
		}
		_, err := sales.NewService(&fakeReadStore{}, &fakeTxManager{tx: tx}).UpdateItem(context.Background(), testScope, saleWithItemsFixture(database.SaleStatusOPEN).ID.String(), item.ID.String(), sales.UpdateSaleItemInput{
			Quantity: "1.000",
			Discount: strPtr("100.00"),
		})
		requireValidationField(t, err, "discount")
	})

	t.Run("rollback on recalc error", func(t *testing.T) {
		item := saleItemFixture(mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8aa"), "1.000", "0.00", "12.90")
		updatedItem := saleItemFixture(mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8aa"), "2.000", "0.00", "25.80")
		tx := &fakeTxQueries{
			lockSaleByIDFn: func(_ context.Context, scope tenancy.StoreScope, id pgtype.UUID) (database.Sale, error) {
				return saleWithItemsFixture(database.SaleStatusOPEN).toSale(), nil
			},
			getSaleItemByIDFn: func(_ context.Context, scope tenancy.StoreScope, arg database.GetSaleItemByIDForStoreParams) (database.SaleItem, error) {
				return item, nil
			},
			updateSaleItemFn: func(_ context.Context, scope tenancy.StoreScope, arg database.UpdateSaleItemForStoreParams) (database.SaleItem, error) {
				return updatedItem, nil
			},
			listSaleItemsBySaleIDFn: func(_ context.Context, scope tenancy.StoreScope, saleID pgtype.UUID) ([]database.SaleItem, error) {
				return []database.SaleItem{updatedItem}, nil
			},
			recalculateSaleTotalsFn: func(_ context.Context, scope tenancy.StoreScope, arg database.RecalculateSaleTotalsForStoreParams) (database.Sale, error) {
				return database.Sale{}, errors.New("recalc failed")
			},
		}
		txManager := &fakeTxManager{tx: tx}
		_, err := sales.NewService(&fakeReadStore{}, txManager).UpdateItem(context.Background(), testScope, saleWithItemsFixture(database.SaleStatusOPEN).ID.String(), item.ID.String(), sales.UpdateSaleItemInput{
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
		var capturedDelete database.DeleteSaleItemForStoreParams
		item := saleItemFixture(mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8aa"), "1.000", "0.00", "12.90")

		tx := &fakeTxQueries{
			lockSaleByIDFn: func(_ context.Context, scope tenancy.StoreScope, id pgtype.UUID) (database.Sale, error) {
				return saleWithItemsFixture(database.SaleStatusOPEN).toSale(), nil
			},
			getSaleItemByIDFn: func(_ context.Context, scope tenancy.StoreScope, arg database.GetSaleItemByIDForStoreParams) (database.SaleItem, error) {
				return item, nil
			},
			deleteSaleItemFn: func(_ context.Context, scope tenancy.StoreScope, arg database.DeleteSaleItemForStoreParams) (database.SaleItem, error) {
				capturedDelete = arg
				return item, nil
			},
			listSaleItemsBySaleIDFn: func(_ context.Context, scope tenancy.StoreScope, saleID pgtype.UUID) ([]database.SaleItem, error) {
				return nil, nil
			},
			recalculateSaleTotalsFn: func(_ context.Context, scope tenancy.StoreScope, arg database.RecalculateSaleTotalsForStoreParams) (database.Sale, error) {
				return saleWithItemsFixture(database.SaleStatusOPEN).toSale(), nil
			},
		}

		txManager := &fakeTxManager{tx: tx}
		resp, err := sales.NewService(&fakeReadStore{}, txManager).RemoveItem(context.Background(), testScope, saleWithItemsFixture(database.SaleStatusOPEN).ID.String(), item.ID.String())
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
			lockSaleByIDFn: func(_ context.Context, scope tenancy.StoreScope, id pgtype.UUID) (database.Sale, error) {
				return saleWithItemsFixture(database.SaleStatusOPEN).toSale(), nil
			},
			getSaleItemByIDFn: func(_ context.Context, scope tenancy.StoreScope, arg database.GetSaleItemByIDForStoreParams) (database.SaleItem, error) {
				return database.SaleItem{}, pgx.ErrNoRows
			},
		}
		_, err := sales.NewService(&fakeReadStore{}, &fakeTxManager{tx: tx}).RemoveItem(context.Background(), testScope, saleWithItemsFixture(database.SaleStatusOPEN).ID.String(), mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8aa").String())
		if !errors.Is(err, sales.ErrSaleItemNotFound) {
			t.Fatalf("expected sales.ErrSaleItemNotFound, got %v", err)
		}
	})

	t.Run("sale not open", func(t *testing.T) {
		tx := &fakeTxQueries{
			lockSaleByIDFn: func(_ context.Context, scope tenancy.StoreScope, id pgtype.UUID) (database.Sale, error) {
				return saleWithItemsFixture(database.SaleStatusCANCELLED).toSale(), nil
			},
		}
		_, err := sales.NewService(&fakeReadStore{}, &fakeTxManager{tx: tx}).RemoveItem(context.Background(), testScope, saleWithItemsFixture(database.SaleStatusOPEN).ID.String(), mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8aa").String())
		if !errors.Is(err, sales.ErrSaleNotOpen) {
			t.Fatalf("expected sales.ErrSaleNotOpen, got %v", err)
		}
	})

	t.Run("rollback on delete error", func(t *testing.T) {
		item := saleItemFixture(mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8aa"), "1.000", "0.00", "12.90")
		tx := &fakeTxQueries{
			lockSaleByIDFn: func(_ context.Context, scope tenancy.StoreScope, id pgtype.UUID) (database.Sale, error) {
				return saleWithItemsFixture(database.SaleStatusOPEN).toSale(), nil
			},
			getSaleItemByIDFn: func(_ context.Context, scope tenancy.StoreScope, arg database.GetSaleItemByIDForStoreParams) (database.SaleItem, error) {
				return item, nil
			},
			deleteSaleItemFn: func(_ context.Context, scope tenancy.StoreScope, arg database.DeleteSaleItemForStoreParams) (database.SaleItem, error) {
				return database.SaleItem{}, errors.New("delete failed")
			},
		}
		txManager := &fakeTxManager{tx: tx}
		_, err := sales.NewService(&fakeReadStore{}, txManager).RemoveItem(context.Background(), testScope, saleWithItemsFixture(database.SaleStatusOPEN).ID.String(), item.ID.String())
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
			lockSaleByIDFn: func(_ context.Context, scope tenancy.StoreScope, id pgtype.UUID) (database.Sale, error) {
				return saleWithItemsFixture(database.SaleStatusOPEN).toSale(), nil
			},
			cancelSaleFn: func(_ context.Context, scope tenancy.ActorScope, params database.CancelSaleForStoreParams) (database.Sale, error) {
				return saleWithItemsFixture(database.SaleStatusCANCELLED).toSale(), nil
			},
			listSaleItemsBySaleIDFn: func(_ context.Context, scope tenancy.StoreScope, saleID pgtype.UUID) ([]database.SaleItem, error) {
				return nil, nil
			},
		}
		txManager := &fakeTxManager{tx: tx}
		resp, err := sales.NewService(&fakeReadStore{}, txManager).Cancel(context.Background(), testScope, saleWithItemsFixture(database.SaleStatusOPEN).ID.String())
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
			lockSaleByIDFn: func(_ context.Context, scope tenancy.StoreScope, id pgtype.UUID) (database.Sale, error) {
				return database.Sale{}, pgx.ErrNoRows
			},
		}
		_, err := sales.NewService(&fakeReadStore{}, &fakeTxManager{tx: tx}).Cancel(context.Background(), testScope, saleWithItemsFixture(database.SaleStatusOPEN).ID.String())
		if !errors.Is(err, sales.ErrSaleNotFound) {
			t.Fatalf("expected sales.ErrSaleNotFound, got %v", err)
		}
	})

	t.Run("already cancelled", func(t *testing.T) {
		tx := &fakeTxQueries{
			lockSaleByIDFn: func(_ context.Context, scope tenancy.StoreScope, id pgtype.UUID) (database.Sale, error) {
				return saleWithItemsFixture(database.SaleStatusCANCELLED).toSale(), nil
			},
		}
		_, err := sales.NewService(&fakeReadStore{}, &fakeTxManager{tx: tx}).Cancel(context.Background(), testScope, saleWithItemsFixture(database.SaleStatusOPEN).ID.String())
		if !errors.Is(err, sales.ErrSaleAlreadyCancelled) {
			t.Fatalf("expected sales.ErrSaleAlreadyCancelled, got %v", err)
		}
	})

	t.Run("completed sale not open", func(t *testing.T) {
		tx := &fakeTxQueries{
			lockSaleByIDFn: func(_ context.Context, scope tenancy.StoreScope, id pgtype.UUID) (database.Sale, error) {
				return saleWithItemsFixture(database.SaleStatusCOMPLETED).toSale(), nil
			},
		}
		_, err := sales.NewService(&fakeReadStore{}, &fakeTxManager{tx: tx}).Cancel(context.Background(), testScope, saleWithItemsFixture(database.SaleStatusOPEN).ID.String())
		if !errors.Is(err, sales.ErrSaleNotOpen) {
			t.Fatalf("expected sales.ErrSaleNotOpen, got %v", err)
		}
	})

	t.Run("rollback on persistence error", func(t *testing.T) {
		tx := &fakeTxQueries{
			lockSaleByIDFn: func(_ context.Context, scope tenancy.StoreScope, id pgtype.UUID) (database.Sale, error) {
				return saleWithItemsFixture(database.SaleStatusOPEN).toSale(), nil
			},
			cancelSaleFn: func(_ context.Context, scope tenancy.ActorScope, params database.CancelSaleForStoreParams) (database.Sale, error) {
				return database.Sale{}, errors.New("cancel failed")
			},
		}
		txManager := &fakeTxManager{tx: tx}
		_, err := sales.NewService(&fakeReadStore{}, txManager).Cancel(context.Background(), testScope, saleWithItemsFixture(database.SaleStatusOPEN).ID.String())
		if err == nil {
			t.Fatalf("expected error")
		}
		if !txManager.rolledBack || txManager.committed {
			t.Fatalf("expected rollback")
		}
	})
}

func TestQuantityAndMoneyValidation(t *testing.T) {
	svc := sales.NewService(&fakeReadStore{}, &fakeTxManager{})

	_, err := svc.AddItem(context.Background(), testScope, saleWithItemsFixture(database.SaleStatusOPEN).ID.String(), sales.AddSaleItemInput{
		ProductID: productFixture(true).ID.String(),
		Quantity:  "0",
	})
	requireValidationField(t, err, "quantity")

	_, err = svc.AddItem(context.Background(), testScope, saleWithItemsFixture(database.SaleStatusOPEN).ID.String(), sales.AddSaleItemInput{
		ProductID: productFixture(true).ID.String(),
		Quantity:  "1.000",
		Discount:  strPtr("-1.00"),
	})
	requireValidationField(t, err, "discount")
}

type fakeReadStore struct {
	getSaleByIDFn           func(context.Context, tenancy.StoreScope, pgtype.UUID) (database.Sale, error)
	listSalesFn             func(context.Context, tenancy.StoreScope, database.ListSalesForStoreParams) ([]database.Sale, error)
	countSalesFn            func(context.Context, tenancy.StoreScope, database.CountSalesForStoreParams) (int64, error)
	listSaleItemsBySaleIDFn func(context.Context, tenancy.StoreScope, pgtype.UUID) ([]database.SaleItem, error)
}

func (f *fakeReadStore) GetSaleByID(ctx context.Context, scope tenancy.StoreScope, id pgtype.UUID) (database.Sale, error) {
	if f.getSaleByIDFn == nil {
		panic("unexpected GetSaleByID call")
	}
	return f.getSaleByIDFn(ctx, scope, id)
}

func (f *fakeReadStore) ListSales(ctx context.Context, scope tenancy.StoreScope, arg database.ListSalesForStoreParams) ([]database.Sale, error) {
	if f.listSalesFn == nil {
		panic("unexpected ListSales call")
	}
	return f.listSalesFn(ctx, scope, arg)
}

func (f *fakeReadStore) CountSales(ctx context.Context, scope tenancy.StoreScope, arg database.CountSalesForStoreParams) (int64, error) {
	if f.countSalesFn == nil {
		panic("unexpected CountSales call")
	}
	return f.countSalesFn(ctx, scope, arg)
}

func (f *fakeReadStore) ListSaleItemsBySaleID(ctx context.Context, scope tenancy.StoreScope, saleID pgtype.UUID) ([]database.SaleItem, error) {
	if f.listSaleItemsBySaleIDFn == nil {
		panic("unexpected ListSaleItemsBySaleID call")
	}
	return f.listSaleItemsBySaleIDFn(ctx, scope, saleID)
}

type fakeTxQueries struct {
	getSaleByIDFn           func(context.Context, tenancy.StoreScope, pgtype.UUID) (database.Sale, error)
	lockSaleByIDFn          func(context.Context, tenancy.StoreScope, pgtype.UUID) (database.Sale, error)
	getProductByIDFn        func(context.Context, tenancy.StoreScope, pgtype.UUID) (database.GetProductByIDForStoreRow, error)
	getSaleItemByIDFn       func(context.Context, tenancy.StoreScope, database.GetSaleItemByIDForStoreParams) (database.SaleItem, error)
	createSaleFn            func(context.Context, tenancy.ActorScope, database.CreateSaleForStoreParams) (database.CreateSaleForStoreRow, error)
	createSaleItemFn        func(context.Context, tenancy.StoreScope, database.CreateSaleItemForStoreParams) (database.SaleItem, error)
	updateSaleItemFn        func(context.Context, tenancy.StoreScope, database.UpdateSaleItemForStoreParams) (database.SaleItem, error)
	deleteSaleItemFn        func(context.Context, tenancy.StoreScope, database.DeleteSaleItemForStoreParams) (database.SaleItem, error)
	recalculateSaleTotalsFn func(context.Context, tenancy.StoreScope, database.RecalculateSaleTotalsForStoreParams) (database.Sale, error)
	cancelSaleFn            func(context.Context, tenancy.ActorScope, database.CancelSaleForStoreParams) (database.Sale, error)
	completeSaleFn          func(context.Context, tenancy.ActorScope, database.CompleteSaleForStoreParams) (database.Sale, error)
	listSaleItemsBySaleIDFn func(context.Context, tenancy.StoreScope, pgtype.UUID) ([]database.SaleItem, error)
}

func (f *fakeTxQueries) GetSaleByID(ctx context.Context, scope tenancy.StoreScope, id pgtype.UUID) (database.Sale, error) {
	if f.getSaleByIDFn == nil {
		panic("unexpected GetSaleByID call")
	}
	return f.getSaleByIDFn(ctx, scope, id)
}

func (f *fakeTxQueries) LockSaleByID(ctx context.Context, scope tenancy.StoreScope, id pgtype.UUID) (database.Sale, error) {
	if f.lockSaleByIDFn == nil {
		panic("unexpected LockSaleByID call")
	}
	return f.lockSaleByIDFn(ctx, scope, id)
}

func (f *fakeTxQueries) GetProductByID(ctx context.Context, scope tenancy.StoreScope, id pgtype.UUID) (database.GetProductByIDForStoreRow, error) {
	if f.getProductByIDFn == nil {
		panic("unexpected GetProductByID call")
	}
	return f.getProductByIDFn(ctx, scope, id)
}

func (f *fakeTxQueries) GetSaleItemByID(ctx context.Context, scope tenancy.StoreScope, arg database.GetSaleItemByIDForStoreParams) (database.SaleItem, error) {
	if f.getSaleItemByIDFn == nil {
		panic("unexpected GetSaleItemByID call")
	}
	return f.getSaleItemByIDFn(ctx, scope, arg)
}

func (f *fakeTxQueries) CreateSaleForStore(ctx context.Context, scope tenancy.ActorScope, arg database.CreateSaleForStoreParams) (database.CreateSaleForStoreRow, error) {
	if f.createSaleFn == nil {
		panic("unexpected CreateSaleForStore call")
	}
	return f.createSaleFn(ctx, scope, arg)
}

func (f *fakeTxQueries) CreateSaleItemForStore(ctx context.Context, scope tenancy.StoreScope, arg database.CreateSaleItemForStoreParams) (database.SaleItem, error) {
	if f.createSaleItemFn == nil {
		panic("unexpected CreateSaleItemForStore call")
	}
	return f.createSaleItemFn(ctx, scope, arg)
}

func (f *fakeTxQueries) UpdateSaleItemForStore(ctx context.Context, scope tenancy.StoreScope, arg database.UpdateSaleItemForStoreParams) (database.SaleItem, error) {
	if f.updateSaleItemFn == nil {
		panic("unexpected UpdateSaleItemForStore call")
	}
	return f.updateSaleItemFn(ctx, scope, arg)
}

func (f *fakeTxQueries) DeleteSaleItemForStore(ctx context.Context, scope tenancy.StoreScope, arg database.DeleteSaleItemForStoreParams) (database.SaleItem, error) {
	if f.deleteSaleItemFn == nil {
		panic("unexpected DeleteSaleItemForStore call")
	}
	return f.deleteSaleItemFn(ctx, scope, arg)
}

func (f *fakeTxQueries) RecalculateSaleTotalsForStore(ctx context.Context, scope tenancy.StoreScope, arg database.RecalculateSaleTotalsForStoreParams) (database.Sale, error) {
	if f.recalculateSaleTotalsFn == nil {
		panic("unexpected RecalculateSaleTotalsForStore call")
	}
	return f.recalculateSaleTotalsFn(ctx, scope, arg)
}

func (f *fakeTxQueries) CancelSaleForStore(ctx context.Context, scope tenancy.ActorScope, arg database.CancelSaleForStoreParams) (database.Sale, error) {
	if f.cancelSaleFn == nil {
		panic("unexpected CancelSaleForStore call")
	}
	return f.cancelSaleFn(ctx, scope, arg)
}

func (f *fakeTxQueries) CompleteSaleForStore(ctx context.Context, scope tenancy.ActorScope, arg database.CompleteSaleForStoreParams) (database.Sale, error) {
	if f.completeSaleFn == nil {
		panic("unexpected CompleteSaleForStore call")
	}
	return f.completeSaleFn(ctx, scope, arg)
}

func (f *fakeTxQueries) ListSaleItemsBySaleID(ctx context.Context, scope tenancy.StoreScope, saleID pgtype.UUID) ([]database.SaleItem, error) {
	if f.listSaleItemsBySaleIDFn == nil {
		panic("unexpected ListSaleItemsBySaleID call")
	}
	return f.listSaleItemsBySaleIDFn(ctx, scope, saleID)
}

type fakeTxManager struct {
	tx         sales.TxQueries
	committed  bool
	rolledBack bool
}

func (f *fakeTxManager) WithTx(ctx context.Context, scope tenancy.ActorScope, fn func(sales.TxQueries) error) error {
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
	ID                      pgtype.UUID
	OrganizationID          pgtype.UUID
	StoreID                 pgtype.UUID
	Number                  int64
	Status                  database.SaleStatus
	Subtotal                pgtype.Numeric
	Discount                pgtype.Numeric
	Addition                pgtype.Numeric
	Total                   pgtype.Numeric
	OpenedByMembershipID    pgtype.UUID
	CompletedByMembershipID pgtype.UUID
	CancelledByMembershipID pgtype.UUID
	OpenedAt                pgtype.Timestamptz
	CompletedAt             pgtype.Timestamptz
	CancelledAt             pgtype.Timestamptz
	CreatedAt               pgtype.Timestamptz
	UpdatedAt               pgtype.Timestamptz
	IdempotencyKey          string
}

func saleRowFixture(status database.SaleStatus) saleFixtureValues {
	f := saleFixtureValues{
		ID:                      mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8a1"),
		OrganizationID:          testOrgID,
		StoreID:                 testStoreID,
		Number:                  42,
		Status:                  status,
		Subtotal:                mustNumeric("0.00"),
		Discount:                mustNumeric("0.00"),
		Addition:                mustNumeric("0.00"),
		Total:                   mustNumeric("0.00"),
		OpenedByMembershipID:    testMembershipID,
		OpenedAt:                mustTime("2026-07-15T10:00:00Z"),
		CreatedAt:               mustTime("2026-07-15T10:00:00Z"),
		UpdatedAt:               mustTime("2026-07-15T10:00:00Z"),
		IdempotencyKey:          "sale-1",
	}

	switch status {
	case database.SaleStatusCOMPLETED:
		f.CompletedAt = mustTime("2026-07-15T11:00:00Z")
		f.CompletedByMembershipID = testMembershipID
	case database.SaleStatusCANCELLED:
		f.CancelledAt = mustTime("2026-07-15T11:00:00Z")
		f.CancelledByMembershipID = testMembershipID
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

func (f saleFixtureValues) toSale() database.Sale {
	return database.Sale{
		ID:                      f.ID,
		OrganizationID:          f.OrganizationID,
		StoreID:                 f.StoreID,
		Number:                  f.Number,
		IdempotencyKey:          f.IdempotencyKey,
		Status:                  f.Status,
		Subtotal:                f.Subtotal,
		Discount:                f.Discount,
		Addition:                f.Addition,
		Total:                   f.Total,
		OpenedByMembershipID:    f.OpenedByMembershipID,
		CompletedByMembershipID: f.CompletedByMembershipID,
		CancelledByMembershipID: f.CancelledByMembershipID,
		OpenedAt:                f.OpenedAt,
		CompletedAt:             f.CompletedAt,
		CancelledAt:             f.CancelledAt,
		CreatedAt:               f.CreatedAt,
		UpdatedAt:               f.UpdatedAt,
	}
}

func (f saleFixtureValues) create() database.CreateSaleForStoreRow {
	return database.CreateSaleForStoreRow{
		ID:                      f.ID,
		OrganizationID:          f.OrganizationID,
		StoreID:                 f.StoreID,
		Number:                  f.Number,
		IdempotencyKey:          f.IdempotencyKey,
		Status:                  f.Status,
		Subtotal:                f.Subtotal,
		Discount:                f.Discount,
		Addition:                f.Addition,
		Total:                   f.Total,
		OpenedByMembershipID:    f.OpenedByMembershipID,
		CompletedByMembershipID: f.CompletedByMembershipID,
		CancelledByMembershipID: f.CancelledByMembershipID,
		OpenedAt:                f.OpenedAt,
		CompletedAt:             f.CompletedAt,
		CancelledAt:             f.CancelledAt,
		CreatedAt:               f.CreatedAt,
		UpdatedAt:               f.UpdatedAt,
	}
}

func saleItemFixture(id pgtype.UUID, quantity, discount, total string) database.SaleItem {
	return database.SaleItem{
		ID:          id,
		OrganizationID: testOrgID,
		StoreID:        testStoreID,
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

func productFixture(active bool) database.GetProductByIDForStoreRow {
	return database.GetProductByIDForStoreRow{
		OrganizationID: testOrgID,
		StoreID:        testStoreID,
		ID:             mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8a9"),
		SKU:            "COCA-2L",
		Barcode:        pgtype.Text{String: "7890000000000", Valid: true},
		Name:           "Coca-Cola 2L",
		Price:          mustNumeric("12.90"),
		Cost:           mustNumeric("8.50"),
		IsActive:       active,
		CreatedAt:      mustTime("2026-07-15T10:00:00Z"),
		UpdatedAt:      mustTime("2026-07-15T10:00:00Z"),
	}
}

func requireValidationField(t *testing.T, err error, field string) {
	t.Helper()

	var validationErr *sales.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected sales.ValidationError, got %v", err)
	}
	if validationErr.Field != field {
		t.Fatalf("expected field %q, got %q", field, validationErr.Field)
	}
}

func strPtr(value string) *string {
	return &value
}

func numericToMoneyString(n pgtype.Numeric) (string, error) {
	if !n.Valid {
		return "", errors.New("numeric is null")
	}
	if n.Int == nil {
		return "0.00", nil
	}
	if n.NaN || n.InfinityModifier != 0 {
		return "", errors.New("invalid numeric")
	}
	val := new(big.Int).Set(n.Int)
	targetExp := int32(-2)
	switch {
	case n.Exp == targetExp:
	case n.Exp > targetExp:
		pow := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(n.Exp-targetExp)), nil)
		val = val.Mul(val, pow)
	default:
		pow := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(targetExp-n.Exp)), nil)
		val = val.Div(val, pow)
	}
	return formatScaled(val, 2), nil
}

func numericToQuantityString(n pgtype.Numeric) (string, error) {
	if !n.Valid {
		return "", errors.New("numeric is null")
	}
	if n.Int == nil {
		return "0.000", nil
	}
	if n.NaN || n.InfinityModifier != 0 {
		return "", errors.New("invalid numeric")
	}
	val := new(big.Int).Set(n.Int)
	targetExp := int32(-3)
	switch {
	case n.Exp == targetExp:
	case n.Exp > targetExp:
		pow := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(n.Exp-targetExp)), nil)
		val = val.Mul(val, pow)
	default:
		pow := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(targetExp-n.Exp)), nil)
		val = val.Div(val, pow)
	}
	return formatScaled(val, 3), nil
}

func formatScaled(val *big.Int, scale int) string {
	sign := ""
	if val.Sign() < 0 {
		sign = "-"
		val = new(big.Int).Abs(val)
	}
	s := val.String()
	if scale == 0 {
		return sign + s
	}
	for len(s) <= scale {
		s = "0" + s
	}
	intPart := s[:len(s)-scale]
	fracPart := s[len(s)-scale:]
	return sign + intPart + "." + fracPart
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
