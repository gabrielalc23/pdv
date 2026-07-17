package inventory

import (
	"context"
	"fmt"

	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/jackc/pgx/v5/pgtype"
)

type ReadStore interface {
	GetProductByID(context.Context, pgtype.UUID) (database.Product, error)
	GetInventoryByProductID(context.Context, pgtype.UUID) (database.Inventory, error)
	ListInventory(context.Context, database.ListInventoryParams) ([]database.ListInventoryRow, error)
	CountInventory(context.Context, database.CountInventoryParams) (int64, error)
	ListInventoryMovementsByProductID(context.Context, database.ListInventoryMovementsByProductIDParams) ([]database.InventoryMovement, error)
	CountInventoryMovementsByProductID(context.Context, database.CountInventoryMovementsByProductIDParams) (int64, error)
}

type TxQueries interface {
	GetProductByID(context.Context, pgtype.UUID) (database.Product, error)
	GetInventoryByProductID(context.Context, pgtype.UUID) (database.Inventory, error)
	IncreaseInventory(context.Context, database.IncreaseInventoryParams) (database.IncreaseInventoryRow, error)
	DecreaseInventory(context.Context, database.DecreaseInventoryParams) (database.DecreaseInventoryRow, error)
	CreateInventoryMovement(context.Context, database.CreateInventoryMovementParams) (database.InventoryMovement, error)
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
