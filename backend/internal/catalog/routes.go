package catalog

import (
	"github.com/go-chi/chi/v5"

	"github.com/gabrielalc23/pdv/internal/platform/authz"
)

func RegisterRoutes(r chi.Router, h *Handler, guard authz.Guard) {
	r.With(
		guard.RequireStoreContext(),
		guard.RequireAll(authz.ScopeCatalogRead),
	).Get("/catalog", h.ListCatalog)

	r.With(
		guard.RequireStoreContext(),
		guard.RequireAll(authz.ScopeCatalogRead),
	).Get("/catalog/barcode/{barcode}", h.GetCatalogProductByBarcode)

	r.With(
		guard.RequireStoreContext(),
		guard.RequireAll(authz.ScopeCatalogRead),
	).Get("/catalog/{id}", h.GetCatalogProduct)
}
