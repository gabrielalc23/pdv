package sales

import (
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/jackc/pgx/v5/pgtype"
)

func toSaleItemResponse(item database.SaleItem) (SaleItemResponse, error) {
	unitPrice, err := numericToMoneyString(item.UnitPrice)
	if err != nil {
		return SaleItemResponse{}, fmt.Errorf("format unit price: %w", err)
	}

	quantity, err := numericToQuantityString(item.Quantity)
	if err != nil {
		return SaleItemResponse{}, fmt.Errorf("format quantity: %w", err)
	}

	discount, err := numericToMoneyString(item.Discount)
	if err != nil {
		return SaleItemResponse{}, fmt.Errorf("format discount: %w", err)
	}

	total, err := numericToMoneyString(item.Total)
	if err != nil {
		return SaleItemResponse{}, fmt.Errorf("format total: %w", err)
	}

	return SaleItemResponse{
		ID:          item.ID.String(),
		SaleID:      item.SaleID.String(),
		ProductID:   item.ProductID.String(),
		ProductName: item.ProductName,
		ProductSKU:  item.ProductSKU,
		UnitPrice:   unitPrice,
		Quantity:    quantity,
		Discount:    discount,
		Total:       total,
		CreatedAt:   timestampOrZero(item.CreatedAt),
	}, nil
}

func toSaleResponseFromColumns(
	id pgtype.UUID,
	number int64,
	status database.SaleStatus,
	subtotal, discount, addition, total pgtype.Numeric,
	openedAt, completedAt, cancelledAt, createdAt, updatedAt pgtype.Timestamptz,
	idempotencyKey string,
	items []database.SaleItem,
) (SaleResponse, error) {
	header, err := toSaleHeaderResponseFromColumns(
		id,
		number,
		status,
		subtotal,
		discount,
		addition,
		total,
		openedAt,
		completedAt,
		cancelledAt,
		createdAt,
		updatedAt,
		idempotencyKey,
	)
	if err != nil {
		return SaleResponse{}, err
	}

	saleItems := make([]SaleItemResponse, 0, len(items))
	for _, item := range items {
		response, err := toSaleItemResponse(item)
		if err != nil {
			return SaleResponse{}, fmt.Errorf("map sale item: %w", err)
		}
		saleItems = append(saleItems, response)
	}

	return SaleResponse{
		ID:             header.ID,
		Number:         header.Number,
		Status:         header.Status,
		Subtotal:       header.Subtotal,
		Discount:       header.Discount,
		Addition:       header.Addition,
		Total:          header.Total,
		OpenedAt:       header.OpenedAt,
		CompletedAt:    header.CompletedAt,
		CancelledAt:    header.CancelledAt,
		CreatedAt:      header.CreatedAt,
		UpdatedAt:      header.UpdatedAt,
		IdempotencyKey: header.IdempotencyKey,
		Items:          saleItems,
	}, nil
}

func toSaleHeaderResponseFromColumns(
	id pgtype.UUID,
	number int64,
	status database.SaleStatus,
	subtotal, discount, addition, total pgtype.Numeric,
	openedAt, completedAt, cancelledAt, createdAt, updatedAt pgtype.Timestamptz,
	idempotencyKey string,
) (SaleListItemResponse, error) {
	subtotalString, err := numericToMoneyString(subtotal)
	if err != nil {
		return SaleListItemResponse{}, fmt.Errorf("format subtotal: %w", err)
	}

	discountString, err := numericToMoneyString(discount)
	if err != nil {
		return SaleListItemResponse{}, fmt.Errorf("format discount: %w", err)
	}

	additionString, err := numericToMoneyString(addition)
	if err != nil {
		return SaleListItemResponse{}, fmt.Errorf("format addition: %w", err)
	}

	totalString, err := numericToMoneyString(total)
	if err != nil {
		return SaleListItemResponse{}, fmt.Errorf("format total: %w", err)
	}

	return SaleListItemResponse{
		ID:             id.String(),
		Number:         number,
		Status:         string(status),
		Subtotal:       subtotalString,
		Discount:       discountString,
		Addition:       additionString,
		Total:          totalString,
		OpenedAt:       timestampOrZero(openedAt),
		CompletedAt:    timestampOrZero(completedAt),
		CancelledAt:    timestampOrZero(cancelledAt),
		CreatedAt:      timestampOrZero(createdAt),
		UpdatedAt:      timestampOrZero(updatedAt),
		IdempotencyKey: idempotencyKey,
	}, nil
}

