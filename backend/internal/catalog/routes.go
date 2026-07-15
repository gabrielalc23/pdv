package catalog

import "github.com/go-chi/chi/v5"

func RegisterRoutes(r chi.Router, h *Handler) {
	r.Get("/catalog", h.ListCatalog)
	r.Get("/catalog/barcode/{barcode}", h.GetCatalogProductByBarcode)
	r.Get("/catalog/{id}", h.GetCatalogProduct)
}
