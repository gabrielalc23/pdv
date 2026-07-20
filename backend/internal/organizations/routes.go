package organizations

import (
	"github.com/go-chi/chi/v5"

	"github.com/gabrielalc23/pdv/internal/platform/authn"
	"github.com/gabrielalc23/pdv/internal/platform/authz"
)

// RegisterRoutes registers paths relative to the API version router supplied by
// the application (for example, a router mounted at /v1).
func RegisterRoutes(r chi.Router, h *Handler, middleware *authn.Middleware, guard authz.Guard) {
	r.Group(func(protected chi.Router) {
		protected.Use(middleware.RequireAccessToken)
		protected.Get("/organizations", h.List)
		protected.Post("/organizations", h.Create)
		protected.With(guard.RequireOrganizationContext(), guard.RequireAll(authz.ScopeOrganizationRead)).Get("/organizations/current", h.Current)
		protected.With(guard.RequireOrganizationContext(), guard.RequireAll(authz.ScopeOrganizationUpdate)).Patch("/organizations/current", h.Update)
		protected.With(guard.RequireOrganizationContext(), guard.RequireAll(authz.ScopeOrganizationArchive)).Post("/organizations/current/archive", h.Archive)
	})
}

func RegisterSelfServiceRoutes(r chi.Router, h *Handler, middleware *authn.Middleware) {
	r.Group(func(protected chi.Router) {
		protected.Use(middleware.RequireAccessToken)
		protected.Get("/me/organizations", h.List)
		protected.Get("/me/stores", h.ListStores)
	})
}
