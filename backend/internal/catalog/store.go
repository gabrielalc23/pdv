package catalog

import (
	"context"

	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/jackc/pgx/v5/pgtype"
)

type Store interface {
	ListCatalogProducts(context.Context, database.ListCatalogProductsParams) ([]database.ListCatalogProductsRow, error)
	CountCatalogProducts(context.Context, database.CountCatalogProductsParams) (int64, error)
	GetCatalogProductByID(context.Context, pgtype.UUID) (database.GetCatalogProductByIDRow, error)
	GetCatalogProductByBarcode(context.Context, pgtype.Text) (database.GetCatalogProductByBarcodeRow, error)
}
