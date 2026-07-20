package fiscal

import (
	"context"

	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/gabrielalc23/pdv/internal/platform/tenancy"
	"github.com/jackc/pgx/v5/pgtype"
)

type Store interface {
	GetSaleByID(ctx context.Context, scope tenancy.StoreScope, saleID pgtype.UUID) (database.Sale, error)
	GetFiscalDocumentBySaleID(ctx context.Context, scope tenancy.StoreScope, saleID pgtype.UUID) (database.FiscalDocument, error)
	GetFiscalDocumentByID(ctx context.Context, scope tenancy.StoreScope, id pgtype.UUID) (database.FiscalDocument, error)
	MarkFiscalDocumentAuthorized(ctx context.Context, scope tenancy.StoreScope, params database.MarkFiscalDocumentAuthorizedForStoreParams) (database.FiscalDocument, error)
	MarkFiscalDocumentError(ctx context.Context, scope tenancy.StoreScope, params database.MarkFiscalDocumentErrorForStoreParams) (database.FiscalDocument, error)
	LockFiscalDocumentBySaleID(ctx context.Context, scope tenancy.StoreScope, saleID pgtype.UUID) (database.FiscalDocument, error)
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

func (s *storeImpl) GetFiscalDocumentBySaleID(ctx context.Context, scope tenancy.StoreScope, saleID pgtype.UUID) (database.FiscalDocument, error) {
	return s.q.GetFiscalDocumentBySaleIDForStore(ctx, database.GetFiscalDocumentBySaleIDForStoreParams{
		OrganizationID: scope.OrganizationID,
		StoreID:        scope.StoreID,
		SaleID:         saleID,
	})
}

func (s *storeImpl) GetFiscalDocumentByID(ctx context.Context, scope tenancy.StoreScope, id pgtype.UUID) (database.FiscalDocument, error) {
	return s.q.GetFiscalDocumentByIDForStore(ctx, database.GetFiscalDocumentByIDForStoreParams{
		OrganizationID: scope.OrganizationID,
		StoreID:        scope.StoreID,
		ID:             id,
	})
}

func (s *storeImpl) MarkFiscalDocumentAuthorized(ctx context.Context, scope tenancy.StoreScope, params database.MarkFiscalDocumentAuthorizedForStoreParams) (database.FiscalDocument, error) {
	params.OrganizationID = scope.OrganizationID
	params.StoreID = scope.StoreID
	return s.q.MarkFiscalDocumentAuthorizedForStore(ctx, params)
}

func (s *storeImpl) MarkFiscalDocumentError(ctx context.Context, scope tenancy.StoreScope, params database.MarkFiscalDocumentErrorForStoreParams) (database.FiscalDocument, error) {
	params.OrganizationID = scope.OrganizationID
	params.StoreID = scope.StoreID
	return s.q.MarkFiscalDocumentErrorForStore(ctx, params)
}

func (s *storeImpl) LockFiscalDocumentBySaleID(ctx context.Context, scope tenancy.StoreScope, saleID pgtype.UUID) (database.FiscalDocument, error) {
	return s.q.LockFiscalDocumentBySaleIDForStore(ctx, database.LockFiscalDocumentBySaleIDForStoreParams{
		OrganizationID: scope.OrganizationID,
		StoreID:        scope.StoreID,
		SaleID:         saleID,
	})
}
