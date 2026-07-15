package catalog

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestListCatalogDefaultPagination(t *testing.T) {
	var capturedCount database.CountCatalogProductsParams
	var capturedList database.ListCatalogProductsParams

	svc := NewService(&fakeStore{
		countCatalogProductsFn: func(_ context.Context, arg database.CountCatalogProductsParams) (int64, error) {
			capturedCount = arg
			return 1, nil
		},
		listCatalogProductsFn: func(_ context.Context, arg database.ListCatalogProductsParams) ([]database.ListCatalogProductsRow, error) {
			capturedList = arg
			return []database.ListCatalogProductsRow{catalogListRowFixture(true, "8.000")}, nil
		},
	})

	resp, err := svc.List(context.Background(), ListCatalogInput{})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}

	if resp.Pagination.Page != 1 || resp.Pagination.PageSize != 20 || resp.Pagination.Total != 1 || resp.Pagination.TotalPages != 1 {
		t.Fatalf("unexpected pagination: %+v", resp.Pagination)
	}
	if capturedList.PageSize != 20 || capturedList.PageOffset != 0 {
		t.Fatalf("unexpected list params: %+v", capturedList)
	}
	if !capturedCount.ActiveOnly {
		t.Fatalf("expected activeOnly true, got %#v", capturedCount.ActiveOnly)
	}
	if capturedCount.InStockOnly {
		t.Fatalf("expected inStockOnly false, got %#v", capturedCount.InStockOnly)
	}
	if capturedCount.Search.Valid {
		t.Fatalf("expected empty search param, got %#v", capturedCount.Search)
	}
}

func TestListCatalogSearchByName(t *testing.T) {
	var capturedCount database.CountCatalogProductsParams

	svc := NewService(&fakeStore{
		countCatalogProductsFn: func(_ context.Context, arg database.CountCatalogProductsParams) (int64, error) {
			capturedCount = arg
			return 0, nil
		},
		listCatalogProductsFn: func(context.Context, database.ListCatalogProductsParams) ([]database.ListCatalogProductsRow, error) {
			return []database.ListCatalogProductsRow{}, nil
		},
	})

	_, err := svc.List(context.Background(), ListCatalogInput{Search: " coca "})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}

	if !capturedCount.Search.Valid || capturedCount.Search.String != "coca" {
		t.Fatalf("unexpected search param: %#v", capturedCount.Search)
	}
}

func TestListCatalogSearchBySKU(t *testing.T) {
	svc := NewService(&fakeStore{
		countCatalogProductsFn: func(_ context.Context, arg database.CountCatalogProductsParams) (int64, error) {
			if !arg.Search.Valid || arg.Search.String != "coca-2l" {
				t.Fatalf("unexpected search param: %#v", arg.Search)
			}
			return 0, nil
		},
		listCatalogProductsFn: func(context.Context, database.ListCatalogProductsParams) ([]database.ListCatalogProductsRow, error) {
			return []database.ListCatalogProductsRow{}, nil
		},
	})

	_, err := svc.List(context.Background(), ListCatalogInput{Search: "coca-2l"})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
}

func TestListCatalogSearchByBarcode(t *testing.T) {
	svc := NewService(&fakeStore{
		countCatalogProductsFn: func(_ context.Context, arg database.CountCatalogProductsParams) (int64, error) {
			if !arg.Search.Valid || arg.Search.String != "7890000000000" {
				t.Fatalf("unexpected search param: %#v", arg.Search)
			}
			return 0, nil
		},
		listCatalogProductsFn: func(context.Context, database.ListCatalogProductsParams) ([]database.ListCatalogProductsRow, error) {
			return []database.ListCatalogProductsRow{}, nil
		},
	})

	_, err := svc.List(context.Background(), ListCatalogInput{Search: "7890000000000"})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
}

