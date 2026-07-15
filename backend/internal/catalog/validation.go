package catalog

import (
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
)

func normalizePagination(page, pageSize *int) (int, int, error) {
	resolvedPage := 1
	if page != nil {
		if *page < 1 {
			return 0, 0, newValidationError("page", "must be greater than zero")
		}
		resolvedPage = *page
	}

	resolvedPageSize := 20
	if pageSize != nil {
		if *pageSize < 1 {
			return 0, 0, newValidationError("pageSize", "must be greater than zero")
		}
		if *pageSize > 100 {
			return 0, 0, newValidationError("pageSize", "must be at most 100")
		}
		resolvedPageSize = *pageSize
	}

	return resolvedPage, resolvedPageSize, nil
}

func parseUUID(raw, field string) (pgtype.UUID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return pgtype.UUID{}, newValidationError(field, "is required")
	}

	var id pgtype.UUID
	if err := id.Scan(raw); err != nil || !id.Valid {
		return pgtype.UUID{}, newValidationError(field, "must be a valid UUID")
	}

	return id, nil
}

func normalizeRequiredText(field, value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", newValidationError(field, "is required")
	}

	return trimmed, nil
}

func normalizeOptionalSearch(value string) pgtype.Text {
	return optionalText(value)
}
