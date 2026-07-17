package categories

import "github.com/go-chi/chi/v5"

func RegisterRoutes(r chi.Router, h *Handler) {
	r.Post("/categories", h.CreateCategory)
	r.Get("/categories", h.ListCategories)
	r.Get("/categories/{id}", h.GetCategory)
	r.Put("/categories/{id}", h.UpdateCategory)
	r.Post("/categories/{id}/activate", h.ActivateCategory)
	r.Post("/categories/{id}/deactivate", h.DeactivateCategory)
}
