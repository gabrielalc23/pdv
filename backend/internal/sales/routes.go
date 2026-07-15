package sales

import "github.com/go-chi/chi/v5"

func RegisterRoutes(r chi.Router, h *Handler) {
	r.Post("/sales", h.CreateSale)
	r.Get("/sales", h.ListSales)
	r.Get("/sales/{id}", h.GetSale)
	r.Post("/sales/{id}/items", h.AddItem)
	r.Put("/sales/{id}/items/{itemId}", h.UpdateItem)
	r.Delete("/sales/{id}/items/{itemId}", h.RemoveItem)
	r.Post("/sales/{id}/cancel", h.CancelSale)
}
