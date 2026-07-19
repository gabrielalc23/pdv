package sales

import (
	"github.com/go-chi/chi/v5"

	"github.com/gabrielalc23/pdv/internal/platform/authz"
)

func RegisterRoutes(r chi.Router, h *Handler, guard authz.Guard) {
	r.With(
		guard.RequireStoreContext(),
		guard.RequireAll(authz.ScopeSalesCreate),
	).Post("/sales", h.CreateSale)

	r.With(
		guard.RequireStoreContext(),
		guard.RequireAll(authz.ScopeSalesRead),
	).Get("/sales", h.ListSales)

	r.With(
		guard.RequireStoreContext(),
		guard.RequireAll(authz.ScopeSalesRead),
	).Get("/sales/{id}", h.GetSale)

	r.With(
		guard.RequireStoreContext(),
		guard.RequireAll(authz.ScopeSalesItemsManage),
	).Post("/sales/{id}/items", h.AddItem)

	r.With(
		guard.RequireStoreContext(),
		guard.RequireAll(authz.ScopeSalesItemsManage),
	).Put("/sales/{id}/items/{itemId}", h.UpdateItem)

	r.With(
		guard.RequireStoreContext(),
		guard.RequireAll(authz.ScopeSalesItemsManage),
	).Delete("/sales/{id}/items/{itemId}", h.RemoveItem)

	r.With(
		guard.RequireStoreContext(),
		guard.RequireAll(authz.ScopeSalesCancel),
	).Post("/sales/{id}/cancel", h.CancelSale)
}
