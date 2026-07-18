package authz_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gabrielalc23/pdv/internal/platform/authcontext"
	"github.com/gabrielalc23/pdv/internal/platform/authz"
)

func uuid(s string) pgtype.UUID {
	var id pgtype.UUID
	if err := id.Scan(s); err != nil {
		panic(err)
	}
	return id
}

func withPrincipal(ctx context.Context, kind authcontext.ContextKind) context.Context {
	p := authcontext.Principal{
		UserID:          uuid("550e8400-e29b-41d4-a716-446655440000"),
		SessionID:       uuid("660e8400-e29b-41d4-a716-446655440001"),
		ClientID:        "test-client",
		ContextKind:     kind,
		RoleKeys:        []string{},
		Scopes:          authcontext.NewScopeSet("read", "write"),
		PasswordVersion: 1,
		AuthTime:        time.Now(),
		TokenID:         uuid("770e8400-e29b-41d4-a716-446655440002"),
	}
	switch kind {
	case authcontext.ContextOrganization:
		oav, mav := int64(1), int64(1)
		p.OrganizationID = uuid("880e8400-e29b-41d4-a716-446655440010")
		p.MembershipID = uuid("990e8400-e29b-41d4-a716-446655440020")
		p.OrgAuthzVersion = &oav
		p.MemberAuthzVersion = &mav
	case authcontext.ContextStore:
		oav, mav := int64(1), int64(1)
		p.OrganizationID = uuid("880e8400-e29b-41d4-a716-446655440010")
		p.MembershipID = uuid("990e8400-e29b-41d4-a716-446655440020")
		p.StoreID = uuid("aa0e8400-e29b-41d4-a716-446655440030")
		p.OrgAuthzVersion = &oav
		p.MemberAuthzVersion = &mav
	}
	return authcontext.SetPrincipal(ctx, p)
}

func okHandler(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}
}

func TestRequireIdentityWithPrincipal(t *testing.T) {
	g := authz.NewGuard()
	handler := g.RequireIdentity()(okHandler(t))

	req := httptest.NewRequest("GET", "/", nil)
	req = req.WithContext(withPrincipal(req.Context(), authcontext.ContextIdentity))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRequireIdentityWithOrganization(t *testing.T) {
	g := authz.NewGuard()
	handler := g.RequireIdentity()(okHandler(t))

	req := httptest.NewRequest("GET", "/", nil)
	req = req.WithContext(withPrincipal(req.Context(), authcontext.ContextOrganization))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRequireIdentityWithStore(t *testing.T) {
	g := authz.NewGuard()
	handler := g.RequireIdentity()(okHandler(t))

	req := httptest.NewRequest("GET", "/", nil)
	req = req.WithContext(withPrincipal(req.Context(), authcontext.ContextStore))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRequireIdentityNoPrincipal(t *testing.T) {
	g := authz.NewGuard()
	handler := g.RequireIdentity()(okHandler(t))

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestRequireOrganizationContext(t *testing.T) {
	tests := []struct {
		name       string
		kind       authcontext.ContextKind
		wantStatus int
	}{
		{"identity rejected", authcontext.ContextIdentity, 400},
		{"organization accepted", authcontext.ContextOrganization, 200},
		{"store accepted", authcontext.ContextStore, 200},
	}

	g := authz.NewGuard()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := g.RequireOrganizationContext()(okHandler(t))
			req := httptest.NewRequest("GET", "/", nil)
			req = req.WithContext(withPrincipal(req.Context(), tt.kind))
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("expected %d, got %d", tt.wantStatus, rec.Code)
			}
		})
	}
}

func TestRequireStoreContext(t *testing.T) {
	tests := []struct {
		name       string
		kind       authcontext.ContextKind
		wantStatus int
	}{
		{"identity rejected", authcontext.ContextIdentity, 400},
		{"organization rejected", authcontext.ContextOrganization, 400},
		{"store accepted", authcontext.ContextStore, 200},
	}

	g := authz.NewGuard()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := g.RequireStoreContext()(okHandler(t))
			req := httptest.NewRequest("GET", "/", nil)
			req = req.WithContext(withPrincipal(req.Context(), tt.kind))
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("expected %d, got %d", tt.wantStatus, rec.Code)
			}
		})
	}
}