func toSaleListItemResponse(row database.ListSalesRow) (SaleListItemResponse, error) {
	return toSaleHeaderResponseFromColumns(
		row.ID,
		row.Number,
		row.Status,
		row.Subtotal,
		row.Discount,
		row.Addition,
		row.Total,
		row.OpenedAt,
		row.CompletedAt,
		row.CancelledAt,
		row.CreatedAt,
		row.UpdatedAt,
		row.IdempotencyKey,
	)
}

func numericToMoneyString(value pgtype.Numeric) (string, error) {
	intVal, err := numericToScaledInt(value, 2)
	if err != nil {
		return "", err
	}

	return scaledIntToString(intVal, 2), nil
}

func numericToQuantityString(value pgtype.Numeric) (string, error) {
	intVal, err := numericToScaledInt(value, 3)
	if err != nil {
		return "", err
	}

	return scaledIntToString(intVal, 3), nil
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

func numericFromScaledInt(value *big.Int, scale int32) pgtype.Numeric {
	if value == nil {
		value = big.NewInt(0)
	}

	return pgtype.Numeric{
		Int:   new(big.Int).Set(value),
		Exp:   -scale,
		Valid: true,
	}
}

func numericToScaledInt(value pgtype.Numeric, scale int32) (*big.Int, error) {
	if !value.Valid {
		return nil, fmt.Errorf("numeric value is null")
	}

	if value.Int == nil {
		return big.NewInt(0), nil
	}

	if value.NaN {
		return nil, fmt.Errorf("numeric value is NaN")
	}

	if value.InfinityModifier != 0 {
		return nil, fmt.Errorf("numeric value is infinite")
	}

	intVal := new(big.Int).Set(value.Int)
	targetExp := -scale

	switch {
	case value.Exp == targetExp:
		return intVal, nil
	case value.Exp > targetExp:
		pow := pow10(int(value.Exp - targetExp))
		return intVal.Mul(intVal, pow), nil
	default:
		divisor := pow10(int(targetExp - value.Exp))
		quotient, remainder := new(big.Int).QuoRem(intVal, divisor, new(big.Int))
		if remainder.Sign() != 0 {
			twiceRemainder := new(big.Int).Lsh(remainder, 1)
			if twiceRemainder.Cmp(divisor) >= 0 {
				quotient.Add(quotient, big.NewInt(1))
			}
		}
		return quotient, nil
	}
}

func scaledIntToString(value *big.Int, scale int32) string {
	if value == nil {
		value = big.NewInt(0)
	}

	sign := ""
	if value.Sign() < 0 {
		sign = "-"
		value = new(big.Int).Abs(value)
	}

	if scale == 0 {
		return sign + value.String()
	}

	digits := value.String()
	if len(digits) <= int(scale) {
		digits = strings.Repeat("0", int(scale)-len(digits)+1) + digits
	}

	cut := len(digits) - int(scale)
	return sign + digits[:cut] + "." + digits[cut:]
}

func pow10(exp int) *big.Int {
	if exp <= 0 {
		return big.NewInt(1)
	}

	return new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(exp)), nil)
}

func compareMoney(a, b pgtype.Numeric) (int, error) {
	left, err := numericToScaledInt(a, 2)
	if err != nil {
		return 0, err
	}

	right, err := numericToScaledInt(b, 2)
	if err != nil {
		return 0, err
	}

	return left.Cmp(right), nil
}

