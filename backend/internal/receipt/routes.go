package receipt

import (
	"github.com/go-chi/chi/v5"

	"github.com/gabrielalc23/pdv/internal/platform/authz"
)

func RegisterRoutes(r chi.Router, h *Handler, guard authz.Guard) {
	r.With(
		guard.RequireStoreContext(),
		guard.RequireAll(authz.ScopeReceiptsRead),
	).Get("/sales/{id}/receipt", h.GetReceipt)
}
