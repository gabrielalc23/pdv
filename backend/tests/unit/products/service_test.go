package products_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/gabrielalc23/pdv/internal/products"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestCreateProductValid(t *testing.T) {
	store := &fakeStore{
		getProductBySKUFn: func(context.Context, string) (database.Product, error) {
			return database.Product{}, pgx.ErrNoRows
		},
		getProductByBarcodeFn: func(context.Context, pgtype.Text) (database.Product, error) {
			return database.Product{}, pgx.ErrNoRows
		},
		createProductFn: func(ctx context.Context, arg database.CreateProductParams) (database.Product, error) {
			if arg.SKU != "COCA-2L" {
				t.Fatalf("unexpected SKU: %s", arg.SKU)
			}
			if !arg.Barcode.Valid || arg.Barcode.String != "7890000000000" {
				t.Fatalf("unexpected barcode: %+v", arg.Barcode)
			}
			if !arg.Price.Valid || !arg.Cost.Valid {
				t.Fatalf("expected price and cost to be valid")
			}
			return productFixture(true), nil
		},
	}

	svc := products.NewService(store)
	resp, err := svc.Create(context.Background(), products.UpsertProductInput{
		SKU:     " COCA-2L ",
		Barcode: strPtr(" 7890000000000 "),
		Name:    " Coca-Cola 2L ",
		Price:   "12.90",
		Cost:    strPtr("8.50"),
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	assertProductResponse(t, resp)
	if resp.SKU != "COCA-2L" {
		t.Fatalf("unexpected SKU: %s", resp.SKU)
	}
	if resp.Barcode == nil || *resp.Barcode != "7890000000000" {
		t.Fatalf("unexpected barcode: %#v", resp.Barcode)
	}
	if resp.Price != "12.90" {
		t.Fatalf("unexpected price: %s", resp.Price)
	}
	if resp.Cost == nil || *resp.Cost != "8.50" {
		t.Fatalf("unexpected cost: %#v", resp.Cost)
	}
}

func TestCreateProductSKUEmpty(t *testing.T) {
	svc := products.NewService(&fakeStore{})

	_, err := svc.Create(context.Background(), products.UpsertProductInput{
		SKU:   "   ",
		Name:  "Coca-Cola 2L",
		Price: "12.90",
	})
	requireValidationField(t, err, "sku")
}

func TestCreateProductNameEmpty(t *testing.T) {
	svc := products.NewService(&fakeStore{})

	_, err := svc.Create(context.Background(), products.UpsertProductInput{
		SKU:   "COCA-2L",
		Name:  "   ",
		Price: "12.90",
	})
	requireValidationField(t, err, "name")
}

func TestCreateProductPriceInvalid(t *testing.T) {
	svc := products.NewService(&fakeStore{})

	_, err := svc.Create(context.Background(), products.UpsertProductInput{
		SKU:   "COCA-2L",
		Name:  "Coca-Cola 2L",
		Price: "abc",
	})
	requireValidationField(t, err, "price")
}

func TestCreateProductPriceNegative(t *testing.T) {
	svc := products.NewService(&fakeStore{})

	_, err := svc.Create(context.Background(), products.UpsertProductInput{
		SKU:   "COCA-2L",
		Name:  "Coca-Cola 2L",
		Price: "-1.00",
	})
	requireValidationField(t, err, "price")
}

func TestCreateProductPriceTooManyDecimals(t *testing.T) {
	svc := products.NewService(&fakeStore{})

	_, err := svc.Create(context.Background(), products.UpsertProductInput{
		SKU:   "COCA-2L",
		Name:  "Coca-Cola 2L",
		Price: "1.234",
	})
	requireValidationField(t, err, "price")
}

func TestCreateProductBarcodeEmptyWhenInformed(t *testing.T) {
	svc := products.NewService(&fakeStore{})

	_, err := svc.Create(context.Background(), products.UpsertProductInput{
		SKU:     "COCA-2L",
		Barcode: strPtr("   "),
		Name:    "Coca-Cola 2L",
		Price:   "12.90",
	})
	requireValidationField(t, err, "barcode")
}

func TestGetProductNotFound(t *testing.T) {
	svc := products.NewService(&fakeStore{
		getProductByIDFn: func(context.Context, pgtype.UUID) (database.Product, error) {
			return database.Product{}, pgx.ErrNoRows
		},
	})

	_, err := svc.Get(context.Background(), "01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8a9")
	if !errors.Is(err, products.ErrProductNotFound) {
		t.Fatalf("expected products.ErrProductNotFound, got %v", err)
	}
}

func TestCreateProductDuplicateSKU(t *testing.T) {
	store := &fakeStore{
		getProductBySKUFn: func(context.Context, string) (database.Product, error) {
			return productFixture(true), nil
		},
	}

	svc := products.NewService(store)
	_, err := svc.Create(context.Background(), products.UpsertProductInput{
		SKU:   "COCA-2L",
		Name:  "Coca-Cola 2L",
		Price: "12.90",
	})
	if !errors.Is(err, products.ErrSKUAlreadyExists) {
		t.Fatalf("expected products.ErrSKUAlreadyExists, got %v", err)
	}
}

func TestCreateProductDuplicateBarcode(t *testing.T) {
	store := &fakeStore{
		getProductBySKUFn: func(context.Context, string) (database.Product, error) {
			return database.Product{}, pgx.ErrNoRows
		},
		getProductByBarcodeFn: func(context.Context, pgtype.Text) (database.Product, error) {
			return productFixture(true), nil
		},
	}

	svc := products.NewService(store)
	_, err := svc.Create(context.Background(), products.UpsertProductInput{
		SKU:     "COCA-2L",
		Barcode: strPtr("7890000000000"),
		Name:    "Coca-Cola 2L",
		Price:   "12.90",
	})
	if !errors.Is(err, products.ErrBarcodeAlreadyExists) {
		t.Fatalf("expected products.ErrBarcodeAlreadyExists, got %v", err)
	}
}

func TestListProductsDefaultPagination(t *testing.T) {
	var capturedCount database.CountProductsParams
	var capturedList database.ListProductsParams

	svc := products.NewService(&fakeStore{
		countProductsFn: func(_ context.Context, arg database.CountProductsParams) (int64, error) {
			capturedCount = arg
			return 1, nil
		},
		listProductsFn: func(_ context.Context, arg database.ListProductsParams) ([]database.Product, error) {
			capturedList = arg
			return []database.Product{productFixture(true)}, nil
		},
	})

	resp, err := svc.List(context.Background(), products.ListProductsInput{})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}

	if resp.Pagination.Page != 1 || resp.Pagination.PageSize != 20 {
		t.Fatalf("unexpected pagination: %+v", resp.Pagination)
	}
	if resp.Pagination.Total != 1 || resp.Pagination.TotalPages != 1 {
		t.Fatalf("unexpected totals: %+v", resp.Pagination)
	}
	if capturedList.PageSize != 20 || capturedList.PageOffset != 0 {
		t.Fatalf("unexpected list params: %+v", capturedList)
	}
	activeOnly, ok := capturedCount.ActiveOnly.(bool)
	if !ok || activeOnly {
		t.Fatalf("expected activeOnly false, got %#v", capturedCount.ActiveOnly)
	}
	if capturedCount.Search.Valid {
		t.Fatalf("expected empty search param, got %#v", capturedCount.Search)
	}
}

