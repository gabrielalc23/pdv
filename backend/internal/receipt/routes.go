package receipt

import "github.com/go-chi/chi/v5"

func RegisterRoutes(r chi.Router, h *Handler) {
	r.Get("/sales/{id}/receipt", h.GetReceipt)
}
