package payments

import (
	"context"

	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/gabrielalc23/pdv/internal/platform/tenancy"
	"github.com/jackc/pgx/v5/pgtype"
)

type Store interface {
	ListPaymentsBySaleID(ctx context.Context, scope tenancy.StoreScope, saleID pgtype.UUID) ([]database.Payment, error)
	GetSaleByID(ctx context.Context, scope tenancy.StoreScope, saleID pgtype.UUID) (database.Sale, error)
	ListActivePaymentMethods(ctx context.Context, scope tenancy.OrganizationScope) ([]database.PaymentMethod, error)
	GetPaymentMethodByID(ctx context.Context, scope tenancy.OrganizationScope, id pgtype.UUID) (database.PaymentMethod, error)
}

type storeImpl struct {
	q *database.Queries
}

func NewStore(q *database.Queries) Store {
	return &storeImpl{q: q}
}

func (s *storeImpl) ListPaymentsBySaleID(ctx context.Context, scope tenancy.StoreScope, saleID pgtype.UUID) ([]database.Payment, error) {
	return s.q.ListPaymentsBySaleIDForStore(ctx, database.ListPaymentsBySaleIDForStoreParams{
		OrganizationID: scope.OrganizationID,
		StoreID:        scope.StoreID,
		SaleID:         saleID,
	})
}

func (s *storeImpl) GetSaleByID(ctx context.Context, scope tenancy.StoreScope, saleID pgtype.UUID) (database.Sale, error) {
	return s.q.GetSaleByIDForStore(ctx, database.GetSaleByIDForStoreParams{
		OrganizationID: scope.OrganizationID,
		StoreID:        scope.StoreID,
		ID:             saleID,
	})
}

func (s *storeImpl) ListActivePaymentMethods(ctx context.Context, scope tenancy.OrganizationScope) ([]database.PaymentMethod, error) {
	return s.q.ListActivePaymentMethodsForOrganization(ctx, scope.OrganizationID)
}

func (s *storeImpl) GetPaymentMethodByID(ctx context.Context, scope tenancy.OrganizationScope, id pgtype.UUID) (database.PaymentMethod, error) {
	return s.q.GetPaymentMethodByIDForOrganization(ctx, database.GetPaymentMethodByIDForOrganizationParams{
		OrganizationID: scope.OrganizationID,
		ID:             id,
	})
}
