package auth

import (
	"testing"

	"github.com/gabrielalc23/pdv/internal/platform/authz"
)

func TestDefaultRoleTemplates(t *testing.T) {
	want := map[string]bool{"catalog_manager": true, "auditor": true, "store_manager": true, "cashier": true, "inventory_operator": true}
	catalog := map[string]bool{}
	for _, scope := range authz.AllScopes() {
		catalog[string(scope)] = true
	}
	for _, role := range defaultRoleTemplates {
		if !want[role.key] {
			t.Fatalf("unexpected role %q", role.key)
		}
		delete(want, role.key)
		seen := map[string]bool{}
		for _, scope := range role.scopes {
			if !catalog[scope] {
				t.Fatalf("role %s references unknown scope %s", role.key, scope)
			}
			if seen[scope] {
				t.Fatalf("role %s duplicates scope %s", role.key, scope)
			}
			seen[scope] = true
		}
	}
	if len(want) != 0 {
		t.Fatalf("missing roles: %v", want)
	}
}

func TestDefaultPaymentMethods(t *testing.T) {
	want := []string{"CASH", "PIX", "DEBIT", "CREDIT", "VOUCHER"}
	if len(defaultPayments) != len(want) {
		t.Fatalf("got %d methods", len(defaultPayments))
	}
	for i, code := range want {
		if defaultPayments[i].code != code {
			t.Fatalf("method %d = %s, want %s", i, defaultPayments[i].code, code)
		}
	}
}
