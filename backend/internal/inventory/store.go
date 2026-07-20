package inventory

import (
	"context"
	"fmt"

	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/gabrielalc23/pdv/internal/platform/tenancy"
	"github.com/jackc/pgx/v5/pgtype"
)

type ReadStore interface {
	GetProductByID(ctx context.Context, scope tenancy.StoreScope, id pgtype.UUID) (database.GetProductByIDForStoreRow, error)
	GetInventoryByProductID(ctx context.Context, scope tenancy.StoreScope, productID pgtype.UUID) (database.Inventory, error)
	ListInventory(ctx context.Context, scope tenancy.StoreScope, params database.ListInventoryForStoreParams) ([]database.ListInventoryForStoreRow, error)
	CountInventory(ctx context.Context, scope tenancy.StoreScope, params database.CountInventoryForStoreParams) (int64, error)
	ListInventoryMovementsByProductID(ctx context.Context, scope tenancy.StoreScope, params database.ListInventoryMovementsByProductIDForStoreParams) ([]database.ListInventoryMovementsByProductIDForStoreRow, error)
	CountInventoryMovementsByProductID(ctx context.Context, scope tenancy.StoreScope, params database.CountInventoryMovementsByProductIDForStoreParams) (int64, error)
}

type TxQueries interface {
	GetProductByID(ctx context.Context, scope tenancy.StoreScope, id pgtype.UUID) (database.GetProductByIDForStoreRow, error)
	GetInventoryByProductID(ctx context.Context, scope tenancy.StoreScope, productID pgtype.UUID) (database.Inventory, error)
	IncreaseInventory(ctx context.Context, scope tenancy.StoreScope, params database.IncreaseInventoryForStoreParams) (database.IncreaseInventoryForStoreRow, error)
	DecreaseInventory(ctx context.Context, scope tenancy.StoreScope, params database.DecreaseInventoryForStoreParams) (database.DecreaseInventoryForStoreRow, error)
	CreateInventoryMovement(ctx context.Context, scope tenancy.ActorScope, params database.CreateInventoryMovementForStoreParams) (database.CreateInventoryMovementForStoreRow, error)
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

func (s *readStoreImpl) GetProductByID(ctx context.Context, scope tenancy.StoreScope, id pgtype.UUID) (database.GetProductByIDForStoreRow, error) {
	return s.q.GetProductByIDForStore(ctx, database.GetProductByIDForStoreParams{
		OrganizationID: scope.OrganizationID,
		StoreID:        scope.StoreID,
		ID:             id,
	})
}

func (s *readStoreImpl) GetInventoryByProductID(ctx context.Context, scope tenancy.StoreScope, productID pgtype.UUID) (database.Inventory, error) {
	return s.q.GetInventoryByProductIDForStore(ctx, database.GetInventoryByProductIDForStoreParams{
		OrganizationID: scope.OrganizationID,
		StoreID:        scope.StoreID,
		ProductID:      productID,
	})
}

func (s *readStoreImpl) ListInventory(ctx context.Context, scope tenancy.StoreScope, params database.ListInventoryForStoreParams) ([]database.ListInventoryForStoreRow, error) {
	params.OrganizationID = scope.OrganizationID
	params.StoreID = scope.StoreID
	return s.q.ListInventoryForStore(ctx, params)
}

func (s *readStoreImpl) CountInventory(ctx context.Context, scope tenancy.StoreScope, params database.CountInventoryForStoreParams) (int64, error) {
	params.OrganizationID = scope.OrganizationID
	params.StoreID = scope.StoreID
	return s.q.CountInventoryForStore(ctx, params)
}

func (s *readStoreImpl) ListInventoryMovementsByProductID(ctx context.Context, scope tenancy.StoreScope, params database.ListInventoryMovementsByProductIDForStoreParams) ([]database.ListInventoryMovementsByProductIDForStoreRow, error) {
	params.OrganizationID = scope.OrganizationID
	params.StoreID = scope.StoreID
	return s.q.ListInventoryMovementsByProductIDForStore(ctx, params)
}

func (s *readStoreImpl) CountInventoryMovementsByProductID(ctx context.Context, scope tenancy.StoreScope, params database.CountInventoryMovementsByProductIDForStoreParams) (int64, error) {
	params.OrganizationID = scope.OrganizationID
	params.StoreID = scope.StoreID
	return s.q.CountInventoryMovementsByProductIDForStore(ctx, params)
}

type txQueriesImpl struct {
	q *database.Queries
}

func (t *txQueriesImpl) GetProductByID(ctx context.Context, scope tenancy.StoreScope, id pgtype.UUID) (database.GetProductByIDForStoreRow, error) {
	return t.q.GetProductByIDForStore(ctx, database.GetProductByIDForStoreParams{
		OrganizationID: scope.OrganizationID,
		StoreID:        scope.StoreID,
		ID:             id,
	})
}

func (t *txQueriesImpl) GetInventoryByProductID(ctx context.Context, scope tenancy.StoreScope, productID pgtype.UUID) (database.Inventory, error) {
	return t.q.GetInventoryByProductIDForStore(ctx, database.GetInventoryByProductIDForStoreParams{
		OrganizationID: scope.OrganizationID,
		StoreID:        scope.StoreID,
		ProductID:      productID,
	})
}

func (t *txQueriesImpl) IncreaseInventory(ctx context.Context, scope tenancy.StoreScope, params database.IncreaseInventoryForStoreParams) (database.IncreaseInventoryForStoreRow, error) {
	params.OrganizationID = scope.OrganizationID
	params.StoreID = scope.StoreID
	return t.q.IncreaseInventoryForStore(ctx, params)
}

func (t *txQueriesImpl) DecreaseInventory(ctx context.Context, scope tenancy.StoreScope, params database.DecreaseInventoryForStoreParams) (database.DecreaseInventoryForStoreRow, error) {
	params.OrganizationID = scope.OrganizationID
	params.StoreID = scope.StoreID
	return t.q.DecreaseInventoryForStore(ctx, params)
}

func (t *txQueriesImpl) CreateInventoryMovement(ctx context.Context, scope tenancy.ActorScope, params database.CreateInventoryMovementForStoreParams) (database.CreateInventoryMovementForStoreRow, error) {
	params.OrganizationID = scope.OrganizationID
	params.StoreID = scope.StoreID
	params.ActorMembershipID = scope.ActorMembershipID
	return t.q.CreateInventoryMovementForStore(ctx, params)
}
