package categories

import (
	"github.com/go-chi/chi/v5"

	"github.com/gabrielalc23/pdv/internal/platform/authz"
)

func RegisterRoutes(r chi.Router, h *Handler, guard authz.Guard) {
	r.With(
		guard.RequireOrganizationContext(),
		guard.RequireAll(authz.ScopeCategoriesCreate),
	).Post("/categories", h.CreateCategory)

	r.With(
		guard.RequireOrganizationContext(),
		guard.RequireAll(authz.ScopeCategoriesRead),
	).Get("/categories", h.ListCategories)

	r.With(
		guard.RequireOrganizationContext(),
		guard.RequireAll(authz.ScopeCategoriesRead),
	).Get("/categories/{id}", h.GetCategory)

	r.With(
		guard.RequireOrganizationContext(),
		guard.RequireAll(authz.ScopeCategoriesUpdate),
	).Put("/categories/{id}", h.UpdateCategory)

	r.With(
		guard.RequireOrganizationContext(),
		guard.RequireAll(authz.ScopeCategoriesStatusUpdate),
	).Post("/categories/{id}/activate", h.ActivateCategory)

	r.With(
		guard.RequireOrganizationContext(),
		guard.RequireAll(authz.ScopeCategoriesStatusUpdate),
	).Post("/categories/{id}/deactivate", h.DeactivateCategory)
}

func RegisterPublicRoutes(r chi.Router, h *Handler) {
	r.Post("/categories", h.CreateCategory)
	r.Get("/categories", h.ListCategories)
	r.Get("/categories/{id}", h.GetCategory)
	r.Put("/categories/{id}", h.UpdateCategory)
	r.Post("/categories/{id}/activate", h.ActivateCategory)
	r.Post("/categories/{id}/deactivate", h.DeactivateCategory)
}

var _ = RegisterPublicRoutes
