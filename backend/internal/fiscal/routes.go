package fiscal

import "github.com/go-chi/chi/v5"

func RegisterRoutes(r chi.Router, h *Handler) {
	r.Get("/sales/{id}/fiscal-document", h.GetFiscalDocument)
}
