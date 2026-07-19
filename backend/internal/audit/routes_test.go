package audit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gabrielalc23/pdv/internal/platform/authcontext"
	"github.com/gabrielalc23/pdv/internal/platform/authz"
	"github.com/gabrielalc23/pdv/internal/platform/database"
)

func TestRoutesRequireOrganizationContextAndAuditReadScope(t *testing.T) {
	router := chi.NewRouter()
	RegisterRoutes(router, NewHandler(nil), authz.NewGuard())

	tests := []struct {
		name      string
		principal *authcontext.Principal
		status    int
	}{
		{name: "authentication", status: http.StatusUnauthorized},
		{name: "organization context", principal: routeIdentityPrincipal(), status: http.StatusBadRequest},
		{name: "audit scope", principal: routeOrganizationPrincipal(), status: http.StatusForbidden},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/v1/audit-events", nil)
			if test.principal != nil {
				request = request.WithContext(authcontext.SetPrincipal(request.Context(), *test.principal))
			}
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)
			if recorder.Code != test.status {
				t.Fatalf("status = %d, want %d", recorder.Code, test.status)
			}
		})
	}
}

func TestHandlerParsesAliasesAndReturnsStablePagination(t *testing.T) {
	organizationID := mustUUID("30000000-0000-4000-8000-000000000001")
	actorUserID := mustUUID("30000000-0000-4000-8000-000000000002")
	store := &fakeReadStore{
		listFn: func(_ context.Context, params database.ListAuditEventsParams) ([]database.SecurityAuditEvent, error) {
			if params.OrganizationID != organizationID || params.ActorUserID != actorUserID || params.PageOffset != 10 || params.PageSize != 10 {
				t.Fatalf("unexpected list params: %+v", params)
			}
			return []database.SecurityAuditEvent{}, nil
		},
		countFn: func(_ context.Context, params database.CountAuditEventsParams) (int64, error) {
			return 0, nil
		},
	}
	router := chi.NewRouter()
	RegisterRoutes(router, NewHandler(NewService(store)), authz.NewGuard())
	principal := auditPrincipal(organizationID, authz.ScopeAuditRead)
	request := httptest.NewRequest(http.MethodGet, "/v1/audit-events?page=2&page_size=10&actor_user_id="+actorUserID.String(), nil)
	request = request.WithContext(authcontext.SetPrincipal(request.Context(), principal))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	if recorder.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("Cache-Control = %q", recorder.Header().Get("Cache-Control"))
	}
	if recorder.Body.String() != "{\"data\":[],\"pagination\":{\"page\":2,\"pageSize\":10,\"total\":0,\"totalPages\":0}}\n" {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func routeIdentityPrincipal() *authcontext.Principal {
	principal := auditPrincipal(mustUUID("40000000-0000-4000-8000-000000000001"), authz.ScopeAuditRead)
	principal.ContextKind = authcontext.ContextIdentity
	principal.OrganizationID = pgtype.UUID{}
	principal.MembershipID = pgtype.UUID{}
	return &principal
}

func routeOrganizationPrincipal() *authcontext.Principal {
	principal := auditPrincipal(mustUUID("40000000-0000-4000-8000-000000000002"))
	return &principal
}
