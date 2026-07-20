package stores

import (
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	codePattern    = regexp.MustCompile(`^[A-Z0-9][A-Z0-9_-]*$`)
	decimalPattern = regexp.MustCompile(`^(?:0|[1-9][0-9]{0,2})(?:\.[0-9]{1,4})?$`)
)

type normalizedStoreInput struct {
	Code     string
	Name     string
	Timezone string
}

type normalizedPaymentMethodInput struct {
	Code                      string
	Name                      string
	Kind                      database.PaymentMethodKind
	Provider                  pgtype.Text
	AllowsChange              bool
	RequiresExternalReference bool
	AllowsInstallments        bool
	MaxInstallments           int16
	FeePercentage             pgtype.Numeric
	SettlementDays            int32
	IsActive                  bool
	SortOrder                 int32
}

func normalizeStoreInput(code, name, timezone string) (normalizedStoreInput, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	name = strings.TrimSpace(name)
	timezone = strings.TrimSpace(timezone)
	if len(code) > 50 || !codePattern.MatchString(code) {
		return normalizedStoreInput{}, validationError("code", "must contain 1 to 50 uppercase letters, numbers, underscores, or hyphens")
	}
	if name == "" || utf8.RuneCountInString(name) > 150 {
		return normalizedStoreInput{}, validationError("name", "must contain 1 to 150 characters")
	}
	if len(timezone) > 64 {
		return normalizedStoreInput{}, validationError("timezone", "must be a valid IANA timezone")
	}
	if _, err := time.LoadLocation(timezone); err != nil {
		return normalizedStoreInput{}, validationError("timezone", "must be a valid IANA timezone")
	}
	return normalizedStoreInput{Code: code, Name: name, Timezone: timezone}, nil
}

func normalizePaymentMethodInput(input UpsertPaymentMethodInput, creating bool) (normalizedPaymentMethodInput, error) {
	code := strings.ToUpper(strings.TrimSpace(input.Code))
	name := strings.TrimSpace(input.Name)
	if len(code) > 50 || !codePattern.MatchString(code) {
		return normalizedPaymentMethodInput{}, validationError("code", "must contain 1 to 50 uppercase letters, numbers, underscores, or hyphens")
	}
	if name == "" || utf8.RuneCountInString(name) > 100 {
		return normalizedPaymentMethodInput{}, validationError("name", "must contain 1 to 100 characters")
	}

	kind := database.PaymentMethodKind(strings.ToUpper(strings.TrimSpace(input.Kind)))
	if !kind.Valid() {
		return normalizedPaymentMethodInput{}, validationError("kind", "must be a valid payment method kind")
	}
	provider := pgtype.Text{}
	if input.Provider != nil {
		value := strings.TrimSpace(*input.Provider)
		if value == "" || utf8.RuneCountInString(value) > 100 {
			return normalizedPaymentMethodInput{}, validationError("provider", "must contain 1 to 100 characters when provided")
		}
		provider = pgtype.Text{String: value, Valid: true}
	}
	if input.AllowsChange && kind != database.PaymentMethodKindCASH {
		return normalizedPaymentMethodInput{}, validationError("allowsChange", "can only be enabled for CASH")
	}
	if input.RequiresExternalReference && !provider.Valid {
		return normalizedPaymentMethodInput{}, validationError("provider", "is required when an external reference is required")
	}
	if input.AllowsInstallments {
		if kind != database.PaymentMethodKindCREDITCARD || input.MaxInstallments <= 1 {
			return normalizedPaymentMethodInput{}, validationError("maxInstallments", "must be greater than one for installment credit cards")
		}
	} else if input.MaxInstallments != 1 {
		return normalizedPaymentMethodInput{}, validationError("maxInstallments", "must be one when installments are disabled")
	}

	fee := strings.TrimSpace(input.FeePercentage)
	if fee == "" {
		fee = "0"
	}
	if !decimalPattern.MatchString(fee) {
		return normalizedPaymentMethodInput{}, validationError("feePercentage", "must be between 0 and 999.9999 with at most four decimal places")
	}
	var feeNumeric pgtype.Numeric
	if err := feeNumeric.Scan(fee); err != nil {
		return normalizedPaymentMethodInput{}, validationError("feePercentage", "must be a valid decimal")
	}
	if input.SettlementDays < 0 {
		return normalizedPaymentMethodInput{}, validationError("settlementDays", "must not be negative")
	}
	if input.SortOrder < 0 {
		return normalizedPaymentMethodInput{}, validationError("sortOrder", "must not be negative")
	}
	isActive := true
	if input.IsActive != nil {
		isActive = *input.IsActive
	} else if !creating {
		isActive = false
	}

	return normalizedPaymentMethodInput{
		Code: code, Name: name, Kind: kind, Provider: provider,
		AllowsChange: input.AllowsChange, RequiresExternalReference: input.RequiresExternalReference,
		AllowsInstallments: input.AllowsInstallments, MaxInstallments: input.MaxInstallments,
		FeePercentage: feeNumeric, SettlementDays: input.SettlementDays,
		IsActive: isActive, SortOrder: input.SortOrder,
	}, nil
}

func parseUUID(raw, field string) (pgtype.UUID, error) {
	var id pgtype.UUID
	if err := id.Scan(strings.TrimSpace(raw)); err != nil || !id.Valid {
		return pgtype.UUID{}, validationError(field, "must be a valid UUID")
	}
	return id, nil
}

func normalizePagination(page, pageSize *int) (int, int, error) {
	resolvedPage, resolvedPageSize := 1, 20
	if page != nil {
		if *page < 1 {
			return 0, 0, validationError("page", "must be greater than zero")
		}
		resolvedPage = *page
	}
	if pageSize != nil {
		if *pageSize < 1 || *pageSize > 100 {
			return 0, 0, validationError("pageSize", "must be between 1 and 100")
		}
		resolvedPageSize = *pageSize
	}
	return resolvedPage, resolvedPageSize, nil
}
