package receipt

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/jackc/pgx/v5/pgtype"
)

func toReceiptResponse(ctx context.Context, sale database.GetSaleByIDRow, items []database.SaleItem, payments []database.ListPaymentsBySaleIDRow, fiscalDoc database.FiscalDocument, store Store) (ReceiptResponse, error) {
	saleResponse, err := toReceiptSaleResponse(sale)
	if err != nil {
		return ReceiptResponse{}, err
	}

	itemResponses := make([]ReceiptItemResponse, 0, len(items))
	for _, item := range items {
		response, err := toReceiptItemResponse(item)
		if err != nil {
			return ReceiptResponse{}, err
		}
		itemResponses = append(itemResponses, response)
	}

	paymentResponses := make([]ReceiptPaymentResponse, 0, len(payments))
	for _, payment := range payments {
		method, err := store.GetPaymentMethodByID(ctx, payment.PaymentMethodID)
		if err != nil {
			return ReceiptResponse{}, fmt.Errorf("get payment method: %w", err)
		}

		response, err := toReceiptPaymentResponse(payment, method)
		if err != nil {
			return ReceiptResponse{}, err
		}
		paymentResponses = append(paymentResponses, response)
	}

	fiscalResponse := toReceiptFiscalResponse(fiscalDoc)
	return ReceiptResponse{
		Sale:           saleResponse,
		Items:          itemResponses,
		Payments:       paymentResponses,
		FiscalDocument: &fiscalResponse,
	}, nil
}

func moneyToString(value pgtype.Numeric) (string, error) {
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

	whole, fraction, hasFraction := strings.Cut(text, ".")
	if !hasFraction {
		return whole + ".00", nil
	}
	if len(fraction) > 2 {
		return "", fmt.Errorf("money value has more than two decimal places")
	}
	if len(fraction) == 1 {
		fraction += "0"
	}
	if len(fraction) == 0 {
		fraction = "00"
	}
	return whole + "." + fraction, nil
}

func quantityToString(value pgtype.Numeric) (string, error) {
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

	whole, fraction, hasFraction := strings.Cut(text, ".")
	if !hasFraction {
		return whole + ".000", nil
	}
	if len(fraction) > 3 {
		return "", fmt.Errorf("quantity value has more than three decimal places")
	}
	for len(fraction) < 3 {
		fraction += "0"
	}
	return whole + "." + fraction, nil
}

func timestampOrZero(value pgtype.Timestamptz) time.Time {
	if !value.Valid {
		return time.Time{}
	}
	return value.Time.UTC()
}

func toReceiptSaleResponse(row database.GetSaleByIDRow) (ReceiptSaleResponse, error) {
	subtotal, err := moneyToString(row.Subtotal)
	if err != nil {
		return ReceiptSaleResponse{}, fmt.Errorf("format subtotal: %w", err)
	}
	discount, err := moneyToString(row.Discount)
	if err != nil {
		return ReceiptSaleResponse{}, fmt.Errorf("format discount: %w", err)
	}
	addition, err := moneyToString(row.Addition)
	if err != nil {
		return ReceiptSaleResponse{}, fmt.Errorf("format addition: %w", err)
	}
	total, err := moneyToString(row.Total)
	if err != nil {
		return ReceiptSaleResponse{}, fmt.Errorf("format total: %w", err)
	}

	return ReceiptSaleResponse{
		ID:             row.ID.String(),
		Number:         row.Number,
		Status:         string(row.Status),
		Subtotal:       subtotal,
		Discount:       discount,
		Addition:       addition,
		Total:          total,
		OpenedAt:       timestampOrZero(row.OpenedAt),
		CompletedAt:    timestampOrZero(row.CompletedAt),
		CancelledAt:    timestampOrZero(row.CancelledAt),
		CreatedAt:      timestampOrZero(row.CreatedAt),
		UpdatedAt:      timestampOrZero(row.UpdatedAt),
		IdempotencyKey: row.IdempotencyKey,
	}, nil
}

