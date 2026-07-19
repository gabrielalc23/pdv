package stores

import (
	"strings"
	"time"

	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/jackc/pgx/v5/pgtype"
)

func toStoreResponse(row database.Store) StoreResponse {
	var archivedAt *time.Time
	if row.ArchivedAt.Valid {
		value := row.ArchivedAt.Time.UTC()
		archivedAt = &value
	}
	return StoreResponse{
		ID: row.ID.String(), Code: row.Code, Name: row.Name, Status: string(row.Status), Timezone: row.Timezone,
		ArchivedAt: archivedAt, CreatedAt: timestamp(row.CreatedAt), UpdatedAt: timestamp(row.UpdatedAt),
	}
}

func toPaymentMethodResponse(row database.PaymentMethod) PaymentMethodResponse {
	var provider *string
	if row.Provider.Valid {
		value := row.Provider.String
		provider = &value
	}
	return PaymentMethodResponse{
		ID: row.ID.String(), Code: row.Code, Name: row.Name, Kind: string(row.Kind), Provider: provider,
		AllowsChange: row.AllowsChange, RequiresExternalReference: row.RequiresExternalReference,
		AllowsInstallments: row.AllowsInstallments, MaxInstallments: row.MaxInstallments,
		FeePercentage: numericString(row.FeePercentage), SettlementDays: row.SettlementDays,
		IsActive: row.IsActive, SortOrder: row.SortOrder,
		CreatedAt: timestamp(row.CreatedAt), UpdatedAt: timestamp(row.UpdatedAt),
	}
}

func toStorePaymentMethodResponse(row database.ListStorePaymentMethodsRow) StorePaymentMethodResponse {
	return StorePaymentMethodResponse{
		PaymentMethodID: row.PaymentMethodID.String(), Code: row.Code, Name: row.Name, Kind: string(row.Kind),
		IsActive: row.IsActive, OrganizationActive: row.OrganizationActive, SortOrder: row.SortOrder,
		CreatedAt: timestamp(row.CreatedAt), UpdatedAt: timestamp(row.UpdatedAt),
	}
}

func timestamp(value pgtype.Timestamptz) time.Time {
	if !value.Valid {
		return time.Time{}
	}
	return value.Time.UTC()
}

func numericString(value pgtype.Numeric) string {
	if !value.Valid || value.Int == nil {
		return "0"
	}
	digits := value.Int.String()
	negative := strings.HasPrefix(digits, "-")
	if negative {
		digits = strings.TrimPrefix(digits, "-")
	}
	if value.Exp >= 0 {
		digits += strings.Repeat("0", int(value.Exp))
	} else {
		scale := int(-value.Exp)
		if len(digits) <= scale {
			digits = strings.Repeat("0", scale-len(digits)+1) + digits
		}
		cut := len(digits) - scale
		digits = digits[:cut] + "." + digits[cut:]
		digits = strings.TrimRight(strings.TrimRight(digits, "0"), ".")
	}
	if negative && digits != "0" {
		return "-" + digits
	}
	return digits
}
