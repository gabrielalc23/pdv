package products

import (
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
)

type normalizedProductInput struct {
	SKU        string
	Barcode    *string
	Name       string
	CategoryID pgtype.UUID
	Price      pgtype.Numeric
	Cost       pgtype.Numeric
}

func normalizeRequiredText(field, value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", newValidationError(field, "is required")
	}

	return trimmed, nil
}

func normalizeOptionalText(field string, value *string) (*string, error) {
	if value == nil {
		return nil, nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil, newValidationError(field, "cannot be blank")
	}

	clean := trimmed
	return &clean, nil
}

func normalizeUpsertInput(input UpsertProductInput) (normalizedProductInput, error) {
	sku, err := normalizeRequiredText("sku", input.SKU)
	if err != nil {
		return normalizedProductInput{}, err
	}

	name, err := normalizeRequiredText("name", input.Name)
	if err != nil {
		return normalizedProductInput{}, err
	}

	price, err := parseMoney("price", input.Price)
	if err != nil {
		return normalizedProductInput{}, err
	}

	cost, err := parseOptionalMoney("cost", input.Cost)
	if err != nil {
		return normalizedProductInput{}, err
	}

	barcode, err := normalizeOptionalText("barcode", input.Barcode)
	if err != nil {
		return normalizedProductInput{}, err
	}

	categoryID := pgtype.UUID{}
	if input.CategoryID != nil && strings.TrimSpace(*input.CategoryID) != "" {
		if err := categoryID.Scan(strings.TrimSpace(*input.CategoryID)); err != nil || !categoryID.Valid {
			return normalizedProductInput{}, newValidationError("categoryId", "must be a valid UUID")
		}
	}

	return normalizedProductInput{
		SKU:        sku,
		Barcode:    barcode,
		Name:       name,
		CategoryID: categoryID,
		Price:      price,
		Cost:       cost,
	}, nil
}

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

func parseMoney(field, value string) (pgtype.Numeric, error) {
	canonical, err := normalizeMoneyString(field, value)
	if err != nil {
		return pgtype.Numeric{}, err
	}

	var numeric pgtype.Numeric
	if err := numeric.ScanScientific(canonical); err != nil {
		return pgtype.Numeric{}, newValidationError(field, "must be a valid monetary amount")
	}

	return numeric, nil
}

func parseOptionalMoney(field string, value *string) (pgtype.Numeric, error) {
	if value == nil {
		return pgtype.Numeric{}, nil
	}

	return parseMoney(field, *value)
}

func normalizeMoneyString(field, value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", newValidationError(field, "is required")
	}

	if strings.HasPrefix(trimmed, "-") || strings.HasPrefix(trimmed, "+") {
		return "", newValidationError(field, "cannot be negative")
	}

	whole, fraction, hasFraction := strings.Cut(trimmed, ".")
	if hasFraction {
		if whole == "" || fraction == "" {
			return "", newValidationError(field, "must be a valid monetary amount")
		}
		if !allDigits(whole) || !allDigits(fraction) || len(fraction) > 2 {
			return "", newValidationError(field, "must have at most two decimal places")
		}
		if len(fraction) == 1 {
			fraction += "0"
		}

		return whole + "." + fraction, nil
	}

	if !allDigits(trimmed) {
		return "", newValidationError(field, "must be a valid monetary amount")
	}

	return trimmed + ".00", nil
}

func allDigits(value string) bool {
	if value == "" {
		return false
	}

	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}

	return true
}
