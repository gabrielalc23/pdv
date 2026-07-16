package receipt

import (
	"math/big"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
)

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