func toReceiptItemResponse(row database.SaleItem) (ReceiptItemResponse, error) {
	unitPrice, err := moneyToString(row.UnitPrice)
	if err != nil {
		return ReceiptItemResponse{}, fmt.Errorf("format unit price: %w", err)
	}
	quantity, err := quantityToString(row.Quantity)
	if err != nil {
		return ReceiptItemResponse{}, fmt.Errorf("format quantity: %w", err)
	}
	subtotal, err := multiplyMoneyQuantity(row.UnitPrice, row.Quantity)
	if err != nil {
		return ReceiptItemResponse{}, fmt.Errorf("calculate subtotal: %w", err)
	}
	subtotalValue, err := moneyToString(subtotal)
	if err != nil {
		return ReceiptItemResponse{}, fmt.Errorf("format subtotal: %w", err)
	}
	discount, err := moneyToString(row.Discount)
	if err != nil {
		return ReceiptItemResponse{}, fmt.Errorf("format discount: %w", err)
	}
	total, err := moneyToString(row.Total)
	if err != nil {
		return ReceiptItemResponse{}, fmt.Errorf("format total: %w", err)
	}

	return ReceiptItemResponse{
		ProductID: row.ProductID.String(),
		SKU:       row.ProductSKU,
		Name:      row.ProductName,
		UnitPrice: unitPrice,
		Quantity:  quantity,
		Subtotal:  subtotalValue,
		Discount:  discount,
		Total:     total,
		CreatedAt: timestampOrZero(row.CreatedAt),
	}, nil
}

func toReceiptPaymentResponse(row database.ListPaymentsBySaleIDRow, method database.PaymentMethod) (ReceiptPaymentResponse, error) {
	amount, err := moneyToString(row.Amount)
	if err != nil {
		return ReceiptPaymentResponse{}, fmt.Errorf("format amount: %w", err)
	}

	var receivedAmount *string
	if row.ReceivedAmount.Valid {
		value, err := moneyToString(row.ReceivedAmount)
		if err != nil {
			return ReceiptPaymentResponse{}, fmt.Errorf("format received amount: %w", err)
		}
		receivedAmount = &value
	}

	var changeAmount *string
	if row.ChangeAmount.Valid {
		value, err := moneyToString(row.ChangeAmount)
		if err != nil {
			return ReceiptPaymentResponse{}, fmt.Errorf("format change amount: %w", err)
		}
		changeAmount = &value
	}

	var externalReference *string
	if row.ExternalReference.Valid {
		value := row.ExternalReference.String
		externalReference = &value
	}

	return ReceiptPaymentResponse{
		Method:            method.Name,
		Amount:            amount,
		Status:            string(row.Status),
		Installments:      row.Installments,
		ReceivedAmount:    receivedAmount,
		ChangeAmount:      changeAmount,
		ExternalReference: externalReference,
	}, nil
}

func toReceiptFiscalResponse(row database.FiscalDocument) ReceiptFiscalResponse {
	var accessKey *string
	if row.AccessKey.Valid {
		value := row.AccessKey.String
		accessKey = &value
	}
	var protocol *string
	if row.Protocol.Valid {
		value := row.Protocol.String
		protocol = &value
	}
	var provider *string
	if row.Provider.Valid {
		value := row.Provider.String
		provider = &value
	}
	var externalReference *string
	if row.ExternalReference.Valid {
		value := row.ExternalReference.String
		externalReference = &value
	}
	var errorCode *string
	if row.ErrorCode.Valid {
		value := row.ErrorCode.String
		errorCode = &value
	}
	var errorMessage *string
	if row.ErrorMessage.Valid {
		value := row.ErrorMessage.String
		errorMessage = &value
	}
	var issuedAt *time.Time
	if row.IssuedAt.Valid {
		value := row.IssuedAt.Time.UTC()
		issuedAt = &value
	}
	return ReceiptFiscalResponse{
		Status:            string(row.Status),
		AccessKey:         accessKey,
		Protocol:          protocol,
		Provider:          provider,
		ExternalReference: externalReference,
		ErrorCode:         errorCode,
		ErrorMessage:      errorMessage,
		IssuedAt:          issuedAt,
	}
}
