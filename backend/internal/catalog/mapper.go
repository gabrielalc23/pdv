package catalog

import (
	"fmt"
	"strings"
	"time"

	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/jackc/pgx/v5/pgtype"
)

type CatalogProductData struct {
	ID        pgtype.UUID
	SKU       string
	Barcode   pgtype.Text
	Name      string
	Price     pgtype.Numeric
	Quantity  pgtype.Numeric
	IsActive  bool
	InStock   bool
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
}

func ToCatalogProductResponse(data CatalogProductData) (CatalogProductResponse, error) {
	price, err := NumericToMoneyString(data.Price)
	if err != nil {
		return CatalogProductResponse{}, fmt.Errorf("format price: %w", err)
	}

	quantity, err := NumericToQuantityString(data.Quantity)
	if err != nil {
		return CatalogProductResponse{}, fmt.Errorf("format quantity: %w", err)
	}

	var barcode *string
	if data.Barcode.Valid {
		value := data.Barcode.String
		barcode = &value
	}

	return CatalogProductResponse{
		ID:        data.ID.String(),
		SKU:       data.SKU,
		Barcode:   barcode,
		Name:      data.Name,
		Price:     price,
		Quantity:  quantity,
		IsActive:  data.IsActive,
		InStock:   data.InStock,
		CreatedAt: timestampOrZero(data.CreatedAt),
		UpdatedAt: timestampOrZero(data.UpdatedAt),
	}, nil
}

func toCatalogProductDataFromListRow(row database.ListCatalogProductsRow) CatalogProductData {
	return CatalogProductData{
		ID:        row.ID,
		SKU:       row.SKU,
		Barcode:   row.Barcode,
		Name:      row.Name,
		Price:     row.Price,
		Quantity:  row.Quantity,
		IsActive:  row.IsActive,
		InStock:   row.InStock,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}

func toCatalogProductDataFromIDRow(row database.GetCatalogProductByIDRow) CatalogProductData {
	return CatalogProductData{
		ID:        row.ID,
		SKU:       row.SKU,
		Barcode:   row.Barcode,
		Name:      row.Name,
		Price:     row.Price,
		Quantity:  row.Quantity,
		IsActive:  row.IsActive,
		InStock:   row.InStock,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}

func toCatalogProductDataFromBarcodeRow(row database.GetCatalogProductByBarcodeRow) CatalogProductData {
	return CatalogProductData{
		ID:        row.ID,
		SKU:       row.SKU,
		Barcode:   row.Barcode,
		Name:      row.Name,
		Price:     row.Price,
		Quantity:  row.Quantity,
		IsActive:  row.IsActive,
		InStock:   row.InStock,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}

func NumericToMoneyString(value pgtype.Numeric) (string, error) {
	return numericToFixedScaleString(value, 2)
}

func NumericToQuantityString(value pgtype.Numeric) (string, error) {
	return numericToFixedScaleString(value, 3)
}

func numericToFixedScaleString(value pgtype.Numeric, scale int) (string, error) {
	if !value.Valid {
		return "", fmt.Errorf("numeric value is null")
	}
	if value.NaN {
		return "", fmt.Errorf("numeric value is NaN")
	}
	if value.InfinityModifier != 0 {
		return "", fmt.Errorf("numeric value is infinite")
	}

	plain, err := numericToPlainString(value)
	if err != nil {
		return "", err
	}

	whole, fraction, hasFraction := strings.Cut(plain, ".")
	if !hasFraction {
		return whole + "." + strings.Repeat("0", scale), nil
	}

	fraction = strings.TrimRight(fraction, "0")
	if len(fraction) > scale {
		return "", fmt.Errorf("numeric value has more than %d decimal places", scale)
	}

	for len(fraction) < scale {
		fraction += "0"
	}

	return whole + "." + fraction, nil
}

func numericToPlainString(value pgtype.Numeric) (string, error) {
	digits := "0"
	if value.Int != nil {
		digits = value.Int.String()
	}

	sign := ""
	if strings.HasPrefix(digits, "-") {
		sign = "-"
		digits = strings.TrimPrefix(digits, "-")
	}

	if value.Exp >= 0 {
		return sign + digits + strings.Repeat("0", int(value.Exp)), nil
	}

	scale := int(-value.Exp)
	if len(digits) > scale {
		whole := digits[:len(digits)-scale]
		fraction := digits[len(digits)-scale:]
		return sign + whole + "." + fraction, nil
	}

	return sign + "0." + strings.Repeat("0", scale-len(digits)) + digits, nil
}

func optionalText(value string) pgtype.Text {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return pgtype.Text{}
	}

	return pgtype.Text{
		String: trimmed,
		Valid:  true,
	}
}

func timestampOrZero(value pgtype.Timestamptz) time.Time {
	if !value.Valid {
		return time.Time{}
	}

	return value.Time.UTC()
}

func NewPaginationResponse(page, pageSize int, total int64) PaginationResponse {
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
