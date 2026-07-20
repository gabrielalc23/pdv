package sales

import (
	"context"
	"fmt"

	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/gabrielalc23/pdv/internal/platform/tenancy"
	"github.com/jackc/pgx/v5/pgtype"
)

type ReadStore interface {
	GetSaleByID(ctx context.Context, scope tenancy.StoreScope, id pgtype.UUID) (database.Sale, error)
	ListSales(ctx context.Context, scope tenancy.StoreScope, params database.ListSalesForStoreParams) ([]database.Sale, error)
	CountSales(ctx context.Context, scope tenancy.StoreScope, params database.CountSalesForStoreParams) (int64, error)
	ListSaleItemsBySaleID(ctx context.Context, scope tenancy.StoreScope, saleID pgtype.UUID) ([]database.SaleItem, error)
}

type TxQueries interface {
	GetSaleByID(ctx context.Context, scope tenancy.StoreScope, id pgtype.UUID) (database.Sale, error)
	LockSaleByID(ctx context.Context, scope tenancy.StoreScope, id pgtype.UUID) (database.Sale, error)
	GetSaleItemByID(ctx context.Context, scope tenancy.StoreScope, params database.GetSaleItemByIDForStoreParams) (database.SaleItem, error)
	CreateSaleForStore(ctx context.Context, scope tenancy.ActorScope, params database.CreateSaleForStoreParams) (database.CreateSaleForStoreRow, error)
	CreateSaleItemForStore(ctx context.Context, scope tenancy.StoreScope, params database.CreateSaleItemForStoreParams) (database.SaleItem, error)
	UpdateSaleItemForStore(ctx context.Context, scope tenancy.StoreScope, params database.UpdateSaleItemForStoreParams) (database.SaleItem, error)
	DeleteSaleItemForStore(ctx context.Context, scope tenancy.StoreScope, params database.DeleteSaleItemForStoreParams) (database.SaleItem, error)
	RecalculateSaleTotalsForStore(ctx context.Context, scope tenancy.StoreScope, params database.RecalculateSaleTotalsForStoreParams) (database.Sale, error)
	CancelSaleForStore(ctx context.Context, scope tenancy.ActorScope, params database.CancelSaleForStoreParams) (database.Sale, error)
	CompleteSaleForStore(ctx context.Context, scope tenancy.ActorScope, params database.CompleteSaleForStoreParams) (database.Sale, error)
	ListSaleItemsBySaleID(ctx context.Context, scope tenancy.StoreScope, saleID pgtype.UUID) ([]database.SaleItem, error)
	GetProductByID(ctx context.Context, scope tenancy.StoreScope, id pgtype.UUID) (database.GetProductByIDForStoreRow, error)
}

type TxManager interface {
	WithTx(ctx context.Context, scope tenancy.ActorScope, fn func(TxQueries) error) error
}

type storeTxManager struct {
	store *database.PostgresStore
}

func NewTxManager(store *database.PostgresStore) TxManager {
	return storeTxManager{store: store}
}

func (m storeTxManager) WithTx(ctx context.Context, scope tenancy.ActorScope, fn func(TxQueries) error) error {
	if m.store == nil {
		return fmt.Errorf("transaction store is nil")
	}
	return m.store.WithTx(ctx, func(tx *database.Tx) error {
		return fn(&txQueriesImpl{q: tx.Queries})
	})
}

type readStoreImpl struct {
	q *database.Queries
}

func NewReadStore(q *database.Queries) ReadStore {
	return &readStoreImpl{q: q}
}

func (s *readStoreImpl) GetSaleByID(ctx context.Context, scope tenancy.StoreScope, id pgtype.UUID) (database.Sale, error) {
	return s.q.GetSaleByIDForStore(ctx, database.GetSaleByIDForStoreParams{
		OrganizationID: scope.OrganizationID,
		StoreID:        scope.StoreID,
		ID:             id,
	})
}

func (s *readStoreImpl) ListSales(ctx context.Context, scope tenancy.StoreScope, params database.ListSalesForStoreParams) ([]database.Sale, error) {
	params.OrganizationID = scope.OrganizationID
	params.StoreID = scope.StoreID
	return s.q.ListSalesForStore(ctx, params)
}

func (s *readStoreImpl) CountSales(ctx context.Context, scope tenancy.StoreScope, params database.CountSalesForStoreParams) (int64, error) {
	params.OrganizationID = scope.OrganizationID
	params.StoreID = scope.StoreID
	return s.q.CountSalesForStore(ctx, params)
}

func (s *readStoreImpl) ListSaleItemsBySaleID(ctx context.Context, scope tenancy.StoreScope, saleID pgtype.UUID) ([]database.SaleItem, error) {
	return s.q.ListSaleItemsBySaleIDForStore(ctx, database.ListSaleItemsBySaleIDForStoreParams{
		OrganizationID: scope.OrganizationID,
		StoreID:        scope.StoreID,
		SaleID:         saleID,
	})
}

