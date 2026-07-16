package receipt

import (
	"context"

	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/jackc/pgx/v5/pgtype"
)

type Store interface {
	GetSaleByID(context.Context, pgtype.UUID) (database.GetSaleByIDRow, error)
	ListSaleItemsBySaleID(context.Context, pgtype.UUID) ([]database.SaleItem, error)
	ListReceiptPaymentsBySaleID(context.Context, pgtype.UUID) ([]database.MvReceiptPayment, error)
	GetFiscalDocumentBySaleID(context.Context, pgtype.UUID) (database.FiscalDocument, error)
}
