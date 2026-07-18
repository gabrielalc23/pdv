package authz_test

import (
	"sort"
	"strings"
	"testing"

	"github.com/gabrielalc23/pdv/internal/platform/authcontext"
	"github.com/gabrielalc23/pdv/internal/platform/authz"
)

func TestAllScopesComplete(t *testing.T) {
	scopes := authz.AllScopes()
	if len(scopes) == 0 {
		t.Fatal("expected at least one scope in catalog")
	}

	// Check for duplicates
	seen := make(map[authcontext.Scope]struct{}, len(scopes))
	for _, s := range scopes {
		if _, ok := seen[s]; ok {
			t.Fatalf("duplicate scope: %s", s)
		}
		seen[s] = struct{}{}
	}
}

func TestAllScopesDeterministic(t *testing.T) {
	first := authz.AllScopes()
	second := authz.AllScopes()

	if len(first) != len(second) {
		t.Fatal("catalog length changed between calls")
	}
	for i := range first {
		if first[i] != second[i] {
			t.Fatalf("catalog not deterministic at index %d: %s vs %s", i, first[i], second[i])
		}
	}
}

func TestNoWildcardScopes(t *testing.T) {
	scopes := authz.AllScopes()
	for _, s := range scopes {
		if strings.Contains(string(s), "*") {
			t.Fatalf("scope %q contains wildcard", s)
		}
	}
}

func TestNoEmptyScopes(t *testing.T) {
	scopes := authz.AllScopes()
	for _, s := range scopes {
		if strings.TrimSpace(string(s)) == "" {
			t.Fatal("found empty scope in catalog")
		}
	}
}

func TestScopeFormat(t *testing.T) {
	scopes := authz.AllScopes()
	for _, s := range scopes {
		parts := strings.Split(string(s), ".")
		if len(parts) < 2 {
			t.Fatalf("scope %q does not follow resource.action format", s)
		}
		for _, p := range parts {
			if p == "" {
				t.Fatalf("scope %q has empty segment", s)
			}
		}
	}
}

func TestScopeSet(t *testing.T) {
	tests := []struct {
		name  string
		set   []string
		check string
		has   bool
	}{
		{"has exact", []string{"read", "write"}, "read", true},
		{"does not have", []string{"read"}, "write", false},
		{"empty set", []string{}, "anything", false},
		{"nil set", nil, "anything", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ss authcontext.ScopeSet
			if tt.set != nil {
				ss = authcontext.NewScopeSet(asScopes(tt.set)...)
			}
			if ss.Has(authcontext.Scope(tt.check)) != tt.has {
				t.Fatalf("expected Has=%v", tt.has)
			}
		})
	}
}

func TestScopeSetClone(t *testing.T) {
	original := authcontext.NewScopeSet("a", "b")
	cloned := original.Clone()
	cloned["c"] = struct{}{}
	if len(original) != 2 {
		t.Fatal("clone was not independent")
	}
}

func TestScopeSetSorted(t *testing.T) {
	ss := authcontext.NewScopeSet("z", "a", "m")
	sorted := ss.Sorted()
	if !sort.SliceIsSorted(sorted, func(i, j int) bool { return sorted[i] < sorted[j] }) {
		t.Fatal("sorted scopes are not sorted")
	}
	if len(sorted) != 3 {
		t.Fatal("expected 3 elements")
	}
}

func TestScopeSetNil(t *testing.T) {
	var nilSet authcontext.ScopeSet
	if nilSet.Has("anything") {
		t.Fatal("nil set should not have anything")
	}
	if nilSet.HasAll("a", "b") {
		t.Fatal("nil set should not have all")
	}
	if nilSet.HasAny("a") {
		t.Fatal("nil set should not have any")
	}
	if len(nilSet.Sorted()) != 0 {
		t.Fatal("nil set sorted should be empty")
	}
	if nilSet.Clone() != nil {
		t.Fatal("nil set clone should be nil")
	}
}

func TestScopePrefixNotMatching(t *testing.T) {
	ss := authcontext.NewScopeSet("inventory.read")
	if ss.Has("inventory") {
		t.Fatal("prefix should not match")
	}
	if ss.Has("inventory.") {
		t.Fatal("prefix with dot should not match")
	}
	if ss.Has("inventory.*") {
		t.Fatal("wildcard should not match")
	}
}