type txQueriesImpl struct {
	q *database.Queries
}

func (t *txQueriesImpl) GetSaleByID(ctx context.Context, scope tenancy.StoreScope, id pgtype.UUID) (database.Sale, error) {
	return t.q.GetSaleByIDForStore(ctx, database.GetSaleByIDForStoreParams{
		OrganizationID: scope.OrganizationID,
		StoreID:        scope.StoreID,
		ID:             id,
	})
}

func (t *txQueriesImpl) LockSaleByID(ctx context.Context, scope tenancy.StoreScope, id pgtype.UUID) (database.Sale, error) {
	return t.q.LockSaleByIDForStore(ctx, database.LockSaleByIDForStoreParams{
		OrganizationID: scope.OrganizationID,
		StoreID:        scope.StoreID,
		ID:             id,
	})
}

func (t *txQueriesImpl) GetSaleItemByID(ctx context.Context, scope tenancy.StoreScope, params database.GetSaleItemByIDForStoreParams) (database.SaleItem, error) {
	params.OrganizationID = scope.OrganizationID
	params.StoreID = scope.StoreID
	return t.q.GetSaleItemByIDForStore(ctx, params)
}

func (t *txQueriesImpl) CreateSaleForStore(ctx context.Context, scope tenancy.ActorScope, params database.CreateSaleForStoreParams) (database.CreateSaleForStoreRow, error) {
	params.OrganizationID = scope.OrganizationID
	params.StoreID = scope.StoreID
	params.OpenedByMembershipID = scope.ActorMembershipID
	return t.q.CreateSaleForStore(ctx, params)
}

func (t *txQueriesImpl) CreateSaleItemForStore(ctx context.Context, scope tenancy.StoreScope, params database.CreateSaleItemForStoreParams) (database.SaleItem, error) {
	params.OrganizationID = scope.OrganizationID
	params.StoreID = scope.StoreID
	return t.q.CreateSaleItemForStore(ctx, params)
}

func (t *txQueriesImpl) UpdateSaleItemForStore(ctx context.Context, scope tenancy.StoreScope, params database.UpdateSaleItemForStoreParams) (database.SaleItem, error) {
	params.OrganizationID = scope.OrganizationID
	params.StoreID = scope.StoreID
	return t.q.UpdateSaleItemForStore(ctx, params)
}

func (t *txQueriesImpl) DeleteSaleItemForStore(ctx context.Context, scope tenancy.StoreScope, params database.DeleteSaleItemForStoreParams) (database.SaleItem, error) {
	params.OrganizationID = scope.OrganizationID
	params.StoreID = scope.StoreID
	return t.q.DeleteSaleItemForStore(ctx, params)
}

func (t *txQueriesImpl) RecalculateSaleTotalsForStore(ctx context.Context, scope tenancy.StoreScope, params database.RecalculateSaleTotalsForStoreParams) (database.Sale, error) {
	params.OrganizationID = scope.OrganizationID
	params.StoreID = scope.StoreID
	return t.q.RecalculateSaleTotalsForStore(ctx, params)
}

func (t *txQueriesImpl) CancelSaleForStore(ctx context.Context, scope tenancy.ActorScope, params database.CancelSaleForStoreParams) (database.Sale, error) {
	params.OrganizationID = scope.OrganizationID
	params.StoreID = scope.StoreID
	params.CancelledByMembershipID = scope.ActorMembershipID
	return t.q.CancelSaleForStore(ctx, params)
}

func (t *txQueriesImpl) CompleteSaleForStore(ctx context.Context, scope tenancy.ActorScope, params database.CompleteSaleForStoreParams) (database.Sale, error) {
	params.OrganizationID = scope.OrganizationID
	params.StoreID = scope.StoreID
	params.CompletedByMembershipID = scope.ActorMembershipID
	return t.q.CompleteSaleForStore(ctx, params)
}

func (t *txQueriesImpl) ListSaleItemsBySaleID(ctx context.Context, scope tenancy.StoreScope, saleID pgtype.UUID) ([]database.SaleItem, error) {
	return t.q.ListSaleItemsBySaleIDForStore(ctx, database.ListSaleItemsBySaleIDForStoreParams{
		OrganizationID: scope.OrganizationID,
		StoreID:        scope.StoreID,
		SaleID:         saleID,
	})
}

func (t *txQueriesImpl) GetProductByID(ctx context.Context, scope tenancy.StoreScope, id pgtype.UUID) (database.GetProductByIDForStoreRow, error) {
	return t.q.GetProductByIDForStore(ctx, database.GetProductByIDForStoreParams{
		OrganizationID: scope.OrganizationID,
		StoreID:        scope.StoreID,
		ID:             id,
	})
}
