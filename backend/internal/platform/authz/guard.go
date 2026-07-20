package authz

import (
	"net/http"

	"github.com/gabrielalc23/pdv/internal/platform/authcontext"
	platformhttp "github.com/gabrielalc23/pdv/internal/platform/http"
)

type Guard struct{}

func NewGuard() Guard { return Guard{} }

func (g Guard) RequireIdentity() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := authcontext.MustPrincipal(r.Context())
			if err != nil {
				writeAuthzError(w, 401, "ACCESS_TOKEN_MISSING", "authentication required")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (g Guard) RequireOrganizationContext() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p, err := authcontext.MustPrincipal(r.Context())
			if err != nil {
				writeAuthzError(w, 401, "ACCESS_TOKEN_MISSING", "authentication required")
				return
			}

			if !p.HasOrganizationScope() {
				writeAuthzError(w, 400, CodeOrganizationContextReq, "organization context is required for this operation")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (g Guard) RequireStoreContext() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p, err := authcontext.MustPrincipal(r.Context())
			if err != nil {
				writeAuthzError(w, 401, "ACCESS_TOKEN_MISSING", "authentication required")
				return
			}

			if !p.HasStoreScope() {
				writeAuthzError(w, 400, CodeStoreContextReq, "store context is required for this operation")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (g Guard) RequireAll(scopes ...authcontext.Scope) func(http.Handler) http.Handler {
	if len(scopes) == 0 {
		panic("authz: RequireAll requires at least one scope")
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p, err := authcontext.MustPrincipal(r.Context())
			if err != nil {
				writeAuthzError(w, 401, "ACCESS_TOKEN_MISSING", "authentication required")
				return
			}

			if !p.Scopes.HasAll(scopes...) {
				writeAuthzError(w, 403, CodeInsufficientScope, "insufficient permissions")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (g Guard) RequireAny(scopes ...authcontext.Scope) func(http.Handler) http.Handler {
	if len(scopes) == 0 {
		panic("authz: RequireAny requires at least one scope")
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p, err := authcontext.MustPrincipal(r.Context())
			if err != nil {
				writeAuthzError(w, 401, "ACCESS_TOKEN_MISSING", "authentication required")
				return
			}

			if !p.Scopes.HasAny(scopes...) {
				writeAuthzError(w, 403, CodeInsufficientScope, "insufficient permissions")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func writeAuthzError(w http.ResponseWriter, status int, code, message string) {
	if status == 401 {
		w.Header().Set("WWW-Authenticate", "Bearer")
	}
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	platformhttp.WriteJSON(w, status, platformhttp.ErrorResponse{
		Error: platformhttp.ErrorDetails{
			Code:    code,
			Message: message,
		},
	})
}
