package checkout

import (
	"context"
	"fmt"

	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/gabrielalc23/pdv/internal/platform/tenancy"
	"github.com/jackc/pgx/v5/pgtype"
)

type TxQueries interface {
	LockSaleByID(ctx context.Context, scope tenancy.StoreScope, id pgtype.UUID) (database.Sale, error)
	ListSaleItemsBySaleID(ctx context.Context, scope tenancy.StoreScope, saleID pgtype.UUID) ([]database.SaleItem, error)
	GetPaymentMethodByID(ctx context.Context, scope tenancy.OrganizationScope, id pgtype.UUID) (database.PaymentMethod, error)
	GetInventoryByProductID(ctx context.Context, scope tenancy.StoreScope, productID pgtype.UUID) (database.Inventory, error)
	DecreaseInventory(ctx context.Context, scope tenancy.StoreScope, params database.DecreaseInventoryForStoreParams) (database.DecreaseInventoryForStoreRow, error)
	CreateInventoryMovement(ctx context.Context, scope tenancy.ActorScope, params database.CreateInventoryMovementForStoreParams) (database.CreateInventoryMovementForStoreRow, error)
	CreatePayment(ctx context.Context, scope tenancy.ActorScope, params database.CreatePaymentForStoreParams) (database.CreatePaymentForStoreRow, error)
	ApprovePayment(ctx context.Context, scope tenancy.StoreScope, params database.ApprovePaymentForStoreParams) (database.Payment, error)
	CompleteSale(ctx context.Context, scope tenancy.ActorScope, params database.CompleteSaleForStoreParams) (database.Sale, error)
	CreateFiscalDocument(ctx context.Context, scope tenancy.StoreScope, params database.CreateFiscalDocumentForStoreParams) (database.CreateFiscalDocumentForStoreRow, error)
	MarkFiscalDocumentAuthorized(ctx context.Context, scope tenancy.StoreScope, params database.MarkFiscalDocumentAuthorizedForStoreParams) (database.FiscalDocument, error)
	MarkFiscalDocumentError(ctx context.Context, scope tenancy.StoreScope, params database.MarkFiscalDocumentErrorForStoreParams) (database.FiscalDocument, error)
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

type txQueriesImpl struct {
	q *database.Queries
}

func (t *txQueriesImpl) LockSaleByID(ctx context.Context, scope tenancy.StoreScope, id pgtype.UUID) (database.Sale, error) {
	return t.q.LockSaleByIDForStore(ctx, database.LockSaleByIDForStoreParams{
		OrganizationID: scope.OrganizationID,
		StoreID:        scope.StoreID,
		ID:             id,
	})
}

func (t *txQueriesImpl) ListSaleItemsBySaleID(ctx context.Context, scope tenancy.StoreScope, saleID pgtype.UUID) ([]database.SaleItem, error) {
	return t.q.ListSaleItemsBySaleIDForStore(ctx, database.ListSaleItemsBySaleIDForStoreParams{
		OrganizationID: scope.OrganizationID,
		StoreID:        scope.StoreID,
		SaleID:         saleID,
	})
}

func (t *txQueriesImpl) GetPaymentMethodByID(ctx context.Context, scope tenancy.OrganizationScope, id pgtype.UUID) (database.PaymentMethod, error) {
	return t.q.GetPaymentMethodByIDForOrganization(ctx, database.GetPaymentMethodByIDForOrganizationParams{
		OrganizationID: scope.OrganizationID,
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

func (t *txQueriesImpl) CreatePayment(ctx context.Context, scope tenancy.ActorScope, params database.CreatePaymentForStoreParams) (database.CreatePaymentForStoreRow, error) {
	params.OrganizationID = scope.OrganizationID
	params.StoreID = scope.StoreID
	return t.q.CreatePaymentForStore(ctx, params)
}

func (t *txQueriesImpl) ApprovePayment(ctx context.Context, scope tenancy.StoreScope, params database.ApprovePaymentForStoreParams) (database.Payment, error) {
	params.OrganizationID = scope.OrganizationID
	params.StoreID = scope.StoreID
	return t.q.ApprovePaymentForStore(ctx, params)
}

func (t *txQueriesImpl) CompleteSale(ctx context.Context, scope tenancy.ActorScope, params database.CompleteSaleForStoreParams) (database.Sale, error) {
	params.OrganizationID = scope.OrganizationID
	params.StoreID = scope.StoreID
	params.CompletedByMembershipID = scope.ActorMembershipID
	return t.q.CompleteSaleForStore(ctx, params)
}

func (t *txQueriesImpl) CreateFiscalDocument(ctx context.Context, scope tenancy.StoreScope, params database.CreateFiscalDocumentForStoreParams) (database.CreateFiscalDocumentForStoreRow, error) {
	params.OrganizationID = scope.OrganizationID
	params.StoreID = scope.StoreID
	return t.q.CreateFiscalDocumentForStore(ctx, params)
}

func (t *txQueriesImpl) MarkFiscalDocumentAuthorized(ctx context.Context, scope tenancy.StoreScope, params database.MarkFiscalDocumentAuthorizedForStoreParams) (database.FiscalDocument, error) {
	params.OrganizationID = scope.OrganizationID
	params.StoreID = scope.StoreID
	return t.q.MarkFiscalDocumentAuthorizedForStore(ctx, params)
}

func (t *txQueriesImpl) MarkFiscalDocumentError(ctx context.Context, scope tenancy.StoreScope, params database.MarkFiscalDocumentErrorForStoreParams) (database.FiscalDocument, error) {
	params.OrganizationID = scope.OrganizationID
	params.StoreID = scope.StoreID
	return t.q.MarkFiscalDocumentErrorForStore(ctx, params)
}