func TestListProductsPageSizeMaximum(t *testing.T) {
	svc := products.NewService(&fakeStore{})

	pageSize := 101
	_, err := svc.List(context.Background(), products.ListProductsInput{
		PageSize: &pageSize,
	})
	requireValidationField(t, err, "pageSize")
}

func TestActivateProduct(t *testing.T) {
	t.Run("activates inactive product", func(t *testing.T) {
		activated := false
		svc := products.NewService(&fakeStore{
			getProductByIDFn: func(context.Context, pgtype.UUID) (database.Product, error) {
				product := productFixture(false)
				return product, nil
			},
			activateProductFn: func(context.Context, pgtype.UUID) (database.Product, error) {
				activated = true
				return productFixture(true), nil
			},
		})

		resp, err := svc.Activate(context.Background(), productFixture(true).ID.String())
		if err != nil {
			t.Fatalf("Activate returned error: %v", err)
		}
		if !activated {
			t.Fatalf("expected ActivateProduct to be called")
		}
		if !resp.IsActive {
			t.Fatalf("expected product to be active")
		}
	})

	t.Run("returns active product unchanged", func(t *testing.T) {
		activated := false
		activeProduct := productFixture(true)
		svc := products.NewService(&fakeStore{
			getProductByIDFn: func(context.Context, pgtype.UUID) (database.Product, error) {
				return activeProduct, nil
			},
			activateProductFn: func(context.Context, pgtype.UUID) (database.Product, error) {
				activated = true
				return activeProduct, nil
			},
		})

		resp, err := svc.Activate(context.Background(), activeProduct.ID.String())
		if err != nil {
			t.Fatalf("Activate returned error: %v", err)
		}
		if activated {
			t.Fatalf("expected ActivateProduct not to be called")
		}
		if !resp.IsActive {
			t.Fatalf("expected product to remain active")
		}
	})
}

