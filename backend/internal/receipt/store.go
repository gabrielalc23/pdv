package receipt

import (
	"context"

	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/jackc/pgx/v5/pgtype"
)

type Store interface {
	GetSaleByID(context.Context, pgtype.UUID) (database.GetSaleByIDRow, error)
	ListSaleItemsBySaleID(context.Context, pgtype.UUID) ([]database.SaleItem, error)
	ListPaymentsBySaleID(context.Context, pgtype.UUID) ([]database.ListPaymentsBySaleIDRow, error)
	GetPaymentMethodByID(context.Context, pgtype.UUID) (database.PaymentMethod, error)
	GetFiscalDocumentBySaleID(context.Context, pgtype.UUID) (database.FiscalDocument, error)
}
