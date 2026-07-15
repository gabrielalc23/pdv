package inventory

import (
	"fmt"
	"strings"
	"time"

	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/jackc/pgx/v5/pgtype"
)

func toInventoryResponse(row database.ListInventoryRow) (InventoryResponse, error) {
	quantity, err := numericToQuantityString(row.Quantity)
	if err != nil {
		return InventoryResponse{}, fmt.Errorf("format quantity: %w", err)
	}

	createdAt := time.Time{}
	if row.CreatedAt.Valid {
		createdAt = row.CreatedAt.Time.UTC()
	}

	updatedAt := time.Time{}
	if row.UpdatedAt.Valid {
		updatedAt = row.UpdatedAt.Time.UTC()
	}

	var barcode *string
	if row.Barcode.Valid {
		value := row.Barcode.String
		barcode = &value
	}

	return InventoryResponse{
		ProductID: row.ProductID.String(),
		SKU:       row.SKU,
		Barcode:   barcode,
		Name:      row.Name,
		Quantity:  quantity,
		IsActive:  row.IsActive,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

func toInventoryChangeSummary(productID pgtype.UUID, previous, current pgtype.Numeric, updatedAt pgtype.Timestamptz) InventoryChangeSummary {
	previousQuantity := "0.000"
	if value, err := numericToQuantityString(previous); err == nil {
		previousQuantity = value
	}

	currentQuantity := "0.000"
	if value, err := numericToQuantityString(current); err == nil {
		currentQuantity = value
	}

	updated := time.Time{}
	if updatedAt.Valid {
		updated = updatedAt.Time.UTC()
	}

	return InventoryChangeSummary{
		ProductID:        productID.String(),
		PreviousQuantity: previousQuantity,
		CurrentQuantity:  currentQuantity,
		UpdatedAt:        updated,
	}
}

func toInventoryMovementResponse(movement database.InventoryMovement) (InventoryMovementResponse, error) {
	quantity := ""
	if value, err := numericToQuantityString(movement.Quantity); err == nil {
		quantity = value
	}

	previousQuantity := ""
	if value, err := numericToQuantityString(movement.PreviousQuantity); err == nil {
		previousQuantity = value
	}

	currentQuantity := ""
	if value, err := numericToQuantityString(movement.CurrentQuantity); err == nil {
		currentQuantity = value
	}

	var reason *string
	if movement.Reason.Valid {
		value := movement.Reason.String
		reason = &value
	}

	createdAt := time.Time{}
	if movement.CreatedAt.Valid {
		createdAt = movement.CreatedAt.Time.UTC()
	}

	return InventoryMovementResponse{
		ID:               movement.ID.String(),
		ProductID:        movement.ProductID.String(),
		Type:             string(movement.MovementType),
		Quantity:         quantity,
		PreviousQuantity: previousQuantity,
		CurrentQuantity:  currentQuantity,
		Reason:           reason,
		ReferenceType:    movement.ReferenceType,
		ReferenceID:      movement.ReferenceID.String(),
		CreatedAt:        createdAt,
	}, nil
}

func toInventoryDetailsResponse(product database.Product, inventory database.Inventory) (InventoryResponse, error) {
	quantity, err := numericToQuantityString(inventory.Quantity)
	if err != nil {
		return InventoryResponse{}, fmt.Errorf("format quantity: %w", err)
	}

	var barcode *string
	if product.Barcode.Valid {
		value := product.Barcode.String
		barcode = &value
	}

	createdAt := time.Time{}
	if inventory.CreatedAt.Valid {
		createdAt = inventory.CreatedAt.Time.UTC()
	}

	updatedAt := time.Time{}
	if inventory.UpdatedAt.Valid {
		updatedAt = inventory.UpdatedAt.Time.UTC()
	}

	return InventoryResponse{
		ProductID: product.ID.String(),
		SKU:       product.SKU,
		Barcode:   barcode,
		Name:      product.Name,
		Quantity:  quantity,
		IsActive:  product.IsActive,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

func numericToQuantityString(value pgtype.Numeric) (string, error) {
	if !value.Valid {
		return "", fmt.Errorf("numeric value is null")
	}

	raw, err := value.Value()
	if err != nil {
		return "", err
	}

	text, ok := raw.(string)
	if !ok {
		return "", fmt.Errorf("unexpected numeric driver value %T", raw)
	}

	return normalizeQuantityString("quantity", text)
}

func paginationResponse(page, pageSize int, total int64) PaginationResponse {
	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}

	return PaginationResponse{
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
	}
}

func optionalText(value string) pgtype.Text {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return pgtype.Text{}
	}

	return pgtype.Text{String: trimmed, Valid: true}
}