func TestListCatalogActiveOnlyFalse(t *testing.T) {
	var capturedCount database.CountCatalogProductsParams

	svc := NewService(&fakeStore{
		countCatalogProductsFn: func(_ context.Context, arg database.CountCatalogProductsParams) (int64, error) {
			capturedCount = arg
			return 0, nil
		},
		listCatalogProductsFn: func(context.Context, database.ListCatalogProductsParams) ([]database.ListCatalogProductsRow, error) {
			return []database.ListCatalogProductsRow{}, nil
		},
	})

	_, err := svc.List(context.Background(), ListCatalogInput{ActiveOnly: false, activeOnlySet: true})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}

	if capturedCount.ActiveOnly {
		t.Fatalf("expected activeOnly false, got %#v", capturedCount.ActiveOnly)
	}
}

func TestListCatalogInStockOnlyTrue(t *testing.T) {
	var capturedCount database.CountCatalogProductsParams

	svc := NewService(&fakeStore{
		countCatalogProductsFn: func(_ context.Context, arg database.CountCatalogProductsParams) (int64, error) {
			capturedCount = arg
			return 0, nil
		},
		listCatalogProductsFn: func(context.Context, database.ListCatalogProductsParams) ([]database.ListCatalogProductsRow, error) {
			return []database.ListCatalogProductsRow{}, nil
		},
	})

	_, err := svc.List(context.Background(), ListCatalogInput{InStockOnly: true})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}

	if !capturedCount.InStockOnly {
		t.Fatalf("expected inStockOnly true, got %#v", capturedCount.InStockOnly)
	}
}

func TestListCatalogProductWithoutInventoryUsesZeroQuantity(t *testing.T) {
	svc := NewService(&fakeStore{
		countCatalogProductsFn: func(context.Context, database.CountCatalogProductsParams) (int64, error) {
			return 1, nil
		},
		listCatalogProductsFn: func(context.Context, database.ListCatalogProductsParams) ([]database.ListCatalogProductsRow, error) {
			return []database.ListCatalogProductsRow{catalogListRowFixture(false, "0.000")}, nil
		},
	})

	resp, err := svc.List(context.Background(), ListCatalogInput{})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}

	if len(resp.Data) != 1 {
		t.Fatalf("unexpected data: %+v", resp.Data)
	}
	if resp.Data[0].Quantity != "0.000" {
		t.Fatalf("unexpected quantity: %s", resp.Data[0].Quantity)
	}
}

func TestListCatalogPageNormalization(t *testing.T) {
	page := 0
	pageSize := 101

	svc := NewService(&fakeStore{})

	_, err := svc.List(context.Background(), ListCatalogInput{Page: &page})
	requireValidationField(t, err, "page")

	_, err = svc.List(context.Background(), ListCatalogInput{PageSize: &pageSize})
	requireValidationField(t, err, "pageSize")
}

