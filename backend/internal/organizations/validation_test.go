package organizations

import (
	"errors"
	"testing"
)

func TestValidateCreateNormalizesInput(t *testing.T) {
	input := CreateOrganizationRequest{
		Organization: OrganizationInput{Name: " Loja Central ", Slug: " LOJA-CENTRAL ", Timezone: "America/Sao_Paulo", Locale: "pt-BR", Currency: "brl"},
		Store:        InitialStoreInput{Code: " matriz ", Name: " Matriz ", Timezone: "America/Sao_Paulo"},
	}
	if err := validateCreate(&input); err != nil {
		t.Fatalf("validateCreate() error = %v", err)
	}
	if input.Organization.Name != "Loja Central" || input.Organization.Slug != "loja-central" || input.Organization.Currency != "BRL" {
		t.Fatalf("organization was not normalized: %+v", input.Organization)
	}
	if input.Store.Code != "MATRIZ" || input.Store.Name != "Matriz" {
		t.Fatalf("store was not normalized: %+v", input.Store)
	}
}

func TestValidateUpdateRequiresAField(t *testing.T) {
	err := validateUpdate(&UpdateOrganizationRequest{})
	var validation *ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("validateUpdate() error = %v, want ValidationError", err)
	}
}

func TestValidateArchiveRequiresExplicitConfirmation(t *testing.T) {
	if err := validateArchive(ArchiveOrganizationRequest{}); err == nil {
		t.Fatal("validateArchive() error = nil, want confirmation error")
	}
	if err := validateArchive(ArchiveOrganizationRequest{Confirm: true}); err != nil {
		t.Fatalf("validateArchive(confirm=true) error = %v", err)
	}
}

func TestMapHTTPErrorSlugConflict(t *testing.T) {
	mapped := mapHTTPError(ErrOrganizationSlugInUse)
	if mapped.status != 409 || mapped.code != "ORGANIZATION_SLUG_ALREADY_IN_USE" {
		t.Fatalf("mapHTTPError() = %+v", mapped)
	}
}
