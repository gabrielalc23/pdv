package checkout

import (
	"fmt"
	"time"

	"github.com/gabrielalc23/pdv/internal/fiscal"
	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/jackc/pgx/v5/pgtype"
)

type checkoutState struct {
	sale           database.CompleteSaleRow
	items          []database.SaleItem
	payments       []checkoutPaymentResult
	fiscalDocument FiscalDocumentResponse
}

type checkoutPaymentResult struct {
	row    database.ApprovePaymentRow
	method database.PaymentMethod
}

func toCheckoutResponse(state checkoutState) (CheckoutResponse, error) {
	sale, err := toCheckoutSaleResponse(state.sale)
	if err != nil {
		return CheckoutResponse{}, err
	}

	payments := make([]CheckoutPaymentResponse, 0, len(state.payments))
	for _, payment := range state.payments {
		item, err := toCheckoutPaymentResponse(payment.row, payment.method)
		if err != nil {
			return CheckoutResponse{}, err
		}
		payments = append(payments, item)
	}

	return CheckoutResponse{
		Sale:           sale,
		Payments:       payments,
		FiscalDocument: state.fiscalDocument,
	}, nil
}

func toCheckoutSaleResponse(row database.CompleteSaleRow) (CheckoutSaleResponse, error) {
	subtotal, err := moneyToString(row.Subtotal)
	if err != nil {
		return CheckoutSaleResponse{}, fmt.Errorf("format subtotal: %w", err)
	}

	discount, err := moneyToString(row.Discount)
	if err != nil {
		return CheckoutSaleResponse{}, fmt.Errorf("format discount: %w", err)
	}

	addition, err := moneyToString(row.Addition)
	if err != nil {
		return CheckoutSaleResponse{}, fmt.Errorf("format addition: %w", err)
	}

	total, err := moneyToString(row.Total)
	if err != nil {
		return CheckoutSaleResponse{}, fmt.Errorf("format total: %w", err)
	}

	return CheckoutSaleResponse{
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

func toCheckoutPaymentResponse(row database.ApprovePaymentRow, method database.PaymentMethod) (CheckoutPaymentResponse, error) {
	amount, err := moneyToString(row.Amount)
	if err != nil {
		return CheckoutPaymentResponse{}, fmt.Errorf("format amount: %w", err)
	}

	var receivedAmount *string
	if row.ReceivedAmount.Valid {
		value, err := moneyToString(row.ReceivedAmount)
		if err != nil {
			return CheckoutPaymentResponse{}, fmt.Errorf("format received amount: %w", err)
		}
		receivedAmount = &value
	}

	var changeAmount *string
	if row.ChangeAmount.Valid {
		value, err := moneyToString(row.ChangeAmount)
		if err != nil {
			return CheckoutPaymentResponse{}, fmt.Errorf("format change amount: %w", err)
		}
		changeAmount = &value
	}

	var externalReference *string
	if row.ExternalReference.Valid {
		value := row.ExternalReference.String
		externalReference = &value
	}

	return CheckoutPaymentResponse{
		ID:                row.ID.String(),
		SaleID:            row.SaleID.String(),
		PaymentMethodID:   row.PaymentMethodID.String(),
		PaymentMethodCode: method.Code,
		PaymentMethodName: method.Name,
		PaymentMethodKind: string(method.Kind),
		Amount:            amount,
		ReceivedAmount:    receivedAmount,
		ChangeAmount:      changeAmount,
		Status:            string(row.Status),
		Installments:      row.Installments,
		ExternalReference: externalReference,
		PaidAt:            timestampOrZero(row.PaidAt),
		CreatedAt:         timestampOrZero(row.CreatedAt),
		UpdatedAt:         timestampOrZero(row.UpdatedAt),
	}, nil
}

func toFiscalDocumentResponse(row database.FiscalDocument) FiscalDocumentResponse {
	var series *int32
	if row.Series.Valid {
		value := row.Series.Int32
		series = &value
	}

	var number *int64
	if row.Number.Valid {
		value := row.Number.Int64
		number = &value
	}

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

	var cancelledAt *time.Time
	if row.CancelledAt.Valid {
		value := row.CancelledAt.Time.UTC()
		cancelledAt = &value
	}

	return FiscalDocumentResponse{
		ID:                row.ID.String(),
		SaleID:            row.SaleID.String(),
		Status:            string(row.Status),
		Environment:       string(row.Environment),
		DocumentModel:     row.DocumentModel,
		Series:            series,
		Number:            number,
		AccessKey:         accessKey,
		Protocol:          protocol,
		Provider:          provider,
		ExternalReference: externalReference,
		ErrorCode:         errorCode,
		ErrorMessage:      errorMessage,
		IssuedAt:          issuedAt,
		CancelledAt:       cancelledAt,
		CreatedAt:         timestampOrZero(row.CreatedAt),
		UpdatedAt:         timestampOrZero(row.UpdatedAt),
	}
}

func toFiscalAuthorizationInput(state checkoutState) (fiscal.AuthorizationInput, error) {
	saleTotal, err := moneyToString(state.sale.Total)
	if err != nil {
		return fiscal.AuthorizationInput{}, fmt.Errorf("format sale total: %w", err)
	}

	return fiscal.AuthorizationInput{
		SaleID:        state.sale.ID.String(),
		SaleNumber:    state.sale.Number,
		SaleTotal:     saleTotal,
		ItemCount:     len(state.items),
		PaymentCount:  len(state.payments),
		CompletedAt:   timestampOrZero(state.sale.CompletedAt),
		Environment:   "HOMOLOGATION",
		DocumentModel: 65,
	}, nil
}

func timestampOrZero(value pgtype.Timestamptz) time.Time {
	if !value.Valid {
		return time.Time{}
	}

	return value.Time.UTC()
}