func multiplyMoneyQuantity(unitPrice, quantity pgtype.Numeric) (pgtype.Numeric, error) {
	if !unitPrice.Valid {
		return pgtype.Numeric{}, fmt.Errorf("numeric value is null")
	}
	if !quantity.Valid {
		return pgtype.Numeric{}, fmt.Errorf("numeric value is null")
	}

	if unitPrice.Int == nil || quantity.Int == nil {
		return zeroMoney(), nil
	}

	product := new(big.Int).Mul(new(big.Int).Set(unitPrice.Int), new(big.Int).Set(quantity.Int))
	exp := unitPrice.Exp + quantity.Exp
	return roundNumeric(product, exp, 2)
}

func subtractMoney(minuend, subtrahend pgtype.Numeric) (pgtype.Numeric, error) {
	left, err := numericToScaledInt(minuend, 2)
	if err != nil {
		return pgtype.Numeric{}, err
	}

	right, err := numericToScaledInt(subtrahend, 2)
	if err != nil {
		return pgtype.Numeric{}, err
	}

	result := new(big.Int).Sub(left, right)
	return numericFromScaledInt(result, 2), nil
}

func sumSaleTotals(items []database.SaleItem) (pgtype.Numeric, pgtype.Numeric, pgtype.Numeric, error) {
	subtotal := big.NewInt(0)
	discount := big.NewInt(0)
	total := big.NewInt(0)

	for _, item := range items {
		itemSubtotal, err := multiplyMoneyQuantity(item.UnitPrice, item.Quantity)
		if err != nil {
			return pgtype.Numeric{}, pgtype.Numeric{}, pgtype.Numeric{}, fmt.Errorf("calculate item subtotal: %w", err)
		}

		itemSubtotalInt, err := numericToScaledInt(itemSubtotal, 2)
		if err != nil {
			return pgtype.Numeric{}, pgtype.Numeric{}, pgtype.Numeric{}, fmt.Errorf("normalize item subtotal: %w", err)
		}

		itemDiscountInt, err := numericToScaledInt(item.Discount, 2)
		if err != nil {
			return pgtype.Numeric{}, pgtype.Numeric{}, pgtype.Numeric{}, fmt.Errorf("normalize item discount: %w", err)
		}

		itemTotalInt, err := numericToScaledInt(item.Total, 2)
		if err != nil {
			return pgtype.Numeric{}, pgtype.Numeric{}, pgtype.Numeric{}, fmt.Errorf("normalize item total: %w", err)
		}

		subtotal.Add(subtotal, itemSubtotalInt)
		discount.Add(discount, itemDiscountInt)
		total.Add(total, itemTotalInt)
	}

	return numericFromScaledInt(subtotal, 2), numericFromScaledInt(discount, 2), numericFromScaledInt(total, 2), nil
}

func roundNumeric(intVal *big.Int, exp int32, scale int32) (pgtype.Numeric, error) {
	if intVal == nil {
		intVal = big.NewInt(0)
	}

	targetExp := -scale
	coeff := new(big.Int).Set(intVal)

	switch {
	case exp == targetExp:
		return numericFromScaledInt(coeff, scale), nil
	case exp > targetExp:
		pow := pow10(int(exp - targetExp))
		coeff.Mul(coeff, pow)
		return numericFromScaledInt(coeff, scale), nil
	default:
		divisor := pow10(int(targetExp - exp))
		quotient, remainder := new(big.Int).QuoRem(coeff, divisor, new(big.Int))
		if remainder.Sign() != 0 {
			twiceRemainder := new(big.Int).Lsh(remainder, 1)
			if twiceRemainder.Cmp(divisor) >= 0 {
				quotient.Add(quotient, big.NewInt(1))
			}
		}
		return numericFromScaledInt(quotient, scale), nil
	}
}
