package sales

import (
	"context"
	"fmt"

	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/jackc/pgx/v5/pgtype"
)

type ReadStore interface {
	GetSaleByID(context.Context, pgtype.UUID) (database.GetSaleByIDRow, error)
	ListSales(context.Context, database.ListSalesParams) ([]database.ListSalesRow, error)
	CountSales(context.Context, database.NullSaleStatus) (int64, error)
	ListSaleItemsBySaleID(context.Context, pgtype.UUID) ([]database.SaleItem, error)
}

type TxQueries interface {
	GetSaleByID(context.Context, pgtype.UUID) (database.GetSaleByIDRow, error)
	LockSaleByID(context.Context, pgtype.UUID) (database.LockSaleByIDRow, error)
	GetProductByID(context.Context, pgtype.UUID) (database.Product, error)
	GetSaleItemByID(context.Context, database.GetSaleItemByIDParams) (database.SaleItem, error)
	CreateSale(context.Context, string) (database.CreateSaleRow, error)
	CreateSaleItem(context.Context, database.CreateSaleItemParams) (database.SaleItem, error)
	UpdateSaleItem(context.Context, database.UpdateSaleItemParams) (database.SaleItem, error)
	DeleteSaleItem(context.Context, database.DeleteSaleItemParams) (database.SaleItem, error)
	RecalculateSaleTotals(context.Context, database.RecalculateSaleTotalsParams) (database.RecalculateSaleTotalsRow, error)
	CancelSale(context.Context, pgtype.UUID) (database.CancelSaleRow, error)
	ListSaleItemsBySaleID(context.Context, pgtype.UUID) ([]database.SaleItem, error)
}

type TxManager interface {
	WithTx(context.Context, func(TxQueries) error) error
}

type storeTxManager struct {
	store *database.PostgresStore
}

func NewTxManager(store *database.PostgresStore) TxManager {
	return storeTxManager{store: store}
}

func (m storeTxManager) WithTx(ctx context.Context, fn func(TxQueries) error) error {
	if m.store == nil {
		return fmt.Errorf("transaction store is nil")
	}

	return m.store.WithTx(ctx, func(tx *database.Tx) error {
		return fn(tx.Queries)
	})
}
