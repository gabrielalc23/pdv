package catalog

import (
	"context"

	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/gabrielalc23/pdv/internal/platform/tenancy"
	"github.com/jackc/pgx/v5/pgtype"
)

type Store interface {
	CountCatalogProducts(ctx context.Context, scope tenancy.StoreScope, params database.CountCatalogProductsForStoreParams) (int64, error)
	ListCatalogProducts(ctx context.Context, scope tenancy.StoreScope, params database.ListCatalogProductsForStoreParams) ([]database.ListCatalogProductsForStoreRow, error)
	GetCatalogProductByID(ctx context.Context, scope tenancy.StoreScope, id pgtype.UUID) (database.GetCatalogProductByIDForStoreRow, error)
	GetCatalogProductByBarcode(ctx context.Context, scope tenancy.StoreScope, barcode pgtype.Text) (database.GetCatalogProductByBarcodeForStoreRow, error)
}

type storeImpl struct {
	q *database.Queries
}

func NewStore(q *database.Queries) Store {
	return &storeImpl{q: q}
}

func (s *storeImpl) CountCatalogProducts(ctx context.Context, scope tenancy.StoreScope, params database.CountCatalogProductsForStoreParams) (int64, error) {
	params.OrganizationID = scope.OrganizationID
	params.StoreID = scope.StoreID
	return s.q.CountCatalogProductsForStore(ctx, params)
}

func (s *storeImpl) ListCatalogProducts(ctx context.Context, scope tenancy.StoreScope, params database.ListCatalogProductsForStoreParams) ([]database.ListCatalogProductsForStoreRow, error) {
	params.OrganizationID = scope.OrganizationID
	params.StoreID = scope.StoreID
	return s.q.ListCatalogProductsForStore(ctx, params)
}

func (s *storeImpl) GetCatalogProductByID(ctx context.Context, scope tenancy.StoreScope, id pgtype.UUID) (database.GetCatalogProductByIDForStoreRow, error) {
	return s.q.GetCatalogProductByIDForStore(ctx, database.GetCatalogProductByIDForStoreParams{
		OrganizationID: scope.OrganizationID,
		StoreID:        scope.StoreID,
		ID:             id,
	})
}

func (s *storeImpl) GetCatalogProductByBarcode(ctx context.Context, scope tenancy.StoreScope, barcode pgtype.Text) (database.GetCatalogProductByBarcodeForStoreRow, error) {
	return s.q.GetCatalogProductByBarcodeForStore(ctx, database.GetCatalogProductByBarcodeForStoreParams{
		OrganizationID: scope.OrganizationID,
		StoreID:        scope.StoreID,
		Barcode:        barcode,
	})
}
