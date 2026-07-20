package categories

import (
	"context"
	"errors"
	"fmt"

	"github.com/gabrielalc23/pdv/internal/platform/authn"
	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/jackc/pgx/v5"
)

func (s *Service) Get(ctx context.Context, actor authn.OrganizationActor, rawID string) (CategoryResponse, error) {
	id, err := parseUUID(rawID)
	if err != nil {
		return CategoryResponse{}, err
	}

	scope := actor.ToOrganizationScope()
	row, err := s.store.GetCategoryByID(ctx, scope, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return CategoryResponse{}, ErrCategoryNotFound
		}
		return CategoryResponse{}, fmt.Errorf("get category: %w", err)
	}

	return toCategoryResponse(row.ID, row.Name, row.Slug, row.IsActive, row.CreatedAt, row.UpdatedAt), nil
}

func (s *Service) List(ctx context.Context, actor authn.OrganizationActor, input ListCategoriesInput) (CategoryListResponse, error) {
	scope := actor.ToOrganizationScope()
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
