package products_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/gabrielalc23/pdv/internal/platform/tenancy"
	"github.com/gabrielalc23/pdv/internal/products"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

var testOrgID = mustParseUUID("00000000-0000-0000-0000-000000000001")
var testScope = tenancy.OrganizationScope{OrganizationID: testOrgID}

func mustParseUUID(s string) pgtype.UUID {
	var id pgtype.UUID
	if err := id.Scan(s); err != nil {
		panic(err)
	}
	return id
}

func strPtr(s string) *string {
	return &s
}

func mustNumeric(s string) pgtype.Numeric {
	var n pgtype.Numeric
	if err := n.Scan(s); err != nil {
		panic(err)
	}
	return n
}

func productRowFixture(includePricing bool) database.CreateProductForOrganizationRow {
	row := database.CreateProductForOrganizationRow{
		OrganizationID: testOrgID,
		ID:             mustParseUUID("00000000-0000-0000-0000-000000000002"),
		SKU:            "COCA-2L",
		Barcode:        pgtype.Text{String: "7890000000000", Valid: true},
		Name:           "Coca-Cola 2L",
		CategoryID:     pgtype.UUID{},
		IsActive:       true,
		CreatedAt:      pgtype.Timestamptz{Time: time.Now(), Valid: true},
		UpdatedAt:      pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}
	if includePricing {
		row.Price = mustNumeric("12.90")
		row.Cost = mustNumeric("8.50")
	}
	return row
}

func assertProductResponse(t *testing.T, resp products.ProductResponse) {
	t.Helper()
	if resp.ID == "" {
		t.Fatal("expected non-empty ID")
	}
	if resp.Name == "" {
		t.Fatal("expected non-empty Name")
	}
	if resp.Price == "" {
		t.Fatal("expected non-empty Price")
	}
}

type fakeStore struct {
	createProductFn       func(context.Context, tenancy.OrganizationScope, database.CreateProductForOrganizationParams) (database.CreateProductForOrganizationRow, error)
	getProductByIDFn      func(context.Context, tenancy.OrganizationScope, pgtype.UUID) (database.GetProductByIDForOrganizationRow, error)
	getProductBySKUFn     func(context.Context, tenancy.OrganizationScope, string) (database.GetProductBySKUForOrganizationRow, error)
	getProductByBarcodeFn func(context.Context, tenancy.OrganizationScope, pgtype.Text) (database.GetProductByBarcodeForOrganizationRow, error)
	listProductsFn        func(context.Context, tenancy.OrganizationScope, database.ListProductsForOrganizationParams) ([]database.ListProductsForOrganizationRow, error)
	countProductsFn       func(context.Context, tenancy.OrganizationScope, database.CountProductsForOrganizationParams) (int64, error)
	updateProductFn       func(context.Context, tenancy.OrganizationScope, database.UpdateProductForOrganizationParams) (database.UpdateProductForOrganizationRow, error)
	activateProductFn     func(context.Context, tenancy.OrganizationScope, pgtype.UUID) (database.ActivateProductForOrganizationRow, error)
	deactivateProductFn   func(context.Context, tenancy.OrganizationScope, pgtype.UUID) (database.DeactivateProductForOrganizationRow, error)
}

func (s *fakeStore) CreateProduct(ctx context.Context, scope tenancy.OrganizationScope, params database.CreateProductForOrganizationParams) (database.CreateProductForOrganizationRow, error) {
	return s.createProductFn(ctx, scope, params)
}

func (s *fakeStore) GetProductByID(ctx context.Context, scope tenancy.OrganizationScope, id pgtype.UUID) (database.GetProductByIDForOrganizationRow, error) {
	return s.getProductByIDFn(ctx, scope, id)
}

func (s *fakeStore) GetProductBySKU(ctx context.Context, scope tenancy.OrganizationScope, sku string) (database.GetProductBySKUForOrganizationRow, error) {
	return s.getProductBySKUFn(ctx, scope, sku)
}

func (s *fakeStore) GetProductByBarcode(ctx context.Context, scope tenancy.OrganizationScope, barcode pgtype.Text) (database.GetProductByBarcodeForOrganizationRow, error) {
	return s.getProductByBarcodeFn(ctx, scope, barcode)
}

func (s *fakeStore) ListProducts(ctx context.Context, scope tenancy.OrganizationScope, params database.ListProductsForOrganizationParams) ([]database.ListProductsForOrganizationRow, error) {
	return s.listProductsFn(ctx, scope, params)
}

func (s *fakeStore) CountProducts(ctx context.Context, scope tenancy.OrganizationScope, params database.CountProductsForOrganizationParams) (int64, error) {
	return s.countProductsFn(ctx, scope, params)
}

