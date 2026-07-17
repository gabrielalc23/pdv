package products

import (
	"context"
	"fmt"

	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s *Service) Get(
	ctx context.Context,
	id string,
) (ProductResponse, error) {
	product, err := s.productByID(ctx, id)
	if err != nil {
		return ProductResponse{}, err
	}

	return toProductResponse(product)
}

func (s *Service) List(
	ctx context.Context,
	input ListProductsInput,
) (ProductListResponse, error) {
	page, pageSize, err := normalizePagination(
		input.Page,
		input.PageSize,
	)
	if err != nil {
		return ProductListResponse{}, err
	}

	search := optionalText(input.Search)
	categoryID := pgtype.UUID{}
	if input.CategoryID != "" {
		categoryID, err = parseUUID(input.CategoryID, "categoryId")
		if err != nil {
			return ProductListResponse{}, err
		}
	}

	total, err := s.store.CountProducts(
		ctx,
		database.CountProductsParams{
			Search:     search,
			CategoryID: categoryID,
			ActiveOnly: input.ActiveOnly,
		},
	)
	if err != nil {
		return ProductListResponse{},
			fmt.Errorf("count products: %w", err)
	}

	items, err := s.store.ListProducts(
		ctx,
		database.ListProductsParams{
			Search:     search,
			CategoryID: categoryID,
			ActiveOnly: input.ActiveOnly,
			PageOffset: int32((page - 1) * pageSize),
			PageSize:   int32(pageSize),
		},
	)
	if err != nil {
		return ProductListResponse{},
			fmt.Errorf("list products: %w", err)
	}

	data := make([]ProductResponse, 0, len(items))
	for _, item := range items {
		response, err := toProductResponse(item)
		if err != nil {
			return ProductListResponse{},
				fmt.Errorf("map product response: %w", err)
		}

		data = append(data, response)
	}

	return ProductListResponse{
		Data:       data,
		Pagination: paginationResponse(page, pageSize, total),
	}, nil
}

func (s *Service) productByID(
	ctx context.Context,
	rawID string,
) (database.Product, error) {
	productID, err := parseUUID(rawID, "id")
	if err != nil {
		return database.Product{}, err
	}

	return s.getProductByID(ctx, productID)
}

func (s *Service) getProductByID(
	ctx context.Context,
	id pgtype.UUID,
) (database.Product, error) {
	product, err := s.store.GetProductByID(ctx, id)
	if err != nil {
		return database.Product{}, translatePersistenceError(err)
	}

	return product, nil
}
