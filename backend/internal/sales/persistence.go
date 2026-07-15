package sales

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func translateSaleReadError(err error) error {
	return translateNotFoundError(err, ErrSaleNotFound)
}

func translateSaleItemReadError(err error) error {
	return translateNotFoundError(err, ErrSaleItemNotFound)
}

func translateSaleItemMutationError(err error) error {
	return translateNotFoundError(err, ErrSaleItemNotFound)
}

func translateProductReadError(err error) error {
	return translateNotFoundError(err, ErrProductNotFound)
}

func translateNotFoundError(err error, notFound error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return notFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23503" {
		switch pgErr.ConstraintName {
		case "sale_items_sale_fk":
			return ErrSaleNotFound
		case "sale_items_product_fk":
			return ErrProductNotFound
		}
	}

	return fmt.Errorf("database operation failed: %w", err)
}

func translateSaleMutationError(err error) error {
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23503":
			switch pgErr.ConstraintName {
			case "sale_items_sale_fk":
				return ErrSaleNotFound
			case "sale_items_product_fk":
				return ErrProductNotFound
			}
		case "23514":
			switch pgErr.ConstraintName {
			case "sales_status_timestamps_consistency", "sales_total_consistency":
				return ErrSaleNotOpen
			case "sale_items_discount_not_greater_than_gross":
				return ErrSaleNotOpen
			}
		}
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return ErrSaleNotFound
	}

	return fmt.Errorf("database operation failed: %w", err)
}
