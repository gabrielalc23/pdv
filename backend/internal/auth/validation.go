package auth

import (
	"net/mail"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5/pgtype"

	usersmodule "github.com/gabrielalc23/pdv/internal/users"
)

var (
	slugPattern      = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	storeCodePattern = regexp.MustCompile(`^[A-Z0-9][A-Z0-9_-]*$`)
	locales          = map[string]bool{"pt-BR": true, "en-US": true}
	currencies       = map[string]bool{"BRL": true, "USD": true, "EUR": true}
	clients          = map[string]bool{"pdv-web": true, "pdv-admin": true}
)

func normalizeEmail(value string) string { return strings.ToLower(strings.TrimSpace(value)) }

func validateEmail(value string) error {
	if value == "" || len(value) > 320 {
		return validationError("email", "E-mail inválido.")
	}
	address, err := mail.ParseAddress(value)
	if err != nil || address.Address != value || !strings.Contains(value, "@") {
		return validationError("email", "E-mail inválido.")
	}
	return nil
}

func validateClientID(value string) error {
	if !clients[value] {
		return validationError("clientId", "Cliente inválido.")
	}
	return nil
}

func validateRegister(in *RegisterRequest) error {
	in.Email = strings.TrimSpace(in.Email)
	if err := validateEmail(in.Email); err != nil {
		return err
	}
	displayName, err := usersmodule.NormalizeDisplayName(in.DisplayName)
	if err != nil {
		return validationError("displayName", "Nome de exibição inválido.")
	}
	in.DisplayName = displayName
	in.Organization.Name = strings.TrimSpace(in.Organization.Name)
	in.Organization.Slug = strings.ToLower(strings.TrimSpace(in.Organization.Slug))
	if in.Organization.Name == "" || utf8.RuneCountInString(in.Organization.Name) > 150 {
		return validationError("organization.name", "Nome da organização inválido.")
	}
	if !slugPattern.MatchString(in.Organization.Slug) || len(in.Organization.Slug) > 120 {
		return validationError("organization.slug", "Slug inválido.")
	}
	if _, err := time.LoadLocation(in.Organization.Timezone); err != nil {
		return validationError("organization.timezone", "Timezone inválida.")
	}
	if !locales[in.Organization.Locale] {
		return validationError("organization.locale", "Locale inválido.")
	}
	in.Organization.Currency = strings.ToUpper(in.Organization.Currency)
	if !currencies[in.Organization.Currency] {
		return validationError("organization.currency", "Moeda inválida.")
	}
	in.Store.Code = strings.ToUpper(strings.TrimSpace(in.Store.Code))
	in.Store.Name = strings.TrimSpace(in.Store.Name)
	if len(in.Store.Code) > 50 || !storeCodePattern.MatchString(in.Store.Code) {
		return validationError("store.code", "Código da loja inválido.")
	}
	if in.Store.Name == "" || utf8.RuneCountInString(in.Store.Name) > 150 {
		return validationError("store.name", "Nome da loja inválido.")
	}
	if _, err := time.LoadLocation(in.Store.Timezone); err != nil {
		return validationError("store.timezone", "Timezone inválida.")
	}
	if err := validateClientID(in.ClientID); err != nil {
		return err
	}
	in.DeviceName = strings.TrimSpace(in.DeviceName)
	if utf8.RuneCountInString(in.DeviceName) > 150 {
		return validationError("deviceName", "Nome do dispositivo inválido.")
	}
	return nil
}

func parseOptionalUUID(value *string, field string) (pgtype.UUID, error) {
	if value == nil {
		return pgtype.UUID{}, nil
	}
	if *value == "" {
		return pgtype.UUID{}, validationError(field, "UUID inválido.")
	}
	var id pgtype.UUID
	if err := id.Scan(*value); err != nil || !id.Valid {
		return pgtype.UUID{}, validationError(field, "UUID inválido.")
	}
	return id, nil
}
