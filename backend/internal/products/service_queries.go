package products

import (
	"context"
	"errors"
	"fmt"

	"github.com/gabrielalc23/pdv/internal/platform/authn"
	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/gabrielalc23/pdv/internal/platform/tenancy"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s *Service) Get(ctx context.Context, actor authn.OrganizationActor, id string) (ProductResponse, error) {
	scope := actor.ToOrganizationScope()

	product, err := s.productByID(ctx, scope, id)
	if err != nil {
		return ProductResponse{}, err
	}
	return toProductResponse(product)
}

func (s *Service) List(ctx context.Context, actor authn.OrganizationActor, input ListProductsInput) (ProductListResponse, error) {
	scope := actor.ToOrganizationScope()

	page, pageSize, err := normalizePagination(input.Page, input.PageSize)
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

	total, err := s.store.CountProducts(ctx, scope, database.CountProductsForOrganizationParams{
		Search:     search,
		CategoryID: categoryID,
		ActiveOnly: input.ActiveOnly,
	})
	if err != nil {
		return ProductListResponse{}, fmt.Errorf("count products: %w", err)
	}

	items, err := s.store.ListProducts(ctx, scope, database.ListProductsForOrganizationParams{
		Search:     search,
		CategoryID: categoryID,
		ActiveOnly: input.ActiveOnly,
		PageOffset: int32((page - 1) * pageSize),
		PageSize:   int32(pageSize),
	})
	if err != nil {
		return ProductListResponse{}, fmt.Errorf("list products: %w", err)
	}

	data := make([]ProductResponse, 0, len(items))
	for _, item := range items {
		response, err := toProductResponse(productFromRow(item.ID, item.SKU, item.Barcode, item.Name, item.CategoryID, item.Price, item.Cost, item.IsActive, item.CreatedAt, item.UpdatedAt))
		if err != nil {
			return ProductListResponse{}, fmt.Errorf("map product response: %w", err)
		}
		data = append(data, response)
	}

	return ProductListResponse{
		Data:       data,
		Pagination: paginationResponse(page, pageSize, total),
	}, nil
}

func (s *Service) productByID(ctx context.Context, scope tenancy.OrganizationScope, rawID string) (productProjection, error) {
	productID, err := parseUUID(rawID, "id")
	if err != nil {
		return productProjection{}, err
	}
	return s.getProductByID(ctx, scope, productID)
}

func (s *Service) getProductByID(ctx context.Context, scope tenancy.OrganizationScope, id pgtype.UUID) (productProjection, error) {
	row, err := s.store.GetProductByID(ctx, scope, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return productProjection{}, ErrProductNotFound
		}
		return productProjection{}, translatePersistenceError(err)
	}

	return productFromRow(row.ID, row.SKU, row.Barcode, row.Name, row.CategoryID, row.Price, row.Cost, row.IsActive, row.CreatedAt, row.UpdatedAt), nil
}
