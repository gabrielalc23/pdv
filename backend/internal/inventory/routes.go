package inventory

import "github.com/go-chi/chi/v5"

func RegisterRoutes(r chi.Router, h *Handler) {
	r.Get("/inventory", h.ListInventory)
	r.Get("/products/{id}/inventory", h.GetProductInventory)
	r.Post("/inventory/entries", h.CreateEntry)
	r.Post("/inventory/adjustments", h.CreateAdjustment)
	r.Get("/products/{id}/inventory/movements", h.ListMovements)
}