func TestGuardNoPrincipal(t *testing.T) {
	tests := []struct {
		name  string
		setup func(g authz.Guard) func(http.Handler) http.Handler
		want  int
	}{
		{"RequireIdentity", func(g authz.Guard) func(http.Handler) http.Handler { return g.RequireIdentity() }, 401},
		{"RequireOrganizationContext", func(g authz.Guard) func(http.Handler) http.Handler { return g.RequireOrganizationContext() }, 401},
		{"RequireStoreContext", func(g authz.Guard) func(http.Handler) http.Handler { return g.RequireStoreContext() }, 401},
		{"RequireAll", func(g authz.Guard) func(http.Handler) http.Handler { return g.RequireAll("read") }, 401},
		{"RequireAny", func(g authz.Guard) func(http.Handler) http.Handler { return g.RequireAny("read") }, 401},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := authz.NewGuard()
			handler := tt.setup(g)(okHandler(t))
			req := httptest.NewRequest("GET", "/", nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != tt.want {
				t.Fatalf("expected %d, got %d", tt.want, rec.Code)
			}
		})
	}
}

func TestRequireAll(t *testing.T) {
	tests := []struct {
		name       string
		scopeSet   []string
		required   []string
		wantStatus int
	}{
		{"has all", []string{"read", "write"}, []string{"read", "write"}, 200},
		{"missing one", []string{"read"}, []string{"read", "write"}, 403},
		{"missing all", []string{}, []string{"read"}, 403},
		{"duplicate in config", []string{"read", "read"}, []string{"read"}, 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := authz.NewGuard()
			handler := g.RequireAll(asScopes(tt.required)...)(okHandler(t))

			ctx := t.Context()
			ctx = withPrincipal(ctx, authcontext.ContextIdentity)
			p, _ := authcontext.MustPrincipal(ctx)
			p.Scopes = authcontext.NewScopeSet(asScopes(tt.scopeSet)...)
			ctx = authcontext.SetPrincipal(ctx, p)

			req := httptest.NewRequest("GET", "/", nil)
			req = req.WithContext(ctx)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("%s: expected %d, got %d", tt.name, tt.wantStatus, rec.Code)
			}
		})
	}
}

func TestRequireAny(t *testing.T) {
	tests := []struct {
		name       string
		scopeSet   []string
		required   []string
		wantStatus int
	}{
		{"has first", []string{"read", "write"}, []string{"read", "delete"}, 200},
		{"has last", []string{"read", "write"}, []string{"delete", "write"}, 200},
		{"none", []string{"read"}, []string{"delete", "update"}, 403},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := authz.NewGuard()
			handler := g.RequireAny(asScopes(tt.required)...)(okHandler(t))

			ctx := t.Context()
			ctx = withPrincipal(ctx, authcontext.ContextIdentity)
			p, _ := authcontext.MustPrincipal(ctx)
			p.Scopes = authcontext.NewScopeSet(asScopes(tt.scopeSet)...)
			ctx = authcontext.SetPrincipal(ctx, p)

			req := httptest.NewRequest("GET", "/", nil)
			req = req.WithContext(ctx)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("%s: expected %d, got %d", tt.name, tt.wantStatus, rec.Code)
			}
		})
	}
}

func TestRequireAllEmptyPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for empty scopes")
		}
	}()
	g := authz.NewGuard()
	_ = g.RequireAll()
}

func TestRequireAnyEmptyPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for empty scopes")
		}
	}()
	g := authz.NewGuard()
	_ = g.RequireAny()
}

func TestGuardSetsCorrectHeaders(t *testing.T) {
	g := authz.NewGuard()
	handler := g.RequireAll("read")(okHandler(t))

	req := httptest.NewRequest("GET", "/", nil)
	req = req.WithContext(withPrincipal(req.Context(), authcontext.ContextIdentity))
	p, _ := authcontext.MustPrincipal(req.Context())
	p.Scopes = authcontext.NewScopeSet()
	req = req.WithContext(authcontext.SetPrincipal(req.Context(), p))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != 403 {
		t.Fatalf("expected 403, got %d", rec.Code)
	}

	cc := rec.Header().Get("Cache-Control")
	if cc != "no-store" {
		t.Fatalf("expected Cache-Control: no-store, got %q", cc)
	}
}

func asScopes(s []string) []authcontext.Scope {
	scopes := make([]authcontext.Scope, len(s))
	for i, v := range s {
		scopes[i] = authcontext.Scope(v)
	}
	return scopes
}
