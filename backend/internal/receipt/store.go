package receipt

import (
	"context"

	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/gabrielalc23/pdv/internal/platform/tenancy"
	"github.com/jackc/pgx/v5/pgtype"
)

type Store interface {
	GetSaleByID(ctx context.Context, scope tenancy.StoreScope, saleID pgtype.UUID) (database.Sale, error)
	ListSaleItemsBySaleID(ctx context.Context, scope tenancy.StoreScope, saleID pgtype.UUID) ([]database.SaleItem, error)
	ListReceiptPaymentsBySaleID(ctx context.Context, scope tenancy.StoreScope, saleID pgtype.UUID) ([]database.ReceiptPayment, error)
	GetFiscalDocumentBySaleID(ctx context.Context, scope tenancy.StoreScope, saleID pgtype.UUID) (database.FiscalDocument, error)
}

type storeImpl struct {
	q *database.Queries
}

func NewStore(q *database.Queries) Store {
	return &storeImpl{q: q}
}

func (s *storeImpl) GetSaleByID(ctx context.Context, scope tenancy.StoreScope, saleID pgtype.UUID) (database.Sale, error) {
	return s.q.GetSaleByIDForStore(ctx, database.GetSaleByIDForStoreParams{
		OrganizationID: scope.OrganizationID,
		StoreID:        scope.StoreID,
		ID:             saleID,
	})
}

func (s *storeImpl) ListSaleItemsBySaleID(ctx context.Context, scope tenancy.StoreScope, saleID pgtype.UUID) ([]database.SaleItem, error) {
	return s.q.ListSaleItemsBySaleIDForStore(ctx, database.ListSaleItemsBySaleIDForStoreParams{
		OrganizationID: scope.OrganizationID,
		StoreID:        scope.StoreID,
		SaleID:         saleID,
	})
}

func (s *storeImpl) ListReceiptPaymentsBySaleID(ctx context.Context, scope tenancy.StoreScope, saleID pgtype.UUID) ([]database.ReceiptPayment, error) {
	return s.q.ListReceiptPaymentsBySaleIDForStore(ctx, database.ListReceiptPaymentsBySaleIDForStoreParams{
		OrganizationID: scope.OrganizationID,
		StoreID:        scope.StoreID,
		SaleID:         saleID,
	})
}

func (s *storeImpl) GetFiscalDocumentBySaleID(ctx context.Context, scope tenancy.StoreScope, saleID pgtype.UUID) (database.FiscalDocument, error) {
	return s.q.GetFiscalDocumentBySaleIDForStore(ctx, database.GetFiscalDocumentBySaleIDForStoreParams{
		OrganizationID: scope.OrganizationID,
		StoreID:        scope.StoreID,
		SaleID:         saleID,
	})
}