func (s *fakeStore) UpdateProduct(ctx context.Context, scope tenancy.OrganizationScope, params database.UpdateProductForOrganizationParams) (database.UpdateProductForOrganizationRow, error) {
	return s.updateProductFn(ctx, scope, params)
}

func (s *fakeStore) ActivateProduct(ctx context.Context, scope tenancy.OrganizationScope, id pgtype.UUID) (database.ActivateProductForOrganizationRow, error) {
	return s.activateProductFn(ctx, scope, id)
}

func (s *fakeStore) DeactivateProduct(ctx context.Context, scope tenancy.OrganizationScope, id pgtype.UUID) (database.DeactivateProductForOrganizationRow, error) {
	return s.deactivateProductFn(ctx, scope, id)
}

func TestCreateProductValid(t *testing.T) {
	store := &fakeStore{
		getProductBySKUFn: func(_ context.Context, _ tenancy.OrganizationScope, _ string) (database.GetProductBySKUForOrganizationRow, error) {
			return database.GetProductBySKUForOrganizationRow{}, pgx.ErrNoRows
		},
		getProductByBarcodeFn: func(_ context.Context, _ tenancy.OrganizationScope, _ pgtype.Text) (database.GetProductByBarcodeForOrganizationRow, error) {
			return database.GetProductByBarcodeForOrganizationRow{}, pgx.ErrNoRows
		},
		createProductFn: func(ctx context.Context, scope tenancy.OrganizationScope, arg database.CreateProductForOrganizationParams) (database.CreateProductForOrganizationRow, error) {
			if arg.SKU != "COCA-2L" {
				t.Fatalf("unexpected SKU: %s", arg.SKU)
			}
			if !arg.Barcode.Valid || arg.Barcode.String != "7890000000000" {
				t.Fatalf("unexpected barcode: %+v", arg.Barcode)
			}
			if !arg.Price.Valid || !arg.Cost.Valid {
				t.Fatalf("expected price and cost to be valid")
			}
			return productRowFixture(true), nil
		},
	}

	svc := products.NewService(store)

	resp, err := svc.Create(context.Background(), testScope, products.UpsertProductInput{
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

func TestCreateProductDuplicateSKU(t *testing.T) {
	store := &fakeStore{
		getProductBySKUFn: func(_ context.Context, _ tenancy.OrganizationScope, sku string) (database.GetProductBySKUForOrganizationRow, error) {
			return database.GetProductBySKUForOrganizationRow{
				ID:  mustParseUUID("00000000-0000-0000-0000-000000000999"),
				SKU: sku,
			}, nil
		},
		createProductFn: func(_ context.Context, _ tenancy.OrganizationScope, _ database.CreateProductForOrganizationParams) (database.CreateProductForOrganizationRow, error) {
			t.Fatal("CreateProduct should not be called")
			return database.CreateProductForOrganizationRow{}, nil
		},
	}

	svc := products.NewService(store)
	_, err := svc.Create(context.Background(), testScope, products.UpsertProductInput{
		SKU:   "COCA-2L",
		Name:  "Coca-Cola 2L",
		Price: "12.90",
	})
	if err == nil {
		t.Fatal("expected error for duplicate SKU")
	}
	if !errors.Is(err, products.ErrSKUAlreadyExists) {
		t.Fatalf("expected ErrSKUAlreadyExists, got: %v", err)
	}
}

func TestCreateProductDuplicateBarcode(t *testing.T) {
	store := &fakeStore{
		getProductBySKUFn: func(_ context.Context, _ tenancy.OrganizationScope, _ string) (database.GetProductBySKUForOrganizationRow, error) {
			return database.GetProductBySKUForOrganizationRow{}, pgx.ErrNoRows
		},
		getProductByBarcodeFn: func(_ context.Context, _ tenancy.OrganizationScope, _ pgtype.Text) (database.GetProductByBarcodeForOrganizationRow, error) {
			return database.GetProductByBarcodeForOrganizationRow{
				ID:      mustParseUUID("00000000-0000-0000-0000-000000000999"),
				Barcode: pgtype.Text{String: "7890000000000", Valid: true},
			}, nil
		},
		createProductFn: func(_ context.Context, _ tenancy.OrganizationScope, _ database.CreateProductForOrganizationParams) (database.CreateProductForOrganizationRow, error) {
			t.Fatal("CreateProduct should not be called")
			return database.CreateProductForOrganizationRow{}, nil
		},
	}

	svc := products.NewService(store)
	_, err := svc.Create(context.Background(), testScope, products.UpsertProductInput{
		SKU:     "COCA-2L",
		Barcode: strPtr("7890000000000"),
		Name:    "Coca-Cola 2L",
		Price:   "12.90",
	})
	if err == nil {
		t.Fatal("expected error for duplicate barcode")
	}
	if !errors.Is(err, products.ErrBarcodeAlreadyExists) {
		t.Fatalf("expected ErrBarcodeAlreadyExists, got: %v", err)
	}
}

func TestCreateProductInvalidInput(t *testing.T) {
	store := &fakeStore{}
	svc := products.NewService(store)

	tests := []struct {
		name  string
		input products.UpsertProductInput
	}{
		{"empty SKU", products.UpsertProductInput{SKU: "", Name: "Test", Price: "10.00"}},
		{"empty Name", products.UpsertProductInput{SKU: "TEST", Name: "", Price: "10.00"}},
		{"invalid Price", products.UpsertProductInput{SKU: "TEST", Name: "Test", Price: "abc"}},
		{"negative Price", products.UpsertProductInput{SKU: "TEST", Name: "Test", Price: "-10.00"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.Create(context.Background(), testScope, tc.input)
			if err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestGetProduct(t *testing.T) {
	expectedID := mustParseUUID("00000000-0000-0000-0000-000000000002")
	store := &fakeStore{
		getProductByIDFn: func(_ context.Context, _ tenancy.OrganizationScope, id pgtype.UUID) (database.GetProductByIDForOrganizationRow, error) {
			if id != expectedID {
				t.Fatalf("unexpected id: %v", id)
			}
			return database.GetProductByIDForOrganizationRow{
				OrganizationID: testOrgID,
				ID:             expectedID,
				SKU:            "COCA-2L",
				Barcode:        pgtype.Text{String: "7890000000000", Valid: true},
				Name:           "Coca-Cola 2L",
				Price:          mustNumeric("12.90"),
				Cost:           mustNumeric("8.50"),
				IsActive:       true,
				CreatedAt:      pgtype.Timestamptz{Time: time.Now(), Valid: true},
				UpdatedAt:      pgtype.Timestamptz{Time: time.Now(), Valid: true},
			}, nil
		},
	}
	svc := products.NewService(store)

	resp, err := svc.Get(context.Background(), testScope, expectedID.String())
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if resp.SKU != "COCA-2L" {
		t.Fatalf("unexpected SKU: %s", resp.SKU)
	}
}

func TestGetProductNotFound(t *testing.T) {
	store := &fakeStore{
		getProductByIDFn: func(_ context.Context, _ tenancy.OrganizationScope, _ pgtype.UUID) (database.GetProductByIDForOrganizationRow, error) {
			return database.GetProductByIDForOrganizationRow{}, pgx.ErrNoRows
		},
	}
	svc := products.NewService(store)

	_, err := svc.Get(context.Background(), testScope, "00000000-0000-0000-0000-000000000999")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !errors.Is(err, products.ErrProductNotFound) {
		t.Fatalf("expected ErrProductNotFound, got: %v", err)
	}
}

func TestUpdateProduct(t *testing.T) {
	existingID := mustParseUUID("00000000-0000-0000-0000-000000000002")
	store := &fakeStore{
		getProductByIDFn: func(_ context.Context, _ tenancy.OrganizationScope, id pgtype.UUID) (database.GetProductByIDForOrganizationRow, error) {
			return database.GetProductByIDForOrganizationRow{
				OrganizationID: testOrgID,
				ID:             id,
				SKU:            "COCA-2L",
				Barcode:        pgtype.Text{String: "7890000000000", Valid: true},
				Name:           "Coca-Cola 2L",
				Price:          mustNumeric("12.90"),
				Cost:           mustNumeric("8.50"),
				IsActive:       true,
				CreatedAt:      pgtype.Timestamptz{Time: time.Now(), Valid: true},
				UpdatedAt:      pgtype.Timestamptz{Time: time.Now(), Valid: true},
			}, nil
		},
		getProductBySKUFn: func(_ context.Context, _ tenancy.OrganizationScope, _ string) (database.GetProductBySKUForOrganizationRow, error) {
			return database.GetProductBySKUForOrganizationRow{}, pgx.ErrNoRows
		},
		getProductByBarcodeFn: func(_ context.Context, _ tenancy.OrganizationScope, _ pgtype.Text) (database.GetProductByBarcodeForOrganizationRow, error) {
			return database.GetProductByBarcodeForOrganizationRow{}, pgx.ErrNoRows
		},
		updateProductFn: func(ctx context.Context, scope tenancy.OrganizationScope, params database.UpdateProductForOrganizationParams) (database.UpdateProductForOrganizationRow, error) {
			return database.UpdateProductForOrganizationRow{
				OrganizationID: testOrgID,
				ID:             params.ID,
				SKU:            params.SKU,
				Barcode:        params.Barcode,
				Name:           params.Name,
				Price:          params.Price,
				Cost:           params.Cost,
				IsActive:       true,
				CreatedAt:      pgtype.Timestamptz{Time: time.Now(), Valid: true},
				UpdatedAt:      pgtype.Timestamptz{Time: time.Now(), Valid: true},
			}, nil
		},
	}
	svc := products.NewService(store)

	resp, err := svc.Update(context.Background(), testScope, existingID.String(), products.UpsertProductInput{
		SKU:     "COCA-2L",
		Barcode: strPtr("7890000000000"),
		Name:    "Coca-Cola 2L Updated",
		Price:   "15.90",
		Cost:    strPtr("10.00"),
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if resp.Name != "Coca-Cola 2L Updated" {
		t.Fatalf("unexpected name: %s", resp.Name)
	}
	if resp.Price != "15.90" {
		t.Fatalf("unexpected price: %s", resp.Price)
	}
}

func TestActivateProduct(t *testing.T) {
	productID := mustParseUUID("00000000-0000-0000-0000-000000000002")
	store := &fakeStore{
		getProductByIDFn: func(_ context.Context, _ tenancy.OrganizationScope, id pgtype.UUID) (database.GetProductByIDForOrganizationRow, error) {
			return database.GetProductByIDForOrganizationRow{
				OrganizationID: testOrgID,
				ID:             id,
				SKU:            "COCA-2L",
				Name:           "Coca-Cola 2L",
				Price:          mustNumeric("12.90"),
				IsActive:       false,
				CreatedAt:      pgtype.Timestamptz{Time: time.Now(), Valid: true},
				UpdatedAt:      pgtype.Timestamptz{Time: time.Now(), Valid: true},
			}, nil
		},
		activateProductFn: func(_ context.Context, _ tenancy.OrganizationScope, id pgtype.UUID) (database.ActivateProductForOrganizationRow, error) {
			return database.ActivateProductForOrganizationRow{
				OrganizationID: testOrgID,
				ID:             id,
				SKU:            "COCA-2L",
				Name:           "Coca-Cola 2L",
				Price:          mustNumeric("12.90"),
				IsActive:       true,
				CreatedAt:      pgtype.Timestamptz{Time: time.Now(), Valid: true},
				UpdatedAt:      pgtype.Timestamptz{Time: time.Now(), Valid: true},
			}, nil
		},
	}
	svc := products.NewService(store)

	resp, err := svc.Activate(context.Background(), testScope, productID.String())
	if err != nil {
		t.Fatalf("Activate returned error: %v", err)
	}
	if !resp.IsActive {
		t.Fatal("expected product to be active")
	}
}

func TestDeactivateProduct(t *testing.T) {
	productID := mustParseUUID("00000000-0000-0000-0000-000000000002")
	store := &fakeStore{
		getProductByIDFn: func(_ context.Context, _ tenancy.OrganizationScope, id pgtype.UUID) (database.GetProductByIDForOrganizationRow, error) {
			return database.GetProductByIDForOrganizationRow{
				OrganizationID: testOrgID,
				ID:             id,
				SKU:            "COCA-2L",
				Name:           "Coca-Cola 2L",
				Price:          mustNumeric("12.90"),
				IsActive:       true,
				CreatedAt:      pgtype.Timestamptz{Time: time.Now(), Valid: true},
				UpdatedAt:      pgtype.Timestamptz{Time: time.Now(), Valid: true},
			}, nil
		},
		deactivateProductFn: func(_ context.Context, _ tenancy.OrganizationScope, id pgtype.UUID) (database.DeactivateProductForOrganizationRow, error) {
			return database.DeactivateProductForOrganizationRow{
				OrganizationID: testOrgID,
				ID:             id,
				SKU:            "COCA-2L",
				Name:           "Coca-Cola 2L",
				Price:          mustNumeric("12.90"),
				IsActive:       false,
				CreatedAt:      pgtype.Timestamptz{Time: time.Now(), Valid: true},
				UpdatedAt:      pgtype.Timestamptz{Time: time.Now(), Valid: true},
			}, nil
		},
	}
	svc := products.NewService(store)

	resp, err := svc.Deactivate(context.Background(), testScope, productID.String())
	if err != nil {
		t.Fatalf("Deactivate returned error: %v", err)
	}
	if resp.IsActive {
		t.Fatal("expected product to be inactive")
	}
}
