package products

import (
	"github.com/go-chi/chi/v5"

	"github.com/gabrielalc23/pdv/internal/platform/authz"
)

func RegisterRoutes(r chi.Router, h *Handler, guard authz.Guard) {
	r.With(
		guard.RequireOrganizationContext(),
		guard.RequireAll(authz.ScopeProductsCreate),
	).Post("/products", h.CreateProduct)

	r.With(
		guard.RequireOrganizationContext(),
		guard.RequireAll(authz.ScopeProductsRead),
	).Get("/products", h.ListProducts)

	r.With(
		guard.RequireOrganizationContext(),
		guard.RequireAll(authz.ScopeProductsRead),
	).Get("/products/{id}", h.GetProduct)

	r.With(
		guard.RequireOrganizationContext(),
		guard.RequireAll(authz.ScopeProductsUpdate),
	).Put("/products/{id}", h.UpdateProduct)

	r.With(
		guard.RequireOrganizationContext(),
		guard.RequireAll(authz.ScopeProductsStatusUpdate),
	).Post("/products/{id}/activate", h.ActivateProduct)

	r.With(
		guard.RequireOrganizationContext(),
		guard.RequireAll(authz.ScopeProductsStatusUpdate),
	).Post("/products/{id}/deactivate", h.DeactivateProduct)
}
