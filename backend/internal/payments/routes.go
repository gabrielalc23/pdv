package payments

import "github.com/go-chi/chi/v5"

func RegisterRoutes(r chi.Router, h *Handler) {
	r.Get("/payment-methods", h.ListPaymentMethods)
	r.Get("/sales/{id}/payments", h.ListSalePayments)
}
