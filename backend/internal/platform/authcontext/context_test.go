package authcontext_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gabrielalc23/pdv/internal/platform/authcontext"
)

func uuid(s string) pgtype.UUID {
	var id pgtype.UUID
	if err := id.Scan(s); err != nil {
		panic(err)
	}
	return id
}

func validPrincipal(t *testing.T, kind authcontext.ContextKind) authcontext.Principal {
	t.Helper()
	p := authcontext.Principal{
		UserID:          uuid("550e8400-e29b-41d4-a716-446655440000"),
		SessionID:       uuid("660e8400-e29b-41d4-a716-446655440001"),
		ClientID:        "test-client",
		ContextKind:     kind,
		RoleKeys:        []string{"admin"},
		Scopes:          authcontext.NewScopeSet("test.scope"),
		PasswordVersion: 1,
		AuthTime:        time.Now(),
		TokenID:         uuid("770e8400-e29b-41d4-a716-446655440002"),
	}

	switch kind {
	case authcontext.ContextOrganization:
		oav := int64(1)
		mav := int64(1)
		p.OrganizationID = uuid("880e8400-e29b-41d4-a716-446655440010")
		p.MembershipID = uuid("990e8400-e29b-41d4-a716-446655440020")
		p.OrgAuthzVersion = &oav
		p.MemberAuthzVersion = &mav
	case authcontext.ContextStore:
		oav := int64(1)
		mav := int64(1)
		p.OrganizationID = uuid("880e8400-e29b-41d4-a716-446655440010")
		p.MembershipID = uuid("990e8400-e29b-41d4-a716-446655440020")
		p.StoreID = uuid("aa0e8400-e29b-41d4-a716-446655440030")
		p.OrgAuthzVersion = &oav
		p.MemberAuthzVersion = &mav
	}
	return p
}

func TestPrincipalIdentityValid(t *testing.T) {
	p := validPrincipal(t, authcontext.ContextIdentity)
	if err := p.Validate(); err != nil {
		t.Fatalf("expected valid, got %v", err)
	}
	if !p.IsIdentity() {
		t.Fatal("expected IsIdentity")
	}
	if p.HasOrganizationScope() {
		t.Fatal("expected no organization scope")
	}
	if p.HasStoreScope() {
		t.Fatal("expected no store scope")
	}
}

func TestPrincipalOrganizationValid(t *testing.T) {
	p := validPrincipal(t, authcontext.ContextOrganization)
	if err := p.Validate(); err != nil {
		t.Fatalf("expected valid, got %v", err)
	}
	if !p.IsOrganization() {
		t.Fatal("expected IsOrganization")
	}
	if !p.HasOrganizationScope() {
		t.Fatal("expected organization scope")
	}
}

func TestPrincipalStoreValid(t *testing.T) {
	p := validPrincipal(t, authcontext.ContextStore)
	if err := p.Validate(); err != nil {
		t.Fatalf("expected valid, got %v", err)
	}
	if !p.IsStore() {
		t.Fatal("expected IsStore")
	}
	if !p.HasStoreScope() {
		t.Fatal("expected store scope")
	}
}

func TestIdentityRejectsTenant(t *testing.T) {
	p := validPrincipal(t, authcontext.ContextIdentity)
	p.OrganizationID = uuid("880e8400-e29b-41d4-a716-446655440010")
	if err := p.Validate(); err == nil {
		t.Fatal("expected error for identity with organization")
	}
}

func TestOrganizationRequiresMembership(t *testing.T) {
	p := validPrincipal(t, authcontext.ContextOrganization)
	p.MembershipID = pgtype.UUID{}
	if err := p.Validate(); err == nil {
		t.Fatal("expected error for organization without membership")
	}
}

func TestOrganizationRejectsStore(t *testing.T) {
	p := validPrincipal(t, authcontext.ContextOrganization)
	p.StoreID = uuid("aa0e8400-e29b-41d4-a716-446655440030")
	if err := p.Validate(); err == nil {
		t.Fatal("expected error for organization with store")
	}
}

func TestStoreRequiresStoreID(t *testing.T) {
	p := validPrincipal(t, authcontext.ContextStore)
	p.StoreID = pgtype.UUID{}
	if err := p.Validate(); err == nil {
		t.Fatal("expected error for store without store_id")
	}
}

func TestTenantRequiresAuthVersions(t *testing.T) {
	p := validPrincipal(t, authcontext.ContextOrganization)
	p.OrgAuthzVersion = nil
	if err := p.Validate(); err == nil {
		t.Fatal("expected error for missing org_authz_version")
	}
}

func TestIdentityRejectsAuthVersions(t *testing.T) {
	p := validPrincipal(t, authcontext.ContextIdentity)
	oav := int64(1)
	p.OrgAuthzVersion = &oav
	if err := p.Validate(); err == nil {
		t.Fatal("expected error for identity with org_authz_version")
	}
}

func TestRoleSlicesAreReferences(t *testing.T) {
	roles := []string{"admin", "owner"}
	p := validPrincipal(t, authcontext.ContextIdentity)
	p.RoleKeys = roles
	roles[0] = "modified"
	if p.RoleKeys[0] != "modified" {
		t.Fatal("direct struct assignment shares reference; buildPrincipal must copy")
	}
}

func TestScopeSetIsReference(t *testing.T) {
	scopes := authcontext.NewScopeSet("a", "b")
	p := validPrincipal(t, authcontext.ContextIdentity)
	p.Scopes = scopes
	scopes["c"] = struct{}{}
	if !p.Scopes.Has("c") {
		t.Fatal("direct struct assignment shares reference; buildPrincipal must clone")
	}
}

func TestPrincipalFromContext(t *testing.T) {
	ctx := t.Context()
	p := validPrincipal(t, authcontext.ContextIdentity)
	ctx = authcontext.SetPrincipal(ctx, p)

	retrieved, ok := authcontext.PrincipalFromContext(ctx)
	if !ok {
		t.Fatal("expected principal in context")
	}
	if retrieved.UserID != p.UserID {
		t.Fatal("principal mismatch")
	}
}

func TestContextWithoutPrincipal(t *testing.T) {
	ctx := t.Context()
	_, ok := authcontext.PrincipalFromContext(ctx)
	if ok {
		t.Fatal("expected no principal")
	}
}

func TestMustPrincipal(t *testing.T) {
	ctx := t.Context()
	_, err := authcontext.MustPrincipal(ctx)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMustPrincipalValid(t *testing.T) {
	ctx := t.Context()
	p := validPrincipal(t, authcontext.ContextIdentity)
	ctx = authcontext.SetPrincipal(ctx, p)

	retrieved, err := authcontext.MustPrincipal(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if retrieved.UserID != p.UserID {
		t.Fatal("principal mismatch")
	}
}

func TestContextWrongType(t *testing.T) {
	ctx := context.WithValue(t.Context(), authcontext.PrincipalKey, "not-a-principal")
	_, ok := authcontext.PrincipalFromContext(ctx)
	if ok {
		t.Fatal("expected no principal for wrong type")
	}
}
