package fiscal

import (
	"context"

	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/jackc/pgx/v5/pgtype"
)

type Store interface {
	GetSaleByID(context.Context, pgtype.UUID) (database.GetSaleByIDRow, error)
	GetFiscalDocumentBySaleID(context.Context, pgtype.UUID) (database.FiscalDocument, error)
	MarkFiscalDocumentAuthorized(context.Context, database.MarkFiscalDocumentAuthorizedParams) (database.FiscalDocument, error)
	MarkFiscalDocumentError(context.Context, database.MarkFiscalDocumentErrorParams) (database.FiscalDocument, error)
}