func TestDeactivateProduct(t *testing.T) {
	t.Run("deactivates active product", func(t *testing.T) {
		deactivated := false
		activeProduct := productFixture(true)
		svc := products.NewService(&fakeStore{
			getProductByIDFn: func(context.Context, pgtype.UUID) (database.Product, error) {
				return activeProduct, nil
			},
			deactivateProductFn: func(context.Context, pgtype.UUID) (database.Product, error) {
				deactivated = true
				return productFixture(false), nil
			},
		})

		resp, err := svc.Deactivate(context.Background(), activeProduct.ID.String())
		if err != nil {
			t.Fatalf("Deactivate returned error: %v", err)
		}
		if !deactivated {
			t.Fatalf("expected DeactivateProduct to be called")
		}
		if resp.IsActive {
			t.Fatalf("expected product to be inactive")
		}
	})

	t.Run("returns inactive product unchanged", func(t *testing.T) {
		deactivated := false
		inactiveProduct := productFixture(false)
		svc := products.NewService(&fakeStore{
			getProductByIDFn: func(context.Context, pgtype.UUID) (database.Product, error) {
				return inactiveProduct, nil
			},
			deactivateProductFn: func(context.Context, pgtype.UUID) (database.Product, error) {
				deactivated = true
				return inactiveProduct, nil
			},
		})

		resp, err := svc.Deactivate(context.Background(), inactiveProduct.ID.String())
		if err != nil {
			t.Fatalf("Deactivate returned error: %v", err)
		}
		if deactivated {
			t.Fatalf("expected DeactivateProduct not to be called")
		}
		if resp.IsActive {
			t.Fatalf("expected product to remain inactive")
		}
	})
}

type fakeStore struct {
	createProductFn       func(context.Context, database.CreateProductParams) (database.Product, error)
	getProductByIDFn      func(context.Context, pgtype.UUID) (database.Product, error)
	getProductBySKUFn     func(context.Context, string) (database.Product, error)
	getProductByBarcodeFn func(context.Context, pgtype.Text) (database.Product, error)
	listProductsFn        func(context.Context, database.ListProductsParams) ([]database.Product, error)
	countProductsFn       func(context.Context, database.CountProductsParams) (int64, error)
	updateProductFn       func(context.Context, database.UpdateProductParams) (database.Product, error)
	activateProductFn     func(context.Context, pgtype.UUID) (database.Product, error)
	deactivateProductFn   func(context.Context, pgtype.UUID) (database.Product, error)
}

func (f *fakeStore) CreateProduct(ctx context.Context, arg database.CreateProductParams) (database.Product, error) {
	if f.createProductFn == nil {
		panic("unexpected CreateProduct call")
	}

	return f.createProductFn(ctx, arg)
}

func (f *fakeStore) GetProductByID(ctx context.Context, id pgtype.UUID) (database.Product, error) {
	if f.getProductByIDFn == nil {
		panic("unexpected GetProductByID call")
	}

	return f.getProductByIDFn(ctx, id)
}

func (f *fakeStore) GetProductBySKU(ctx context.Context, sku string) (database.Product, error) {
	if f.getProductBySKUFn == nil {
		panic("unexpected GetProductBySKU call")
	}

	return f.getProductBySKUFn(ctx, sku)
}

func (f *fakeStore) GetProductByBarcode(ctx context.Context, barcode pgtype.Text) (database.Product, error) {
	if f.getProductByBarcodeFn == nil {
		panic("unexpected GetProductByBarcode call")
	}

	return f.getProductByBarcodeFn(ctx, barcode)
}

func (f *fakeStore) ListProducts(ctx context.Context, arg database.ListProductsParams) ([]database.Product, error) {
	if f.listProductsFn == nil {
		panic("unexpected ListProducts call")
	}

	return f.listProductsFn(ctx, arg)
}

func (f *fakeStore) CountProducts(ctx context.Context, arg database.CountProductsParams) (int64, error) {
	if f.countProductsFn == nil {
		panic("unexpected CountProducts call")
	}

	return f.countProductsFn(ctx, arg)
}

func (f *fakeStore) UpdateProduct(ctx context.Context, arg database.UpdateProductParams) (database.Product, error) {
	if f.updateProductFn == nil {
		panic("unexpected UpdateProduct call")
	}

	return f.updateProductFn(ctx, arg)
}

func (f *fakeStore) ActivateProduct(ctx context.Context, id pgtype.UUID) (database.Product, error) {
	if f.activateProductFn == nil {
		panic("unexpected ActivateProduct call")
	}

	return f.activateProductFn(ctx, id)
}

func (f *fakeStore) DeactivateProduct(ctx context.Context, id pgtype.UUID) (database.Product, error) {
	if f.deactivateProductFn == nil {
		panic("unexpected DeactivateProduct call")
	}

	return f.deactivateProductFn(ctx, id)
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

func assertProductResponse(t *testing.T, resp products.ProductResponse) {
	t.Helper()

	if resp.ID == "" {
		t.Fatalf("expected ID to be set")
	}
	if resp.Name == "" {
		t.Fatalf("expected name to be set")
	}
	if resp.CreatedAt.IsZero() || resp.UpdatedAt.IsZero() {
		t.Fatalf("expected timestamps to be set")
	}
}

func requireValidationField(t *testing.T, err error, field string) {
	t.Helper()

	var validationErr *products.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected products.ValidationError, got %v", err)
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
