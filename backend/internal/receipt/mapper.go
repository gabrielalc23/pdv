package receipt

import (
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/jackc/pgx/v5/pgtype"
)

func toReceiptResponse(
	sale database.Sale,
	items []database.SaleItem,
	payments []database.ReceiptPayment,
	fiscalDoc database.FiscalDocument,
) (ReceiptResponse, error) {
	saleResp, err := toReceiptSaleResponse(sale)
	if err != nil {
		return ReceiptResponse{}, err
	}

	itemResponses := make([]ReceiptItemResponse, 0, len(items))
	for _, item := range items {
		unitPrice, err := numericToMoneyString(item.UnitPrice)
		if err != nil {
			return ReceiptResponse{}, fmt.Errorf("format unit price: %w", err)
		}

		quantity, err := numericToQuantityString(item.Quantity)
		if err != nil {
			return ReceiptResponse{}, fmt.Errorf("format quantity: %w", err)
		}

		discount, err := numericToMoneyString(item.Discount)
		if err != nil {
			return ReceiptResponse{}, fmt.Errorf("format item discount: %w", err)
		}

		itemTotal, err := numericToMoneyString(item.Total)
		if err != nil {
			return ReceiptResponse{}, fmt.Errorf("format item total: %w", err)
		}

		itemSubtotal, err := multiplyMoneyQuantity(item.UnitPrice, item.Quantity)
		if err != nil {
			return ReceiptResponse{}, fmt.Errorf("calculate item subtotal: %w", err)
		}
		subtotal, err := numericToMoneyString(itemSubtotal)
		if err != nil {
			return ReceiptResponse{}, fmt.Errorf("format item subtotal: %w", err)
		}

		itemResponses = append(itemResponses, ReceiptItemResponse{
			ProductID: item.ProductID.String(),
			SKU:       item.ProductSKU,
			Name:      item.ProductName,
			UnitPrice: unitPrice,
			Quantity:  quantity,
			Subtotal:  subtotal,
			Discount:  discount,
			Total:     itemTotal,
			CreatedAt: timestampOrZero(item.CreatedAt),
		})
	}

	paymentResponses := make([]ReceiptPaymentResponse, 0, len(payments))
	for _, p := range payments {
		amount, err := numericToMoneyString(p.Amount)
		if err != nil {
			return ReceiptResponse{}, fmt.Errorf("format payment amount: %w", err)
		}

		paymentResponses = append(paymentResponses, ReceiptPaymentResponse{
			Method:       p.PaymentMethodName,
			Amount:       amount,
			Status:       string(p.Status),
			Installments: p.Installments,
		})
	}

	fiscalResp := toReceiptFiscalResponse(fiscalDoc)

	return ReceiptResponse{
		Sale:           saleResp,
		Items:          itemResponses,
		Payments:       paymentResponses,
		FiscalDocument: &fiscalResp,
	}, nil
}

func toReceiptSaleResponse(sale database.Sale) (ReceiptSaleResponse, error) {
	subtotal, err := numericToMoneyString(sale.Subtotal)
	if err != nil {
		return ReceiptSaleResponse{}, fmt.Errorf("format subtotal: %w", err)
	}

	discount, err := numericToMoneyString(sale.Discount)
	if err != nil {
		return ReceiptSaleResponse{}, fmt.Errorf("format discount: %w", err)
	}

	addition, err := numericToMoneyString(sale.Addition)
	if err != nil {
		return ReceiptSaleResponse{}, fmt.Errorf("format addition: %w", err)
	}

	total, err := numericToMoneyString(sale.Total)
	if err != nil {
		return ReceiptSaleResponse{}, fmt.Errorf("format total: %w", err)
	}

	return ReceiptSaleResponse{
		ID:             sale.ID.String(),
		Number:         sale.Number,
		Status:         string(sale.Status),
		Subtotal:       subtotal,
		Discount:       discount,
		Addition:       addition,
		Total:          total,
		OpenedAt:       timestampOrZero(sale.OpenedAt),
		CompletedAt:    timestampOrZero(sale.CompletedAt),
		CancelledAt:    timestampOrZero(sale.CancelledAt),
		CreatedAt:      timestampOrZero(sale.CreatedAt),
		UpdatedAt:      timestampOrZero(sale.UpdatedAt),
		IdempotencyKey: sale.IdempotencyKey,
	}, nil
}

func toReceiptFiscalResponse(doc database.FiscalDocument) ReceiptFiscalResponse {
	accessKey := nullString(doc.AccessKey)
	protocol := nullString(doc.Protocol)
	provider := nullString(doc.Provider)
	externalReference := nullString(doc.ExternalReference)
	errorCode := nullString(doc.ErrorCode)
	errorMessage := nullString(doc.ErrorMessage)
	issuedAt := nullTime(doc.IssuedAt)

	return ReceiptFiscalResponse{
		Status:            string(doc.Status),
		AccessKey:         accessKey,
		Protocol:          protocol,
		Provider:          provider,
		ExternalReference: externalReference,
		ErrorCode:         errorCode,
		ErrorMessage:      errorMessage,
		IssuedAt:          issuedAt,
	}
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

func multiplyMoneyQuantity(unitPrice, quantity pgtype.Numeric) (pgtype.Numeric, error) {
	if !unitPrice.Valid || !quantity.Valid {
		return pgtype.Numeric{}, fmt.Errorf("numeric value is null")
	}
	if unitPrice.Int == nil || quantity.Int == nil {
		return numericFromScaledInt(nil, 2), nil
	}
	product := new(big.Int).Mul(new(big.Int).Set(unitPrice.Int), new(big.Int).Set(quantity.Int))
	exp := unitPrice.Exp + quantity.Exp
	return roundNumeric(product, exp, 2), nil
}

func numericFromScaledInt(value *big.Int, scale int32) pgtype.Numeric {
	if value == nil {
		value = big.NewInt(0)
	}
	return pgtype.Numeric{Int: new(big.Int).Set(value), Exp: -scale, Valid: true}
}

func roundNumeric(intVal *big.Int, exp int32, scale int32) pgtype.Numeric {
	if intVal == nil {
		intVal = big.NewInt(0)
	}
	targetExp := -scale
	coeff := new(big.Int).Set(intVal)

	switch {
	case exp == targetExp:
		return numericFromScaledInt(coeff, scale)
	case exp > targetExp:
		pow := pow10(int(exp - targetExp))
		coeff.Mul(coeff, pow)
		return numericFromScaledInt(coeff, scale)
	default:
		divisor := pow10(int(targetExp - exp))
		quotient, remainder := new(big.Int).QuoRem(coeff, divisor, new(big.Int))
		if remainder.Sign() != 0 {
			twiceRemainder := new(big.Int).Lsh(remainder, 1)
			if twiceRemainder.Cmp(divisor) >= 0 {
				quotient.Add(quotient, big.NewInt(1))
			}
		}
		return numericFromScaledInt(quotient, scale)
	}
}

func timestampOrZero(value pgtype.Timestamptz) time.Time {
	if !value.Valid {
		return time.Time{}
	}
	return value.Time.UTC()
}

func nullString(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

func nullTime(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	t := value.Time.UTC()
	return &t
}
