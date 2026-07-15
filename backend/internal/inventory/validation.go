package inventory

import (
	"strings"

	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/jackc/pgx/v5/pgtype"
)

type normalizedEntryInput struct {
	ProductID     pgtype.UUID
	Quantity      pgtype.Numeric
	Reason        pgtype.Text
	ReferenceType string
	ReferenceID   pgtype.UUID
}

type normalizedAdjustmentInput struct {
	ProductID     pgtype.UUID
	Direction     string
	Quantity      pgtype.Numeric
	Reason        pgtype.Text
	ReferenceType string
	ReferenceID   pgtype.UUID
}

func normalizeEntryInput(input CreateInventoryEntryInput) (normalizedEntryInput, error) {
	productID, err := parseUUID(input.ProductID, "productId")
	if err != nil {
		return normalizedEntryInput{}, err
	}

	quantity, err := parseQuantity("quantity", input.Quantity)
	if err != nil {
		return normalizedEntryInput{}, err
	}

	reason, err := normalizeOptionalReason(input.Reason)
	if err != nil {
		return normalizedEntryInput{}, err
	}

	referenceType, err := normalizeRequiredText("referenceType", input.ReferenceType)
	if err != nil {
		return normalizedEntryInput{}, err
	}

	referenceID, err := parseUUID(input.ReferenceID, "referenceId")
	if err != nil {
		return normalizedEntryInput{}, err
	}

	return normalizedEntryInput{
		ProductID:     productID,
		Quantity:      quantity,
		Reason:        reason,
		ReferenceType: referenceType,
		ReferenceID:   referenceID,
	}, nil
}

func normalizeAdjustmentInput(input CreateInventoryAdjustmentInput) (normalizedAdjustmentInput, error) {
	productID, err := parseUUID(input.ProductID, "productId")
	if err != nil {
		return normalizedAdjustmentInput{}, err
	}

	quantity, err := parseQuantity("quantity", input.Quantity)
	if err != nil {
		return normalizedAdjustmentInput{}, err
	}

	if input.Direction == "" {
		return normalizedAdjustmentInput{}, newValidationError("direction", "is required")
	}

	direction := strings.ToUpper(strings.TrimSpace(input.Direction))
	if direction != "IN" && direction != "OUT" {
		return normalizedAdjustmentInput{}, newValidationError("direction", "must be IN or OUT")
	}

	reason, err := normalizeRequiredText("reason", input.Reason)
	if err != nil {
		return normalizedAdjustmentInput{}, err
	}

	referenceType, err := normalizeRequiredText("referenceType", input.ReferenceType)
	if err != nil {
		return normalizedAdjustmentInput{}, err
	}

	referenceID, err := parseUUID(input.ReferenceID, "referenceId")
	if err != nil {
		return normalizedAdjustmentInput{}, err
	}

	return normalizedAdjustmentInput{
		ProductID:     productID,
		Direction:     direction,
		Quantity:      quantity,
		Reason:        pgtype.Text{String: reason, Valid: true},
		ReferenceType: referenceType,
		ReferenceID:   referenceID,
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

func normalizeRequiredText(field, value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", newValidationError(field, "is required")
	}

	return trimmed, nil
}

func normalizeOptionalReason(value *string) (pgtype.Text, error) {
	if value == nil {
		return pgtype.Text{}, nil
	}

	trimmed, err := normalizeRequiredText("reason", *value)
	if err != nil {
		return pgtype.Text{}, err
	}

	return pgtype.Text{String: trimmed, Valid: true}, nil
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

func hasPositiveDigit(value string) bool {
	for _, r := range value {
		if r >= '1' && r <= '9' {
			return true
		}
	}

	return false
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

func parseMovementTypeFilter(value string) (database.NullInventoryMovementType, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return database.NullInventoryMovementType{}, nil
	}

	movementType := database.InventoryMovementType(trimmed)
	if !movementType.Valid() {
		return database.NullInventoryMovementType{}, newValidationError("type", "must be a valid inventory movement type")
	}

	return database.NullInventoryMovementType{
		InventoryMovementType: movementType,
		Valid:                 true,
	}, nil
}
