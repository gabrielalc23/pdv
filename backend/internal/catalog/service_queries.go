package catalog

import (
	"context"
	"errors"
	"fmt"

	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/jackc/pgx/v5"
)

func (s *Service) List(ctx context.Context, input ListCatalogInput) (CatalogListResponse, error) {
	page, pageSize, err := normalizePagination(input.Page, input.PageSize)
	if err != nil {
		return CatalogListResponse{}, err
	}

	search := normalizeOptionalSearch(input.Search)
	activeOnly := input.ActiveOnly
	if !input.ActiveOnlySet && !input.ActiveOnly {
		activeOnly = true
	}

	total, err := s.store.CountCatalogProducts(ctx, database.CountCatalogProductsParams{
		Search:      search,
		ActiveOnly:  activeOnly,
		InStockOnly: input.InStockOnly,
	})
	if err != nil {
		return CatalogListResponse{}, fmt.Errorf("count catalog products: %w", err)
	}

	rows, err := s.store.ListCatalogProducts(ctx, database.ListCatalogProductsParams{
		Search:      search,
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

func (s *Service) GetByID(ctx context.Context, rawID string) (CatalogProductResponse, error) {
	productID, err := parseUUID(rawID, "id")
	if err != nil {
		return CatalogProductResponse{}, err
	}

	row, err := s.store.GetCatalogProductByID(ctx, productID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return CatalogProductResponse{}, ErrCatalogProductNotFound
		}
		return CatalogProductResponse{}, fmt.Errorf("get catalog product by id: %w", err)
	}

	return ToCatalogProductResponse(toCatalogProductDataFromIDRow(row))
}

func (s *Service) GetByBarcode(ctx context.Context, rawBarcode string) (CatalogProductResponse, error) {
	barcode, err := normalizeRequiredText("barcode", rawBarcode)
	if err != nil {
		return CatalogProductResponse{}, err
	}

	row, err := s.store.GetCatalogProductByBarcode(ctx, optionalText(barcode))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return CatalogProductResponse{}, ErrCatalogProductNotFound
		}
		return CatalogProductResponse{}, fmt.Errorf("get catalog product by barcode: %w", err)
	}

	return ToCatalogProductResponse(toCatalogProductDataFromBarcodeRow(row))
}
