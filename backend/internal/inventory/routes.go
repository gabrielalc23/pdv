package inventory

import (
	"github.com/go-chi/chi/v5"

	"github.com/gabrielalc23/pdv/internal/platform/authz"
)

func RegisterRoutes(r chi.Router, h *Handler, guard authz.Guard) {
	r.With(
		guard.RequireStoreContext(),
		guard.RequireAll(authz.ScopeInventoryRead),
	).Get("/inventory", h.ListInventory)

	r.With(
		guard.RequireStoreContext(),
		guard.RequireAll(authz.ScopeInventoryRead),
	).Get("/products/{id}/inventory", h.GetProductInventory)

	r.With(
		guard.RequireStoreContext(),
		guard.RequireAll(authz.ScopeInventoryEntriesCreate),
	).Post("/inventory/entries", h.CreateEntry)

	r.With(
		guard.RequireStoreContext(),
		guard.RequireAll(authz.ScopeInventoryAdjustments),
	).Post("/inventory/adjustments", h.CreateAdjustment)

	r.With(
		guard.RequireStoreContext(),
		guard.RequireAll(authz.ScopeInventoryMovementsRead),
	).Get("/products/{id}/inventory/movements", h.ListMovements)
}
