package invitations

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gabrielalc23/pdv/internal/platform/database"
)

func TestValidateCreateNormalizesEmailAndRequiresAssignments(t *testing.T) {
	input := CreateInput{Email: "  User@Example.com  ", Assignments: []AssignmentInput{{RoleID: "018f2f9a-8d4b-7f35-8b31-84b75f216456"}}}
	if err := validateCreate(&input); err != nil {
		t.Fatal(err)
	}
	if input.Email != "User@Example.com" || normalizeEmail(input.Email) != "user@example.com" {
		t.Fatalf("unexpected normalization: %q", input.Email)
	}

	input.Assignments = nil
	var validation *ValidationError
	if err := validateCreate(&input); !errors.As(err, &validation) || validation.Field != "assignments" {
		t.Fatalf("expected assignments validation error, got %v", err)
	}
}

func TestValidateRoleStoreCompatibility(t *testing.T) {
	storeID := testUUID(t, "018f2f9a-8d4b-7f35-8b31-84b75f216456")
	tests := []struct {
		name    string
		scope   database.RoleAssignmentScope
		storeID pgtype.UUID
		wantErr bool
	}{
		{name: "organization without store", scope: database.RoleAssignmentScopeORGANIZATION},
		{name: "organization with store", scope: database.RoleAssignmentScopeORGANIZATION, storeID: storeID, wantErr: true},
		{name: "store with store", scope: database.RoleAssignmentScopeSTORE, storeID: storeID},
		{name: "store without store", scope: database.RoleAssignmentScopeSTORE, wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateRoleStore(database.Role{AssignmentScope: test.scope}, test.storeID, "storeId")
			if (err != nil) != test.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}

func TestValidateAnonymousAcceptRequiresCompleteCredentials(t *testing.T) {
	valid := AcceptInput{DisplayName: "  Maria Silva ", Password: "a password validated by policy later", ClientID: "pdv-admin"}
	if err := validateAnonymousAccept(&valid); err != nil {
		t.Fatal(err)
	}
	if valid.DisplayName != "Maria Silva" {
		t.Fatalf("display name = %q", valid.DisplayName)
	}

	for _, input := range []AcceptInput{
		{Password: "password", ClientID: "pdv-admin"},
		{DisplayName: "Maria", ClientID: "pdv-admin"},
		{DisplayName: "Maria", Password: "password", ClientID: "unknown"},
	} {
		if err := validateAnonymousAccept(&input); err == nil {
			t.Fatalf("expected validation failure for %#v", input)
		}
	}
}
