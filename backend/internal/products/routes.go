package products

import "github.com/go-chi/chi/v5"

func RegisterRoutes(r chi.Router, h *Handler) {
	r.Post("/products", h.CreateProduct)
	r.Get("/products", h.ListProducts)
	r.Get("/products/{id}", h.GetProduct)
	r.Put("/products/{id}", h.UpdateProduct)
	r.Post("/products/{id}/activate", h.ActivateProduct)
	r.Post("/products/{id}/deactivate", h.DeactivateProduct)
}
