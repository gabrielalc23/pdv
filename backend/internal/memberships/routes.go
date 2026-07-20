package memberships

import (
	"github.com/go-chi/chi/v5"

	"github.com/gabrielalc23/pdv/internal/platform/authn"
	"github.com/gabrielalc23/pdv/internal/platform/authz"
)

func RegisterRoutes(r chi.Router, handler *Handler, authentication *authn.Middleware, guard authz.Guard) {
	r.Group(func(protected chi.Router) {
		protected.Use(authentication.RequireAccessToken)
		protected.Use(guard.RequireOrganizationContext())
		protected.With(guard.RequireAll(authz.ScopeMembersRead)).Get("/members", handler.List)
		protected.With(guard.RequireAll(authz.ScopeMembersRead)).Get("/members/{membershipId}", handler.Get)
		protected.With(guard.RequireAll(authz.ScopeMembersStatusUpdate)).Patch("/members/{membershipId}/default-store", handler.UpdateDefaultStore)
		protected.With(guard.RequireAll(authz.ScopeMembersStatusUpdate)).Post("/members/{membershipId}/suspend", handler.Suspend)
		protected.With(guard.RequireAll(authz.ScopeMembersStatusUpdate)).Post("/members/{membershipId}/reactivate", handler.Reactivate)
		protected.With(guard.RequireAll(authz.ScopeMembersRemove)).Delete("/members/{membershipId}", handler.Remove)
	})
}
