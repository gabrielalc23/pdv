package payments

import (
	"fmt"
	"strings"
	"time"

	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/jackc/pgx/v5/pgtype"
)

func toPaymentMethodResponse(row database.PaymentMethod) (PaymentMethodResponse, error) {
	fee, err := numericToString(row.FeePercentage, 4)
	if err != nil {
		return PaymentMethodResponse{}, fmt.Errorf("format fee percentage: %w", err)
	}

	return PaymentMethodResponse{
		ID:                 row.ID.String(),
		Code:               row.Code,
		Name:               row.Name,
		Kind:               string(row.Kind),
		IsActive:           row.IsActive,
		AllowsChange:       row.AllowsChange,
		AllowsInstallments: row.AllowsInstallments,
		MaxInstallments:    row.MaxInstallments,
		FeePercentage:      fee,
		CreatedAt:          timestampOrZero(row.CreatedAt),
		UpdatedAt:          timestampOrZero(row.UpdatedAt),
	}, nil
}

func toSalePaymentResponse(row database.ListPaymentsBySaleIDRow, method database.PaymentMethod) (SalePaymentResponse, error) {
	amount, err := moneyToString(row.Amount)
	if err != nil {
		return SalePaymentResponse{}, fmt.Errorf("format amount: %w", err)
	}

	var receivedAmount *string
	if row.ReceivedAmount.Valid {
		value, err := moneyToString(row.ReceivedAmount)
		if err != nil {
			return SalePaymentResponse{}, fmt.Errorf("format received amount: %w", err)
		}
		receivedAmount = &value
	}

	var changeAmount *string
	if row.ChangeAmount.Valid {
		value, err := moneyToString(row.ChangeAmount)
		if err != nil {
			return SalePaymentResponse{}, fmt.Errorf("format change amount: %w", err)
		}
		changeAmount = &value
	}

	var externalReference *string
	if row.ExternalReference.Valid {
		value := row.ExternalReference.String
		externalReference = &value
	}

	var paidAt *time.Time
	if row.PaidAt.Valid {
		value := row.PaidAt.Time.UTC()
		paidAt = &value
	}

	return SalePaymentResponse{
		ID:                row.ID.String(),
		SaleID:            row.SaleID.String(),
		PaymentMethodID:   row.PaymentMethodID.String(),
		PaymentMethodCode: method.Code,
		PaymentMethodName: method.Name,
		Amount:            amount,
		ReceivedAmount:    receivedAmount,
		ChangeAmount:      changeAmount,
		Status:            string(row.Status),
		Installments:      row.Installments,
		ExternalReference: externalReference,
		PaidAt:            paidAt,
		CreatedAt:         timestampOrZero(row.CreatedAt),
		UpdatedAt:         timestampOrZero(row.UpdatedAt),
	}, nil
}

func moneyToString(value pgtype.Numeric) (string, error) {
	return numericToString(value, 2)
}

func numericToString(value pgtype.Numeric, scale int) (string, error) {
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
		return whole + "." + strings.Repeat("0", scale), nil
	}

	if len(fraction) > scale {
		return "", fmt.Errorf("numeric value has more than %d decimal places", scale)
	}
	for len(fraction) < scale {
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
