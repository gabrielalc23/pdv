package auth

import (
	"errors"
	"testing"
)

func validRegisterRequest() RegisterRequest {
	return RegisterRequest{Email: "owner@example.com", Password: "uma senha longa e segura", DisplayName: "Owner", Organization: OrganizationRequest{Name: "Empresa", Slug: "empresa", Timezone: "America/Sao_Paulo", Locale: "pt-BR", Currency: "BRL"}, Store: StoreRequest{Code: "MATRIZ", Name: "Matriz", Timezone: "America/Sao_Paulo"}, ClientID: "pdv-admin"}
}

func TestValidateRegister(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*RegisterRequest)
		field  string
	}{
		{"email", func(r *RegisterRequest) { r.Email = "invalid" }, "email"},
		{"display name", func(r *RegisterRequest) { r.DisplayName = "  " }, "displayName"},
		{"slug", func(r *RegisterRequest) { r.Organization.Slug = "Invalid Slug" }, "organization.slug"},
		{"timezone", func(r *RegisterRequest) { r.Organization.Timezone = "Mars/Olympus" }, "organization.timezone"},
		{"locale", func(r *RegisterRequest) { r.Organization.Locale = "invalid" }, "organization.locale"},
		{"currency", func(r *RegisterRequest) { r.Organization.Currency = "XXX" }, "organization.currency"},
		{"store code", func(r *RegisterRequest) { r.Store.Code = "invalid code" }, "store.code"},
		{"client", func(r *RegisterRequest) { r.ClientID = "unknown" }, "clientId"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := validRegisterRequest()
			test.mutate(&input)
			err := validateRegister(&input)
			var validation *ValidationError
			if !errors.As(err, &validation) || validation.Field != test.field {
				t.Fatalf("got %#v, want field %s", err, test.field)
			}
		})
	}
}

func TestValidateRegisterNormalizesAllowedFields(t *testing.T) {
	input := validRegisterRequest()
	input.Email = " Owner@Example.com "
	input.Organization.Slug = " Empresa "
	input.Store.Code = " matriz "
	if err := validateRegister(&input); err != nil {
		t.Fatal(err)
	}
	if input.Email != "Owner@Example.com" || input.Organization.Slug != "empresa" || input.Store.Code != "MATRIZ" {
		t.Fatalf("unexpected normalization: %+v", input)
	}
}
