package roles

import (
	"github.com/go-chi/chi/v5"

	"github.com/gabrielalc23/pdv/internal/platform/authz"
)

func RegisterRoutes(r chi.Router, h *Handler, guard authz.Guard) {
	r.With(guard.RequireOrganizationContext(), guard.RequireAll(authz.ScopeScopesRead)).Get("/v1/scopes", h.ListScopes)
	r.With(guard.RequireOrganizationContext(), guard.RequireAll(authz.ScopeRolesRead)).Get("/v1/roles", h.ListRoles)
	r.With(guard.RequireOrganizationContext(), guard.RequireAll(authz.ScopeRolesCreate)).Post("/v1/roles", h.CreateRole)
	r.With(guard.RequireOrganizationContext(), guard.RequireAll(authz.ScopeRolesRead)).Get("/v1/roles/{roleId}", h.GetRole)
	r.With(guard.RequireOrganizationContext(), guard.RequireAll(authz.ScopeRolesUpdate)).Put("/v1/roles/{roleId}", h.UpdateRole)
	r.With(guard.RequireOrganizationContext(), guard.RequireAll(authz.ScopeRolesStatusUpdate)).Post("/v1/roles/{roleId}/activate", h.ActivateRole)
	r.With(guard.RequireOrganizationContext(), guard.RequireAll(authz.ScopeRolesStatusUpdate)).Post("/v1/roles/{roleId}/deactivate", h.DeactivateRole)
	r.With(guard.RequireOrganizationContext(), guard.RequireAll(authz.ScopeRolesAssign)).Post("/v1/members/{membershipId}/role-bindings", h.CreateBinding)
	r.With(guard.RequireOrganizationContext(), guard.RequireAll(authz.ScopeRolesAssign)).Delete("/v1/members/{membershipId}/role-bindings/{bindingId}", h.DeleteBinding)
}
