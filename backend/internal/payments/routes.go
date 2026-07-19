package payments

import (
	"github.com/go-chi/chi/v5"

	"github.com/gabrielalc23/pdv/internal/platform/authz"
)

func RegisterRoutes(r chi.Router, h *Handler, guard authz.Guard) {
	r.With(
		guard.RequireStoreContext(),
		guard.RequireAll(authz.ScopePaymentMethodsRead),
	).Get("/payment-methods", h.ListPaymentMethods)

	r.With(
		guard.RequireStoreContext(),
		guard.RequireAll(authz.ScopePaymentsRead),
	).Get("/sales/{id}/payments", h.ListSalePayments)
}
