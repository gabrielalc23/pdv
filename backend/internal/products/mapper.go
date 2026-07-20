package products

import (
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type productProjection struct {
	ID         pgtype.UUID
	SKU        string
	Barcode    pgtype.Text
	Name       string
	CategoryID pgtype.UUID
	Price      pgtype.Numeric
	Cost       pgtype.Numeric
	IsActive   bool
	CreatedAt  pgtype.Timestamptz
	UpdatedAt  pgtype.Timestamptz
}

func productFromRow(id pgtype.UUID, sku string, barcode pgtype.Text, name string, categoryID pgtype.UUID, price, cost pgtype.Numeric, isActive bool, createdAt, updatedAt pgtype.Timestamptz) productProjection {
	return productProjection{
		ID: id, SKU: sku, Barcode: barcode, Name: name,
		CategoryID: categoryID, Price: price, Cost: cost,
		IsActive: isActive, CreatedAt: createdAt, UpdatedAt: updatedAt,
	}
}

func toProductResponse(product productProjection) (ProductResponse, error) {
	price, err := numericToMoneyString(product.Price)
	if err != nil {
		return ProductResponse{}, fmt.Errorf("format price: %w", err)
	}

	var barcode *string
	if product.Barcode.Valid {
		barcode = &product.Barcode.String
	}

	var cost *string
	if product.Cost.Valid {
		value, err := numericToMoneyString(product.Cost)
		if err != nil {
			return ProductResponse{}, fmt.Errorf("format cost: %w", err)
		}
		cost = &value
	}

	var categoryID *string
	if product.CategoryID.Valid {
		value := product.CategoryID.String()
		categoryID = &value
	}

	return ProductResponse{
		ID:         product.ID.String(),
		SKU:        product.SKU,
		Barcode:    barcode,
		Name:       product.Name,
		CategoryID: categoryID,
		Price:      price,
		Cost:       cost,
		IsActive:   product.IsActive,
		CreatedAt:  timestampOrZero(product.CreatedAt),
		UpdatedAt:  timestampOrZero(product.UpdatedAt),
	}, nil
}

func optionalText(value string) pgtype.Text {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: trimmed, Valid: true}
}

func toText(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *value, Valid: true}
}

func numericToMoneyString(value pgtype.Numeric) (string, error) {
	if !value.Valid {
		return "", fmt.Errorf("numeric value is null")
	}
	raw, err := value.Value()
	if err != nil {
		return "", fmt.Errorf("read numeric value: %w", err)
	}
	text, ok := raw.(string)
	if !ok {
		return "", fmt.Errorf("unexpected numeric driver value %T", raw)
	}
	return formatMoneyString(text)
}

func formatMoneyString(value string) (string, error) {
	whole, fraction, hasFraction := strings.Cut(value, ".")
	if !hasFraction {
		return value + ".00", nil
	}
	fraction = strings.TrimRight(fraction, "0")
	switch len(fraction) {
	case 0:
		return whole + ".00", nil
	case 1:
		return whole + "." + fraction + "0", nil
	case 2:
		return whole + "." + fraction, nil
	default:
		return "", fmt.Errorf("money value has more than two decimal places")
	}
}

func timestampOrZero(value pgtype.Timestamptz) time.Time {
	if !value.Valid {
		return time.Time{}
	}
	return value.Time.UTC()
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