func TestListCatalogCountError(t *testing.T) {
	svc := NewService(&fakeStore{
		countCatalogProductsFn: func(context.Context, database.CountCatalogProductsParams) (int64, error) {
			return 0, errors.New("count failed")
		},
	})

	_, err := svc.List(context.Background(), ListCatalogInput{})
	if err == nil || err.Error() != "count catalog products: count failed" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListCatalogListError(t *testing.T) {
	svc := NewService(&fakeStore{
		countCatalogProductsFn: func(context.Context, database.CountCatalogProductsParams) (int64, error) {
			return 0, nil
		},
		listCatalogProductsFn: func(context.Context, database.ListCatalogProductsParams) ([]database.ListCatalogProductsRow, error) {
			return nil, errors.New("list failed")
		},
	})

	_, err := svc.List(context.Background(), ListCatalogInput{})
	if err == nil || err.Error() != "list catalog products: list failed" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListCatalogMapperError(t *testing.T) {
	svc := NewService(&fakeStore{
		countCatalogProductsFn: func(context.Context, database.CountCatalogProductsParams) (int64, error) {
			return 1, nil
		},
		listCatalogProductsFn: func(context.Context, database.ListCatalogProductsParams) ([]database.ListCatalogProductsRow, error) {
			return []database.ListCatalogProductsRow{{
				ID:        productFixture().ID,
				SKU:       "COCA-2L",
				Barcode:   pgtype.Text{String: "7890000000000", Valid: true},
				Name:      "Coca-Cola 2L",
				Price:     mustNumeric("12.90"),
				Quantity:  pgtype.Numeric{},
				IsActive:  true,
				InStock:   true,
				CreatedAt: mustTime("2026-07-15T10:00:00Z"),
				UpdatedAt: mustTime("2026-07-15T10:00:00Z"),
			}}, nil
		},
	})

	_, err := svc.List(context.Background(), ListCatalogInput{})
	if err == nil || err.Error() == "" {
		t.Fatalf("expected mapper error, got %v", err)
	}
}

func TestListCatalogEmptySliceNotNil(t *testing.T) {
	svc := NewService(&fakeStore{
		countCatalogProductsFn: func(context.Context, database.CountCatalogProductsParams) (int64, error) {
			return 0, nil
		},
		listCatalogProductsFn: func(context.Context, database.ListCatalogProductsParams) ([]database.ListCatalogProductsRow, error) {
			return []database.ListCatalogProductsRow{}, nil
		},
	})

	resp, err := svc.List(context.Background(), ListCatalogInput{})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if resp.Data == nil {
		t.Fatalf("expected empty slice, got nil")
	}
	if len(resp.Data) != 0 {
		t.Fatalf("expected empty slice, got %+v", resp.Data)
	}
}

func TestGetCatalogProductByID(t *testing.T) {
	svc := NewService(&fakeStore{
		getCatalogProductByIDFn: func(context.Context, pgtype.UUID) (database.GetCatalogProductByIDRow, error) {
			return catalogByIDRowFixture(true, "8.000"), nil
		},
	})

	resp, err := svc.GetByID(context.Background(), productFixture().ID.String())
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	assertCatalogProductResponse(t, resp)
}

func TestGetCatalogProductByIDInvalidUUID(t *testing.T) {
	svc := NewService(&fakeStore{})

	_, err := svc.GetByID(context.Background(), "not-a-uuid")
	requireValidationField(t, err, "id")
}

func TestGetCatalogProductByIDNotFound(t *testing.T) {
	svc := NewService(&fakeStore{
		getCatalogProductByIDFn: func(context.Context, pgtype.UUID) (database.GetCatalogProductByIDRow, error) {
			return database.GetCatalogProductByIDRow{}, pgx.ErrNoRows
		},
	})

	_, err := svc.GetByID(context.Background(), productFixture().ID.String())
	if !errors.Is(err, ErrCatalogProductNotFound) {
		t.Fatalf("expected ErrCatalogProductNotFound, got %v", err)
	}
}

func TestGetCatalogProductByIDDBError(t *testing.T) {
	svc := NewService(&fakeStore{
		getCatalogProductByIDFn: func(context.Context, pgtype.UUID) (database.GetCatalogProductByIDRow, error) {
			return database.GetCatalogProductByIDRow{}, errors.New("db failed")
		},
	})

	_, err := svc.GetByID(context.Background(), productFixture().ID.String())
	if err == nil || err.Error() != "get catalog product by id: db failed" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetCatalogProductByIDMapperError(t *testing.T) {
	svc := NewService(&fakeStore{
		getCatalogProductByIDFn: func(context.Context, pgtype.UUID) (database.GetCatalogProductByIDRow, error) {
			return database.GetCatalogProductByIDRow{
				ID:        productFixture().ID,
				SKU:       "COCA-2L",
				Barcode:   pgtype.Text{String: "7890000000000", Valid: true},
				Name:      "Coca-Cola 2L",
				Price:     mustNumeric("12.90"),
				Quantity:  pgtype.Numeric{},
				IsActive:  true,
				InStock:   true,
				CreatedAt: mustTime("2026-07-15T10:00:00Z"),
				UpdatedAt: mustTime("2026-07-15T10:00:00Z"),
			}, nil
		},
	})

	_, err := svc.GetByID(context.Background(), productFixture().ID.String())
	if err == nil {
		t.Fatalf("expected mapper error")
	}
}

func TestGetCatalogProductByBarcode(t *testing.T) {
	svc := NewService(&fakeStore{
		getCatalogProductByBarcodeFn: func(context.Context, pgtype.Text) (database.GetCatalogProductByBarcodeRow, error) {
			return catalogByBarcodeRowFixture(false, "0.000"), nil
		},
	})

	resp, err := svc.GetByBarcode(context.Background(), " 7890000000000 ")
	if err != nil {
		t.Fatalf("GetByBarcode returned error: %v", err)
	}
	if resp.Barcode == nil || *resp.Barcode != "7890000000000" {
		t.Fatalf("unexpected barcode: %#v", resp.Barcode)
	}
}

func TestGetCatalogProductByBarcodeEmpty(t *testing.T) {
	svc := NewService(&fakeStore{})

	_, err := svc.GetByBarcode(context.Background(), "   ")
	requireValidationField(t, err, "barcode")
}

func TestGetCatalogProductByBarcodeNotFound(t *testing.T) {
	svc := NewService(&fakeStore{
		getCatalogProductByBarcodeFn: func(context.Context, pgtype.Text) (database.GetCatalogProductByBarcodeRow, error) {
			return database.GetCatalogProductByBarcodeRow{}, pgx.ErrNoRows
		},
	})

	_, err := svc.GetByBarcode(context.Background(), "7890000000000")
	if !errors.Is(err, ErrCatalogProductNotFound) {
		t.Fatalf("expected ErrCatalogProductNotFound, got %v", err)
	}
}

func TestGetCatalogProductByBarcodeDBError(t *testing.T) {
	svc := NewService(&fakeStore{
		getCatalogProductByBarcodeFn: func(context.Context, pgtype.Text) (database.GetCatalogProductByBarcodeRow, error) {
			return database.GetCatalogProductByBarcodeRow{}, errors.New("db failed")
		},
	})

	_, err := svc.GetByBarcode(context.Background(), "7890000000000")
	if err == nil || err.Error() != "get catalog product by barcode: db failed" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetCatalogProductByBarcodeMapperError(t *testing.T) {
	svc := NewService(&fakeStore{
		getCatalogProductByBarcodeFn: func(context.Context, pgtype.Text) (database.GetCatalogProductByBarcodeRow, error) {
			return database.GetCatalogProductByBarcodeRow{
				ID:        productFixture().ID,
				SKU:       "COCA-2L",
				Barcode:   pgtype.Text{String: "7890000000000", Valid: true},
				Name:      "Coca-Cola 2L",
				Price:     mustNumeric("12.90"),
				Quantity:  pgtype.Numeric{},
				IsActive:  true,
				InStock:   true,
				CreatedAt: mustTime("2026-07-15T10:00:00Z"),
				UpdatedAt: mustTime("2026-07-15T10:00:00Z"),
			}, nil
		},
	})

	_, err := svc.GetByBarcode(context.Background(), "7890000000000")
	if err == nil {
		t.Fatalf("expected mapper error")
	}
}

func TestCatalogMapper(t *testing.T) {
	resp, err := toCatalogProductResponse(catalogProductData{
		ID:        productFixture().ID,
		SKU:       "COCA-2L",
		Barcode:   pgtype.Text{String: "7890000000000", Valid: true},
		Name:      "Coca-Cola 2L",
		Price:     mustNumeric("12.9"),
		Quantity:  mustNumeric("2.5"),
		IsActive:  true,
		InStock:   true,
		CreatedAt: mustTime("2026-07-15T10:00:00Z"),
		UpdatedAt: mustTime("2026-07-15T10:00:00Z"),
	})
	if err != nil {
		t.Fatalf("toCatalogProductResponse returned error: %v", err)
	}
	if resp.Price != "12.90" {
		t.Fatalf("unexpected price: %s", resp.Price)
	}
	if resp.Quantity != "2.500" {
		t.Fatalf("unexpected quantity: %s", resp.Quantity)
	}
	if resp.Barcode == nil || *resp.Barcode != "7890000000000" {
		t.Fatalf("unexpected barcode: %#v", resp.Barcode)
	}
	if !resp.InStock {
		t.Fatalf("expected inStock true")
	}
}

func TestCatalogMapperBarcodeNil(t *testing.T) {
	resp, err := toCatalogProductResponse(catalogProductData{
		ID:        productFixture().ID,
		SKU:       "COCA-2L",
		Name:      "Coca-Cola 2L",
		Price:     mustNumeric("12.90"),
		Quantity:  mustNumeric("0.000"),
		IsActive:  true,
		CreatedAt: mustTime("2026-07-15T10:00:00Z"),
		UpdatedAt: mustTime("2026-07-15T10:00:00Z"),
	})
	if err != nil {
		t.Fatalf("toCatalogProductResponse returned error: %v", err)
	}
	if resp.Barcode != nil {
		t.Fatalf("expected nil barcode, got %#v", resp.Barcode)
	}
	if resp.InStock {
		t.Fatalf("expected inStock false")
	}
}

func TestCatalogMapperQuantityNull(t *testing.T) {
	_, err := toCatalogProductResponse(catalogProductData{
		ID:        productFixture().ID,
		SKU:       "COCA-2L",
		Name:      "Coca-Cola 2L",
		Price:     mustNumeric("12.90"),
		Quantity:  pgtype.Numeric{},
		IsActive:  true,
		CreatedAt: mustTime("2026-07-15T10:00:00Z"),
		UpdatedAt: mustTime("2026-07-15T10:00:00Z"),
	})
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestCatalogMapperPriceNull(t *testing.T) {
	_, err := toCatalogProductResponse(catalogProductData{
		ID:        productFixture().ID,
		SKU:       "COCA-2L",
		Name:      "Coca-Cola 2L",
		Price:     pgtype.Numeric{},
		Quantity:  mustNumeric("1.000"),
		IsActive:  true,
		CreatedAt: mustTime("2026-07-15T10:00:00Z"),
		UpdatedAt: mustTime("2026-07-15T10:00:00Z"),
	})
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestNumericFormatting(t *testing.T) {
	price, err := numericToMoneyString(mustNumeric("10"))
	if err != nil || price != "10.00" {
		t.Fatalf("unexpected price: %s err=%v", price, err)
	}

	price, err = numericToMoneyString(mustNumeric("10.5"))
	if err != nil || price != "10.50" {
		t.Fatalf("unexpected price: %s err=%v", price, err)
	}

	quantity, err := numericToQuantityString(mustNumeric("2"))
	if err != nil || quantity != "2.000" {
		t.Fatalf("unexpected quantity: %s err=%v", quantity, err)
	}

	quantity, err = numericToQuantityString(mustNumeric("2.125"))
	if err != nil || quantity != "2.125" {
		t.Fatalf("unexpected quantity: %s err=%v", quantity, err)
	}
}

func TestPaginationResponse(t *testing.T) {
	resp := paginationResponse(2, 20, 21)
	if resp.TotalPages != 2 {
		t.Fatalf("unexpected total pages: %+v", resp)
	}
}

type fakeStore struct {
	listCatalogProductsFn        func(context.Context, database.ListCatalogProductsParams) ([]database.ListCatalogProductsRow, error)
	countCatalogProductsFn       func(context.Context, database.CountCatalogProductsParams) (int64, error)
	getCatalogProductByIDFn      func(context.Context, pgtype.UUID) (database.GetCatalogProductByIDRow, error)
	getCatalogProductByBarcodeFn func(context.Context, pgtype.Text) (database.GetCatalogProductByBarcodeRow, error)
}

func (f *fakeStore) ListCatalogProducts(ctx context.Context, arg database.ListCatalogProductsParams) ([]database.ListCatalogProductsRow, error) {
	if f.listCatalogProductsFn == nil {
		return []database.ListCatalogProductsRow{}, nil
	}
	return f.listCatalogProductsFn(ctx, arg)
}

func (f *fakeStore) CountCatalogProducts(ctx context.Context, arg database.CountCatalogProductsParams) (int64, error) {
	if f.countCatalogProductsFn == nil {
		return 0, nil
	}
	return f.countCatalogProductsFn(ctx, arg)
}

func (f *fakeStore) GetCatalogProductByID(ctx context.Context, id pgtype.UUID) (database.GetCatalogProductByIDRow, error) {
	if f.getCatalogProductByIDFn == nil {
		return database.GetCatalogProductByIDRow{}, pgx.ErrNoRows
	}
	return f.getCatalogProductByIDFn(ctx, id)
}

func (f *fakeStore) GetCatalogProductByBarcode(ctx context.Context, barcode pgtype.Text) (database.GetCatalogProductByBarcodeRow, error) {
	if f.getCatalogProductByBarcodeFn == nil {
		return database.GetCatalogProductByBarcodeRow{}, pgx.ErrNoRows
	}
	return f.getCatalogProductByBarcodeFn(ctx, barcode)
}

func catalogListRowFixture(inStock bool, quantity string) database.ListCatalogProductsRow {
	return database.ListCatalogProductsRow{
		ID:        productFixture().ID,
		SKU:       "COCA-2L",
		Barcode:   pgtype.Text{String: "7890000000000", Valid: true},
		Name:      "Coca-Cola 2L",
		Price:     mustNumeric("12.90"),
		Quantity:  mustNumeric(quantity),
		IsActive:  true,
		InStock:   inStock,
		CreatedAt: mustTime("2026-07-15T10:00:00Z"),
		UpdatedAt: mustTime("2026-07-15T10:00:00Z"),
	}
}

func catalogByIDRowFixture(inStock bool, quantity string) database.GetCatalogProductByIDRow {
	return database.GetCatalogProductByIDRow{
		ID:        productFixture().ID,
		SKU:       "COCA-2L",
		Barcode:   pgtype.Text{String: "7890000000000", Valid: true},
		Name:      "Coca-Cola 2L",
		Price:     mustNumeric("12.90"),
		Quantity:  mustNumeric(quantity),
		IsActive:  true,
		InStock:   inStock,
		CreatedAt: mustTime("2026-07-15T10:00:00Z"),
		UpdatedAt: mustTime("2026-07-15T10:00:00Z"),
	}
}

func catalogByBarcodeRowFixture(inStock bool, quantity string) database.GetCatalogProductByBarcodeRow {
	return database.GetCatalogProductByBarcodeRow{
		ID:        productFixture().ID,
		SKU:       "COCA-2L",
		Barcode:   pgtype.Text{String: "7890000000000", Valid: true},
		Name:      "Coca-Cola 2L",
		Price:     mustNumeric("12.90"),
		Quantity:  mustNumeric(quantity),
		IsActive:  true,
		InStock:   inStock,
		CreatedAt: mustTime("2026-07-15T10:00:00Z"),
		UpdatedAt: mustTime("2026-07-15T10:00:00Z"),
	}
}

func assertCatalogProductResponse(t *testing.T, resp CatalogProductResponse) {
	t.Helper()

	if resp.ID == "" || resp.SKU == "" || resp.Name == "" {
		t.Fatalf("unexpected response: %+v", resp)
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

func productFixture() database.Product {
	return database.Product{
		ID:        mustUUID("01972d6b-bf3a-7f1f-a4f8-1d2f31c3b8a9"),
		SKU:       "COCA-2L",
		Barcode:   pgtype.Text{String: "7890000000000", Valid: true},
		Name:      "Coca-Cola 2L",
		Price:     mustNumeric("12.90"),
		Cost:      mustNumeric("8.50"),
		IsActive:  true,
		CreatedAt: mustTime("2026-07-15T10:00:00Z"),
		UpdatedAt: mustTime("2026-07-15T10:00:00Z"),
	}
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
