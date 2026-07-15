package products

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func translatePersistenceError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return ErrProductNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		switch pgErr.ConstraintName {
		case "products_sku_unique":
			return ErrSKUAlreadyExists
		case "products_barcode_unique":
			return ErrBarcodeAlreadyExists
		}
	}

	return fmt.Errorf("database operation failed: %w", err)
}
