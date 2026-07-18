package payments

import (
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/jackc/pgx/v5/pgtype"
)

func toSalePaymentResponse(payment database.Payment, method database.PaymentMethod) (SalePaymentResponse, error) {
	amount, err := numericToMoneyString(payment.Amount)
	if err != nil {
		return SalePaymentResponse{}, fmt.Errorf("format amount: %w", err)
	}

	paidAt := (*time.Time)(nil)
	if payment.PaidAt.Valid {
		t := payment.PaidAt.Time.UTC()
		paidAt = &t
	}

	receivedAmount, err := nullableMoneyToString(payment.ReceivedAmount)
	if err != nil {
		return SalePaymentResponse{}, fmt.Errorf("format received amount: %w", err)
	}

	changeAmount, err := nullableMoneyToString(payment.ChangeAmount)
	if err != nil {
		return SalePaymentResponse{}, fmt.Errorf("format change amount: %w", err)
	}

	return SalePaymentResponse{
		ID:                payment.ID.String(),
		SaleID:            payment.SaleID.String(),
		PaymentMethodID:   method.ID.String(),
		PaymentMethodCode: method.Code,
		PaymentMethodName: method.Name,
		Status:            string(payment.Status),
		Amount:            amount,
		ReceivedAmount:    receivedAmount,
		ChangeAmount:      changeAmount,
		Installments:      payment.Installments,
		ExternalReference: externalReferenceOrNull(payment.ExternalReference),
		PaidAt:            paidAt,
		CreatedAt:         timestampOrZero(payment.CreatedAt),
		UpdatedAt:         timestampOrZero(payment.UpdatedAt),
	}, nil
}

func toPaymentMethodResponse(method database.PaymentMethod) PaymentMethodResponse {
	return PaymentMethodResponse{
		ID:                 method.ID.String(),
		Code:               method.Code,
		Name:               method.Name,
		Kind:               string(method.Kind),
		IsActive:           method.IsActive,
		AllowsChange:       method.AllowsChange,
		AllowsInstallments: method.AllowsInstallments,
		MaxInstallments:    method.MaxInstallments,
		FeePercentage:      feePercentageOrZero(method.FeePercentage),
		CreatedAt:          timestampOrZero(method.CreatedAt),
		UpdatedAt:          timestampOrZero(method.UpdatedAt),
	}
}

func feePercentageOrZero(value pgtype.Numeric) string {
	if !value.Valid {
		return "0"
	}
	intVal, err := numericToScaledInt(value, 2)
	if err != nil {
		return "0"
	}
	return scaledIntToString(intVal, 2)
}

func externalReferenceOrNull(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	result := value.String
	return &result
}

func nullableMoneyToString(value pgtype.Numeric) (*string, error) {
	if !value.Valid {
		return nil, nil
	}
	s, err := numericToMoneyString(value)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func numericToMoneyString(value pgtype.Numeric) (string, error) {
	intVal, err := numericToScaledInt(value, 2)
	if err != nil {
		return "", err
	}
	return scaledIntToString(intVal, 2), nil
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

func timestampOrZero(value pgtype.Timestamptz) time.Time {
	if !value.Valid {
		return time.Time{}
	}
	return value.Time.UTC()
}
