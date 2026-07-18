package products

import (
	"context"

	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/gabrielalc23/pdv/internal/platform/tenancy"
	"github.com/jackc/pgx/v5/pgtype"
)

type Store interface {
	CreateProduct(ctx context.Context, scope tenancy.OrganizationScope, params database.CreateProductForOrganizationParams) (database.CreateProductForOrganizationRow, error)
	GetProductByID(ctx context.Context, scope tenancy.OrganizationScope, id pgtype.UUID) (database.GetProductByIDForOrganizationRow, error)
	GetProductBySKU(ctx context.Context, scope tenancy.OrganizationScope, sku string) (database.GetProductBySKUForOrganizationRow, error)
	GetProductByBarcode(ctx context.Context, scope tenancy.OrganizationScope, barcode pgtype.Text) (database.GetProductByBarcodeForOrganizationRow, error)
	ListProducts(ctx context.Context, scope tenancy.OrganizationScope, params database.ListProductsForOrganizationParams) ([]database.ListProductsForOrganizationRow, error)
	CountProducts(ctx context.Context, scope tenancy.OrganizationScope, params database.CountProductsForOrganizationParams) (int64, error)
	UpdateProduct(ctx context.Context, scope tenancy.OrganizationScope, params database.UpdateProductForOrganizationParams) (database.UpdateProductForOrganizationRow, error)
	ActivateProduct(ctx context.Context, scope tenancy.OrganizationScope, id pgtype.UUID) (database.ActivateProductForOrganizationRow, error)
	DeactivateProduct(ctx context.Context, scope tenancy.OrganizationScope, id pgtype.UUID) (database.DeactivateProductForOrganizationRow, error)
}

type storeImpl struct {
	q *database.Queries
}

func NewStore(q *database.Queries) Store {
	return &storeImpl{q: q}
}

func (s *storeImpl) CreateProduct(ctx context.Context, scope tenancy.OrganizationScope, params database.CreateProductForOrganizationParams) (database.CreateProductForOrganizationRow, error) {
	params.OrganizationID = scope.OrganizationID
	return s.q.CreateProductForOrganization(ctx, params)
}

func (s *storeImpl) GetProductByID(ctx context.Context, scope tenancy.OrganizationScope, id pgtype.UUID) (database.GetProductByIDForOrganizationRow, error) {
	return s.q.GetProductByIDForOrganization(ctx, database.GetProductByIDForOrganizationParams{
		OrganizationID: scope.OrganizationID,
		ID:             id,
	})
}

func (s *storeImpl) GetProductBySKU(ctx context.Context, scope tenancy.OrganizationScope, sku string) (database.GetProductBySKUForOrganizationRow, error) {
	return s.q.GetProductBySKUForOrganization(ctx, database.GetProductBySKUForOrganizationParams{
		OrganizationID: scope.OrganizationID,
		SKU:            sku,
	})
}

func (s *storeImpl) GetProductByBarcode(ctx context.Context, scope tenancy.OrganizationScope, barcode pgtype.Text) (database.GetProductByBarcodeForOrganizationRow, error) {
	return s.q.GetProductByBarcodeForOrganization(ctx, database.GetProductByBarcodeForOrganizationParams{
		OrganizationID: scope.OrganizationID,
		Barcode:        barcode,
	})
}

func (s *storeImpl) ListProducts(ctx context.Context, scope tenancy.OrganizationScope, params database.ListProductsForOrganizationParams) ([]database.ListProductsForOrganizationRow, error) {
	params.OrganizationID = scope.OrganizationID
	return s.q.ListProductsForOrganization(ctx, params)
}

func (s *storeImpl) CountProducts(ctx context.Context, scope tenancy.OrganizationScope, params database.CountProductsForOrganizationParams) (int64, error) {
	params.OrganizationID = scope.OrganizationID
	return s.q.CountProductsForOrganization(ctx, params)
}

func (s *storeImpl) UpdateProduct(ctx context.Context, scope tenancy.OrganizationScope, params database.UpdateProductForOrganizationParams) (database.UpdateProductForOrganizationRow, error) {
	params.OrganizationID = scope.OrganizationID
	return s.q.UpdateProductForOrganization(ctx, params)
}

func (s *storeImpl) ActivateProduct(ctx context.Context, scope tenancy.OrganizationScope, id pgtype.UUID) (database.ActivateProductForOrganizationRow, error) {
	return s.q.ActivateProductForOrganization(ctx, database.ActivateProductForOrganizationParams{
		OrganizationID: scope.OrganizationID,
		ID:             id,
	})
}

func (s *storeImpl) DeactivateProduct(ctx context.Context, scope tenancy.OrganizationScope, id pgtype.UUID) (database.DeactivateProductForOrganizationRow, error) {
	return s.q.DeactivateProductForOrganization(ctx, database.DeactivateProductForOrganizationParams{
		OrganizationID: scope.OrganizationID,
		ID:             id,
	})
}
