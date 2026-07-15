package products

import (
	"context"

	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/jackc/pgx/v5/pgtype"
)

type Store interface {
	CreateProduct(context.Context, database.CreateProductParams) (database.Product, error)
	GetProductByID(context.Context, pgtype.UUID) (database.Product, error)
	GetProductBySKU(context.Context, string) (database.Product, error)
	GetProductByBarcode(context.Context, pgtype.Text) (database.Product, error)
	ListProducts(context.Context, database.ListProductsParams) ([]database.Product, error)
	CountProducts(context.Context, database.CountProductsParams) (int64, error)
	UpdateProduct(context.Context, database.UpdateProductParams) (database.Product, error)
	ActivateProduct(context.Context, pgtype.UUID) (database.Product, error)
	DeactivateProduct(context.Context, pgtype.UUID) (database.Product, error)
}
