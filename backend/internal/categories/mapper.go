package categories

import (
	"time"

	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/jackc/pgx/v5/pgtype"
)

func toCategoryResponse(category database.Category) CategoryResponse {
	return CategoryResponse{
		ID:        category.ID.String(),
		Name:      category.Name,
		Slug:      category.Slug,
		IsActive:  category.IsActive,
		CreatedAt: timestampOrZero(category.CreatedAt),
		UpdatedAt: timestampOrZero(category.UpdatedAt),
	}
}

func timestampOrZero(value pgtype.Timestamptz) time.Time {
	if !value.Valid {
		return time.Time{}
	}
	return value.Time.UTC()
}
