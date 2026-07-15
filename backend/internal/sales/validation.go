package sales

import (
	"math/big"
	"strings"

	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/jackc/pgx/v5/pgtype"
)

type normalizedCreateSaleInput struct {
	IdempotencyKey string
}

type normalizedAddSaleItemInput struct {
	ProductID pgtype.UUID
	Quantity  pgtype.Numeric
	Discount  pgtype.Numeric
}

type normalizedUpdateSaleItemInput struct {
	Quantity pgtype.Numeric
	Discount pgtype.Numeric
}

func normalizeCreateSaleInput(input CreateSaleInput) (normalizedCreateSaleInput, error) {
	idempotencyKey, err := normalizeRequiredText("idempotencyKey", input.IdempotencyKey)
	if err != nil {
		return normalizedCreateSaleInput{}, err
	}

	return normalizedCreateSaleInput{IdempotencyKey: idempotencyKey}, nil
}

func normalizeAddSaleItemInput(input AddSaleItemInput) (normalizedAddSaleItemInput, error) {
	productID, err := parseUUID(input.ProductID, "productId")
	if err != nil {
		return normalizedAddSaleItemInput{}, err
	}

	quantity, err := parseQuantity("quantity", input.Quantity)
	if err != nil {
		return normalizedAddSaleItemInput{}, err
	}

	discount, err := parseOptionalMoney("discount", input.Discount)
	if err != nil {
		return normalizedAddSaleItemInput{}, err
	}

	return normalizedAddSaleItemInput{
		ProductID: productID,
		Quantity:  quantity,
		Discount:  discount,
	}, nil
}

func normalizeUpdateSaleItemInput(input UpdateSaleItemInput) (normalizedUpdateSaleItemInput, error) {
	quantity, err := parseQuantity("quantity", input.Quantity)
	if err != nil {
		return normalizedUpdateSaleItemInput{}, err
	}

	discount, err := parseOptionalMoney("discount", input.Discount)
	if err != nil {
		return normalizedUpdateSaleItemInput{}, err
	}

	return normalizedUpdateSaleItemInput{
		Quantity: quantity,
		Discount: discount,
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

func parseQuantity(field, value string) (pgtype.Numeric, error) {
	canonical, err := normalizeQuantityString(field, value)
	if err != nil {
		return pgtype.Numeric{}, err
	}

	if !hasPositiveDigit(canonical) {
		return pgtype.Numeric{}, newValidationError(field, "must be greater than zero")
	}

	var numeric pgtype.Numeric
	if err := numeric.ScanScientific(canonical); err != nil {
		return pgtype.Numeric{}, newValidationError(field, "must be a valid decimal number")
	}

	return numeric, nil
}

func normalizeQuantityString(field, value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", newValidationError(field, "is required")
	}

	if strings.HasPrefix(trimmed, "-") || strings.HasPrefix(trimmed, "+") {
		return "", newValidationError(field, "must be greater than zero")
	}

	whole, fraction, hasFraction := strings.Cut(trimmed, ".")
	if hasFraction {
		if whole == "" || fraction == "" {
			return "", newValidationError(field, "must be a valid decimal number")
		}
		if !allDigits(whole) || !allDigits(fraction) {
			return "", newValidationError(field, "must have at most three decimal places")
		}

		if len(fraction) > 3 {
			trimmedFraction := strings.TrimRight(fraction, "0")
			if trimmedFraction == "" {
				trimmedFraction = "0"
			}
			if len(trimmedFraction) > 3 {
				return "", newValidationError(field, "must have at most three decimal places")
			}
			fraction = trimmedFraction
		}

		for len(fraction) < 3 {
			fraction += "0"
		}

		return whole + "." + fraction, nil
	}

	if !allDigits(trimmed) {
		return "", newValidationError(field, "must be a valid decimal number")
	}

	return trimmed + ".000", nil
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
		return zeroMoney(), nil
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

func parseSaleStatusFilter(value string) (database.NullSaleStatus, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return database.NullSaleStatus{}, nil
	}

	status := database.SaleStatus(strings.ToUpper(trimmed))
	if !status.Valid() {
		return database.NullSaleStatus{}, newValidationError("status", "must be a valid sale status")
	}

	return database.NullSaleStatus{
		SaleStatus: status,
		Valid:      true,
	}, nil
}

func normalizeRequiredText(field, value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", newValidationError(field, "is required")
	}

	return trimmed, nil
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

func hasPositiveDigit(value string) bool {
	for _, r := range value {
		if r >= '1' && r <= '9' {
			return true
		}
	}

	return false
}

func zeroMoney() pgtype.Numeric {
	return numericFromScaledInt(big.NewInt(0), 2)
}
