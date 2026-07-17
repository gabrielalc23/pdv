package categories

import (
	"context"

	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/jackc/pgx/v5/pgtype"
)

type Store interface {
	CreateCategory(context.Context, database.CreateCategoryParams) (database.Category, error)
	GetCategoryByID(context.Context, pgtype.UUID) (database.Category, error)
	GetCategoryByName(context.Context, string) (database.Category, error)
	GetCategoryBySlug(context.Context, string) (database.Category, error)
	ListCategories(context.Context, database.ListCategoriesParams) ([]database.Category, error)
	UpdateCategory(context.Context, database.UpdateCategoryParams) (database.Category, error)
	ActivateCategory(context.Context, pgtype.UUID) (database.Category, error)
	DeactivateCategory(context.Context, pgtype.UUID) (database.Category, error)
}
