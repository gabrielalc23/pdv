package payments

import (
	"context"

	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/jackc/pgx/v5/pgtype"
)

type Store interface {
	ListActivePaymentMethods(context.Context) ([]database.PaymentMethod, error)
	ListPaymentsBySaleID(context.Context, pgtype.UUID) ([]database.ListPaymentsBySaleIDRow, error)
	GetSaleByID(context.Context, pgtype.UUID) (database.GetSaleByIDRow, error)
	GetPaymentMethodByID(context.Context, pgtype.UUID) (database.PaymentMethod, error)
}
