package categories

import (
	"context"
	"errors"
	"fmt"

	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/gabrielalc23/pdv/internal/platform/tenancy"
	"github.com/jackc/pgx/v5"
)

func (s *Service) Get(ctx context.Context, scope tenancy.OrganizationScope, rawID string) (CategoryResponse, error) {
	id, err := parseUUID(rawID)
	if err != nil {
		return CategoryResponse{}, err
	}

	row, err := s.store.GetCategoryByID(ctx, scope, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return CategoryResponse{}, ErrCategoryNotFound
		}
		return CategoryResponse{}, fmt.Errorf("get category: %w", err)
	}

	return toCategoryResponse(row.ID, row.Name, row.Slug, row.IsActive, row.CreatedAt, row.UpdatedAt), nil
}

func (s *Service) List(ctx context.Context, scope tenancy.OrganizationScope, input ListCategoriesInput) (CategoryListResponse, error) {
	rows, err := s.store.ListCategories(ctx, scope, database.ListCategoriesForOrganizationParams{
		Search:     optionalText(input.Search),
		ActiveOnly: input.ActiveOnly,
	})
	if err != nil {
		return CategoryListResponse{}, fmt.Errorf("list categories: %w", err)
	}

	data := make([]CategoryResponse, 0, len(rows))
	for _, row := range rows {
		data = append(data, toCategoryResponse(row.ID, row.Name, row.Slug, row.IsActive, row.CreatedAt, row.UpdatedAt))
	}

	return CategoryListResponse{Data: data}, nil
}
