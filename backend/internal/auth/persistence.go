package auth

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func mapPersistenceError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrSessionNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.ConstraintName {
		case "users_email_normalized_unique":
			return ErrEmailAlreadyInUse
		case "organizations_slug_unique":
			return ErrOrganizationSlugInUse
		case "stores_organization_id_code_unique":
			return ErrStoreCodeInUse
		}
	}
	return fmt.Errorf("%w: persistence operation: %w", ErrDependencyUnavailable, err)
}
