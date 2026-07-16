package checkout

import "github.com/go-chi/chi/v5"

func RegisterRoutes(r chi.Router, h *Handler) {
	r.Post("/sales/{id}/checkout", h.CheckoutSale)
}
