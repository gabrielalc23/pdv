package checkout

import (
	"context"
	"fmt"

	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/jackc/pgx/v5/pgtype"
)

type TxQueries interface {
	LockSaleByID(context.Context, pgtype.UUID) (database.LockSaleByIDRow, error)
	ListSaleItemsBySaleID(context.Context, pgtype.UUID) ([]database.SaleItem, error)
	GetPaymentMethodByID(context.Context, pgtype.UUID) (database.PaymentMethod, error)
	GetInventoryByProductID(context.Context, pgtype.UUID) (database.Inventory, error)
	DecreaseInventory(context.Context, database.DecreaseInventoryParams) (database.DecreaseInventoryRow, error)
	CreateInventoryMovement(context.Context, database.CreateInventoryMovementParams) (database.InventoryMovement, error)
	CreatePayment(context.Context, database.CreatePaymentParams) (database.CreatePaymentRow, error)
	ApprovePayment(context.Context, database.ApprovePaymentParams) (database.ApprovePaymentRow, error)
	CompleteSale(context.Context, pgtype.UUID) (database.CompleteSaleRow, error)
	CreateFiscalDocument(context.Context, database.CreateFiscalDocumentParams) (database.CreateFiscalDocumentRow, error)
	MarkFiscalDocumentAuthorized(context.Context, database.MarkFiscalDocumentAuthorizedParams) (database.FiscalDocument, error)
	MarkFiscalDocumentError(context.Context, database.MarkFiscalDocumentErrorParams) (database.FiscalDocument, error)
}

type TxManager interface {
	WithTx(context.Context, func(TxQueries) error) error
}

type storeTxManager struct {
	store *database.Store
}

func NewTxManager(store *database.Store) TxManager {
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
