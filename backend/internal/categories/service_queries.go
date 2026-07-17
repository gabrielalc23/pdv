package categories

import (
	"context"
	"errors"
	"fmt"

	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/jackc/pgx/v5"
)

func (s *Service) Get(ctx context.Context, rawID string) (CategoryResponse, error) {
	id, err := parseUUID(rawID)
	if err != nil {
		return CategoryResponse{}, err
	}

	category, err := s.store.GetCategoryByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return CategoryResponse{}, ErrCategoryNotFound
		}
		return CategoryResponse{}, fmt.Errorf("get category: %w", err)
	}

	return toCategoryResponse(category), nil
}

func (s *Service) List(ctx context.Context, input ListCategoriesInput) (CategoryListResponse, error) {
	rows, err := s.store.ListCategories(ctx, database.ListCategoriesParams{
		Search:     optionalText(input.Search),
		ActiveOnly: input.ActiveOnly,
	})
	if err != nil {
		return CategoryListResponse{}, fmt.Errorf("list categories: %w", err)
	}

	data := make([]CategoryResponse, 0, len(rows))
	for _, row := range rows {
		data = append(data, toCategoryResponse(row))
	}

	return CategoryListResponse{Data: data}, nil
}
