package catalog

import (
	"net/http"

	apphttp "github.com/gabrielalc23/pdv/internal/platform/http"
	"github.com/gabrielalc23/pdv/internal/platform/tenancy"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service  *Service
	resolver tenancy.Resolver
}

func NewHandler(service *Service, resolver tenancy.Resolver) *Handler {
	return &Handler{service: service, resolver: resolver}
}

func (h *Handler) resolveStore(w http.ResponseWriter, r *http.Request) (tenancy.StoreScope, bool) {
	if h.resolver == nil {
		apphttp.WriteError(w, http.StatusInternalServerError, "tenant_context_unavailable", "tenant resolver not configured", "")
		return tenancy.StoreScope{}, false
	}
	scope, err := h.resolver.Store(r.Context())
	if err != nil {
		apphttp.WriteError(w, http.StatusUnauthorized, "tenant_context_unavailable", "store scope is required", "")
		return tenancy.StoreScope{}, false
	}
	return scope, true
}

func (h *Handler) ListCatalog(w http.ResponseWriter, r *http.Request) {
	scope, ok := h.resolveStore(w, r)
	if !ok {
		return
	}

	input, err := parseListQuery(r)
	if err != nil {
		h.writeValidationError(w, err, http.StatusBadRequest)
		return
	}

	result, err := h.service.List(r.Context(), scope, input)
	if err != nil {
		h.writeServiceError(w, err, http.StatusBadRequest)
		return
	}

	apphttp.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) GetCatalogProduct(w http.ResponseWriter, r *http.Request) {
	scope, ok := h.resolveStore(w, r)
	if !ok {
		return
	}

	result, err := h.service.GetByID(r.Context(), scope, chi.URLParam(r, "id"))
	if err != nil {
		h.writeServiceError(w, err, http.StatusBadRequest)
		return
	}

	apphttp.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) GetCatalogProductByBarcode(w http.ResponseWriter, r *http.Request) {
	scope, ok := h.resolveStore(w, r)
	if !ok {
		return
	}

	result, err := h.service.GetByBarcode(r.Context(), scope, chi.URLParam(r, "barcode"))
	if err != nil {
		h.writeServiceError(w, err, http.StatusBadRequest)
		return
	}

	apphttp.WriteJSON(w, http.StatusOK, result)
}
