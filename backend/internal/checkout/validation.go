package checkout

import (
	"math/big"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
)

type normalizedCheckoutPaymentInput struct {
	PaymentMethodID   pgtype.UUID
	Amount            pgtype.Numeric
	ReceivedAmount    *pgtype.Numeric
	Installments      int16
	ExternalReference *string
}

type normalizedCheckoutInput struct {
	Payments []normalizedCheckoutPaymentInput
}

func normalizeCheckoutInput(input CheckoutInput) (normalizedCheckoutInput, error) {
	payments := make([]normalizedCheckoutPaymentInput, 0, len(input.Payments))
	for i, payment := range input.Payments {
		normalized, err := normalizeCheckoutPaymentInput(payment, i)
		if err != nil {
			return normalizedCheckoutInput{}, err
		}
		payments = append(payments, normalized)
	}

	return normalizedCheckoutInput{Payments: payments}, nil
}

func normalizeCheckoutPaymentInput(input CheckoutPaymentInput, index int) (normalizedCheckoutPaymentInput, error) {
	paymentMethodID, err := parseUUID(input.PaymentMethodID, fieldName(index, "paymentMethodId"))
	if err != nil {
		return normalizedCheckoutPaymentInput{}, err
	}

	amount, err := parseMoney(fieldName(index, "amount"), input.Amount)
	if err != nil {
		return normalizedCheckoutPaymentInput{}, err
	}

	var receivedAmount *pgtype.Numeric
	if input.ReceivedAmount != nil {
		value, err := parseMoney(fieldName(index, "receivedAmount"), *input.ReceivedAmount)
		if err != nil {
			return normalizedCheckoutPaymentInput{}, err
		}
		receivedAmount = &value
	}

	installments := int16(1)
	if input.Installments != nil {
		if *input.Installments > 32767 {
			return normalizedCheckoutPaymentInput{}, newValidationError(fieldName(index, "installments"), "is too large")
		}
		installments = int16(*input.Installments)
	}

	var externalReference *string
	if input.ExternalReference != nil {
		value, err := normalizeRequiredText(fieldName(index, "externalReference"), *input.ExternalReference)
		if err != nil {
			return normalizedCheckoutPaymentInput{}, err
		}
		externalReference = &value
	}

	return normalizedCheckoutPaymentInput{
		PaymentMethodID:   paymentMethodID,
		Amount:            amount,
		ReceivedAmount:    receivedAmount,
		Installments:      installments,
		ExternalReference: externalReference,
	}, nil
}

func normalizeRequiredText(field, value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", newValidationError(field, "is required")
	}
	return trimmed, nil
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

func parseMoney(field, value string) (pgtype.Numeric, error) {
	canonical, err := normalizeMoneyString(field, value)
	if err != nil {
		return pgtype.Numeric{}, err
	}

	var numeric pgtype.Numeric
	if err := numeric.ScanScientific(canonical); err != nil {
		return pgtype.Numeric{}, newValidationError(field, "must be a valid monetary amount")
	}

	return numeric, nil
}

func normalizeMoneyString(field, value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", newValidationError(field, "is required")
	}
	if strings.HasPrefix(trimmed, "-") || strings.HasPrefix(trimmed, "+") {
		return "", newValidationError(field, "cannot be negative")
	}

	whole, fraction, hasFraction := strings.Cut(trimmed, ".")
	if hasFraction {
		if whole == "" || fraction == "" || !allDigits(whole) || !allDigits(fraction) || len(fraction) > 2 {
			return "", newValidationError(field, "must have at most two decimal places")
		}
		if len(fraction) == 1 {
			fraction += "0"
		}
		return whole + "." + fraction, nil
	}

	if !allDigits(trimmed) {
		return "", newValidationError(field, "must be a valid monetary amount")
	}

	return trimmed + ".00", nil
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

func ptrString(value string) *string {
	return &value
}

func fieldName(index int, name string) string {
	return "payments[" + intToString(index) + "]." + name
}

func intToString(value int) string {
	if value == 0 {
		return "0"
	}

	negative := value < 0
	if negative {
		value = -value
	}

	buf := make([]byte, 0, 16)
	for value > 0 {
		buf = append(buf, byte('0'+(value%10)))
		value /= 10
	}

	if negative {
		buf = append(buf, '-')
	}

	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}

	return string(buf)
}

func zeroMoney() pgtype.Numeric {
	return numericFromScaledInt(big.NewInt(0), 2)
}

func numericFromScaledInt(value *big.Int, scale int32) pgtype.Numeric {
	if value == nil {
		value = big.NewInt(0)
	}

	return pgtype.Numeric{Int: new(big.Int).Set(value), Exp: -scale, Valid: true}
}

func numericToScaledInt(value pgtype.Numeric, scale int32) (*big.Int, error) {
	if !value.Valid {
		return nil, newValidationError("", "numeric value is null")
	}

	if value.Int == nil {
		return big.NewInt(0), nil
	}
	if value.NaN {
		return nil, newValidationError("", "numeric value is NaN")
	}
	if value.InfinityModifier != 0 {
		return nil, newValidationError("", "numeric value is infinite")
	}

	intVal := new(big.Int).Set(value.Int)
	targetExp := -scale

	switch {
	case value.Exp == targetExp:
		return intVal, nil
	case value.Exp > targetExp:
		return intVal.Mul(intVal, pow10(int(value.Exp-targetExp))), nil
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

func addMoney(a, b pgtype.Numeric) (pgtype.Numeric, error) {
	left, err := numericToScaledInt(a, 2)
	if err != nil {
		return pgtype.Numeric{}, err
	}

	right, err := numericToScaledInt(b, 2)
	if err != nil {
		return pgtype.Numeric{}, err
	}

	return numericFromScaledInt(new(big.Int).Add(left, right), 2), nil
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

	return numericFromScaledInt(new(big.Int).Sub(left, right), 2), nil
}

func multiplyMoneyQuantity(unitPrice, quantity pgtype.Numeric) (pgtype.Numeric, error) {
	left, err := numericToScaledInt(unitPrice, 2)
	if err != nil {
		return pgtype.Numeric{}, err
	}

	right, err := numericToScaledInt(quantity, 3)
	if err != nil {
		return pgtype.Numeric{}, err
	}

	product := new(big.Int).Mul(left, right)
	return roundNumeric(product, -5, 2), nil
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
		return numericFromScaledInt(coeff.Mul(coeff, pow10(int(exp-targetExp))), scale)
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

func sumPaymentAmounts(payments []normalizedPaymentCalculation) (pgtype.Numeric, error) {
	total := big.NewInt(0)
	for _, payment := range payments {
		value, err := numericToScaledInt(payment.Amount, 2)
		if err != nil {
			return pgtype.Numeric{}, err
		}
		total.Add(total, value)
	}

	return numericFromScaledInt(total, 2), nil
}

type normalizedPaymentCalculation struct {
	Amount pgtype.Numeric
}

func moneyToString(value pgtype.Numeric) (string, error) {
	scaled, err := numericToScaledInt(value, 2)
	if err != nil {
		return "", err
	}
	return scaledIntToString(scaled, 2), nil
}

func quantityToString(value pgtype.Numeric) (string, error) {
	scaled, err := numericToScaledInt(value, 3)
	if err != nil {
		return "", err
	}
	return scaledIntToString(scaled, 3), nil
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

func paymentExternalReference(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}

	return pgtype.Text{String: *value, Valid: true}
}
