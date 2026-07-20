package categories

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func toCategoryResponse(id pgtype.UUID, name, slug string, isActive bool, createdAt, updatedAt pgtype.Timestamptz) CategoryResponse {
	return CategoryResponse{
		ID:        id.String(),
		Name:      name,
		Slug:      slug,
		IsActive:  isActive,
		CreatedAt: timestampOrZero(createdAt),
		UpdatedAt: timestampOrZero(updatedAt),
	}
}

func timestampOrZero(value pgtype.Timestamptz) time.Time {
	if !value.Valid {
		return time.Time{}
	}
	return value.Time.UTC()
}
