package roles

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/gabrielalc23/pdv/internal/platform/authcontext"
	"github.com/gabrielalc23/pdv/internal/platform/authz"
)

func TestRoutesRequirePrincipalAndEndpointScope(t *testing.T) {
	router := chi.NewRouter()
	RegisterRoutes(router, NewHandler(nil), authz.NewGuard())

	request := httptest.NewRequest(http.MethodGet, "/v1/scopes", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("without principal status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}

	actor := testActor()
	request = httptest.NewRequest(http.MethodGet, "/v1/scopes", nil)
	request = request.WithContext(authcontext.SetPrincipal(request.Context(), actor))
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("without scopes.read status = %d, want %d", recorder.Code, http.StatusForbidden)
	}
}
