package invitations

import (
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5/pgtype"

	usersmodule "github.com/gabrielalc23/pdv/internal/users"
)

func normalizeEmail(value string) string { return strings.ToLower(strings.TrimSpace(value)) }

func validateCreate(input *CreateInput) error {
	input.Email = strings.TrimSpace(input.Email)
	if input.Email == "" || len(input.Email) > 320 {
		return validationError("email", "E-mail inválido.")
	}
	address, err := mail.ParseAddress(input.Email)
	if err != nil || address.Address != input.Email || !strings.Contains(input.Email, "@") {
		return validationError("email", "E-mail inválido.")
	}
	if len(input.Assignments) == 0 {
		return validationError("assignments", "Pelo menos uma atribuição é obrigatória.")
	}
	if len(input.Assignments) > 50 {
		return validationError("assignments", "Muitas atribuições.")
	}
	return nil
}

func validateAnonymousAccept(input *AcceptInput) error {
	displayName, err := usersmodule.NormalizeDisplayName(input.DisplayName)
	if err != nil {
		return validationError("displayName", "Nome de exibição inválido.")
	}
	input.DisplayName = displayName
	if input.Password == "" {
		return validationError("password", "Senha obrigatória.")
	}
	if input.ClientID != "pdv-web" && input.ClientID != "pdv-admin" {
		return validationError("clientId", "Cliente inválido.")
	}
	input.DeviceName = strings.TrimSpace(input.DeviceName)
	if utf8.RuneCountInString(input.DeviceName) > 150 {
		return validationError("deviceName", "Nome do dispositivo inválido.")
	}
	return nil
}

func parseUUID(field, value string) (pgtype.UUID, error) {
	var id pgtype.UUID
	value = strings.TrimSpace(value)
	if id.Scan(value) != nil || !validUUID(id) || id.String() != strings.ToLower(value) {
		return pgtype.UUID{}, validationError(field, "UUID inválido.")
	}
	return id, nil
}

func optionalUUID(field string, value *string) (pgtype.UUID, error) {
	if value == nil {
		return pgtype.UUID{}, nil
	}
	return parseUUID(field, *value)
}

func validUUID(id pgtype.UUID) bool {
	if !id.Valid {
		return false
	}
	for _, b := range id.Bytes {
		if b != 0 {
			return true
		}
	}
	return false
}

func parseOptionalTime(field, value string) (pgtype.Timestamptz, error) {
	if strings.TrimSpace(value) == "" {
		return pgtype.Timestamptz{}, nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return pgtype.Timestamptz{}, validationError(field, "Data inválida.")
	}
	return pgtype.Timestamptz{Time: parsed, Valid: true}, nil
}
