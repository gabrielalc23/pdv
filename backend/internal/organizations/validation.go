package organizations

import (
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	organizationSlugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	initialStoreCodePattern = regexp.MustCompile(`^[A-Z0-9][A-Z0-9_-]*$`)
	allowedLocales          = map[string]bool{"pt-BR": true, "en-US": true}
	allowedCurrencies       = map[string]bool{"BRL": true, "USD": true, "EUR": true}
)

func validateCreate(input *CreateOrganizationRequest) error {
	if input == nil {
		return ErrInvalidRequest
	}
	if err := normalizeOrganizationInput(&input.Organization, "organization."); err != nil {
		return err
	}

	input.Store.Code = strings.ToUpper(strings.TrimSpace(input.Store.Code))
	input.Store.Name = strings.TrimSpace(input.Store.Name)
	input.Store.Timezone = strings.TrimSpace(input.Store.Timezone)
	if len(input.Store.Code) > 50 || !initialStoreCodePattern.MatchString(input.Store.Code) {
		return validationError("store.code", "Código da loja inválido.")
	}
	if input.Store.Name == "" || utf8.RuneCountInString(input.Store.Name) > 150 {
		return validationError("store.name", "Nome da loja inválido.")
	}
	if len(input.Store.Timezone) > 64 {
		return validationError("store.timezone", "Timezone inválida.")
	}
	if _, err := time.LoadLocation(input.Store.Timezone); err != nil {
		return validationError("store.timezone", "Timezone inválida.")
	}
	return nil
}

func validateUpdate(input *UpdateOrganizationRequest) error {
	if input == nil {
		return ErrInvalidRequest
	}
	if input.Name == nil && input.Slug == nil && input.Timezone == nil && input.Locale == nil && input.Currency == nil {
		return validationError("", "Informe ao menos um campo para atualização.")
	}

	if input.Name != nil {
		value := strings.TrimSpace(*input.Name)
		if value == "" || utf8.RuneCountInString(value) > 150 {
			return validationError("name", "Nome da organização inválido.")
		}
		input.Name = &value
	}
	if input.Slug != nil {
		value := strings.ToLower(strings.TrimSpace(*input.Slug))
		if len(value) > 120 || !organizationSlugPattern.MatchString(value) {
			return validationError("slug", "Slug inválido.")
		}
		input.Slug = &value
	}
	if input.Timezone != nil {
		value := strings.TrimSpace(*input.Timezone)
		if len(value) > 64 {
			return validationError("timezone", "Timezone inválida.")
		}
		if _, err := time.LoadLocation(value); err != nil {
			return validationError("timezone", "Timezone inválida.")
		}
		input.Timezone = &value
	}
	if input.Locale != nil {
		value := strings.TrimSpace(*input.Locale)
		if !allowedLocales[value] {
			return validationError("locale", "Locale inválido.")
		}
		input.Locale = &value
	}
	if input.Currency != nil {
		value := strings.ToUpper(strings.TrimSpace(*input.Currency))
		if !allowedCurrencies[value] {
			return validationError("currency", "Moeda inválida.")
		}
		input.Currency = &value
	}
	return nil
}

func validateArchive(input ArchiveOrganizationRequest) error {
	if !input.Confirm {
		return validationError("confirm", "Confirmação explícita é obrigatória.")
	}
	return nil
}

func normalizeOrganizationInput(input *OrganizationInput, fieldPrefix string) error {
	input.Name = strings.TrimSpace(input.Name)
	input.Slug = strings.ToLower(strings.TrimSpace(input.Slug))
	input.Timezone = strings.TrimSpace(input.Timezone)
	input.Locale = strings.TrimSpace(input.Locale)
	input.Currency = strings.ToUpper(strings.TrimSpace(input.Currency))

	if input.Name == "" || utf8.RuneCountInString(input.Name) > 150 {
		return validationError(fieldPrefix+"name", "Nome da organização inválido.")
	}
	if len(input.Slug) > 120 || !organizationSlugPattern.MatchString(input.Slug) {
		return validationError(fieldPrefix+"slug", "Slug inválido.")
	}
	if len(input.Timezone) > 64 {
		return validationError(fieldPrefix+"timezone", "Timezone inválida.")
	}
	if _, err := time.LoadLocation(input.Timezone); err != nil {
		return validationError(fieldPrefix+"timezone", "Timezone inválida.")
	}
	if !allowedLocales[input.Locale] {
		return validationError(fieldPrefix+"locale", "Locale inválido.")
	}
	if !allowedCurrencies[input.Currency] {
		return validationError(fieldPrefix+"currency", "Moeda inválida.")
	}
	return nil
}
