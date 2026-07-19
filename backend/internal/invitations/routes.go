package invitations

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/gabrielalc23/pdv/internal/platform/authn"
	"github.com/gabrielalc23/pdv/internal/platform/authz"
)

// RegisterPublicRoutes registers invitation inspection and acceptance on the
// root router because their paths are outside the versioned administration API.
func RegisterPublicRoutes(r chi.Router, handler *Handler, authentication *authn.Middleware) {
	r.Post("/auth/invitations/inspect", handler.Inspect)
	r.With(optionalAccessToken(authentication)).Post("/auth/invitations/accept", handler.Accept)
}

// RegisterAdminRoutes registers paths relative to the version router supplied
// by the application, for example a router mounted at /v1.
func RegisterAdminRoutes(r chi.Router, handler *Handler, authentication *authn.Middleware, guard authz.Guard) {
	r.Group(func(protected chi.Router) {
		protected.Use(authentication.RequireAccessToken)
		protected.Use(guard.RequireOrganizationContext())
		protected.With(guard.RequireAll(authz.ScopeInvitationsRead)).Get("/invitations", handler.List)
		protected.With(guard.RequireAll(authz.ScopeMembersInvite)).Post("/invitations", handler.Create)
		protected.With(guard.RequireAll(authz.ScopeInvitationsManage)).Post("/invitations/{id}/resend", handler.Resend)
		protected.With(guard.RequireAll(authz.ScopeInvitationsManage)).Post("/invitations/{id}/revoke", handler.Revoke)
	})
}

func optionalAccessToken(authentication *authn.Middleware) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		protected := authentication.RequireAccessToken(next)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.TrimSpace(r.Header.Get("Authorization")) == "" {
				next.ServeHTTP(w, r)
				return
			}
			protected.ServeHTTP(w, r)
		})
	}
}
