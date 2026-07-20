package catalog

import (
	"context"
	"errors"
	"fmt"

	"github.com/gabrielalc23/pdv/internal/platform/authn"
	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s *Service) List(ctx context.Context, actor authn.StoreActor, input ListCatalogInput) (CatalogListResponse, error) {
	scope := actor.ToStoreScope()

	page, pageSize, err := normalizePagination(input.Page, input.PageSize)
	if err != nil {
		return CatalogListResponse{}, err
	}

	search := normalizeOptionalSearch(input.Search)
	activeOnly := input.ActiveOnly
	if !input.ActiveOnlySet && !input.ActiveOnly {
		activeOnly = true
	}
	categoryID := pgtype.UUID{}
	if input.CategoryID != "" {
		categoryID, err = parseUUID(input.CategoryID, "categoryId")
		if err != nil {
			return CatalogListResponse{}, err
		}
	}

	total, err := s.store.CountCatalogProducts(ctx, scope, database.CountCatalogProductsForStoreParams{
		Search:      search,
		CategoryID:  categoryID,
		ActiveOnly:  activeOnly,
		InStockOnly: input.InStockOnly,
	})
	if err != nil {
		return CatalogListResponse{}, fmt.Errorf("count catalog products: %w", err)
	}

	rows, err := s.store.ListCatalogProducts(ctx, scope, database.ListCatalogProductsForStoreParams{
		Search:      search,
		CategoryID:  categoryID,
		ActiveOnly:  activeOnly,
		InStockOnly: input.InStockOnly,
		PageOffset:  int32((page - 1) * pageSize),
		PageSize:    int32(pageSize),
	})
	if err != nil {
		return CatalogListResponse{}, fmt.Errorf("list catalog products: %w", err)
	}

	items := make([]CatalogProductResponse, 0, len(rows))
	for _, row := range rows {
		item, err := ToCatalogProductResponse(toCatalogProductDataFromListRow(row))
		if err != nil {
			return CatalogListResponse{}, fmt.Errorf("map catalog product response: %w", err)
		}
		items = append(items, item)
	}

	return CatalogListResponse{
		Data:       items,
		Pagination: NewPaginationResponse(page, pageSize, total),
	}, nil
}

func (s *Service) GetByID(ctx context.Context, actor authn.StoreActor, rawID string) (CatalogProductResponse, error) {
	scope := actor.ToStoreScope()

	productID, err := parseUUID(rawID, "id")
	if err != nil {
		return CatalogProductResponse{}, err
	}

	row, err := s.store.GetCatalogProductByID(ctx, scope, productID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return CatalogProductResponse{}, ErrCatalogProductNotFound
		}
		return CatalogProductResponse{}, fmt.Errorf("get catalog product by id: %w", err)
	}

	return ToCatalogProductResponse(toCatalogProductDataFromIDRow(row))
}

func (s *Service) GetByBarcode(ctx context.Context, actor authn.StoreActor, rawBarcode string) (CatalogProductResponse, error) {
	scope := actor.ToStoreScope()

	barcode, err := normalizeRequiredText("barcode", rawBarcode)
	if err != nil {
		return CatalogProductResponse{}, err
	}

	row, err := s.store.GetCatalogProductByBarcode(ctx, scope, optionalText(barcode))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return CatalogProductResponse{}, ErrCatalogProductNotFound
		}
		return CatalogProductResponse{}, fmt.Errorf("get catalog product by barcode: %w", err)
	}

	return ToCatalogProductResponse(toCatalogProductDataFromBarcodeRow(row))
}
