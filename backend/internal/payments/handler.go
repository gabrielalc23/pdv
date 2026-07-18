package payments

import (
	"net/http"

	"github.com/gabrielalc23/pdv/internal/platform/tenancy"
	apphttp "github.com/gabrielalc23/pdv/internal/platform/http"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service  *Service
	resolver tenancy.Resolver
}

func NewHandler(service *Service, resolver tenancy.Resolver) *Handler {
	return &Handler{service: service, resolver: resolver}
}

func (h *Handler) resolveOrg(w http.ResponseWriter, r *http.Request) (tenancy.OrganizationScope, bool) {
	if h.resolver == nil {
		apphttp.WriteError(w, http.StatusInternalServerError, "tenant_context_unavailable", "tenant resolver not configured", "")
		return tenancy.OrganizationScope{}, false
	}
	scope, err := h.resolver.Organization(r.Context())
	if err != nil {
		apphttp.WriteError(w, http.StatusUnauthorized, "tenant_context_unavailable", "organization scope is required", "")
		return tenancy.OrganizationScope{}, false
	}
	return scope, true
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

func (h *Handler) ListPaymentMethods(w http.ResponseWriter, r *http.Request) {
	scope, ok := h.resolveOrg(w, r)
	if !ok {
		return
	}

	result, err := h.service.ListPaymentMethods(r.Context(), scope)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	apphttp.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) ListSalePayments(w http.ResponseWriter, r *http.Request) {
	scope, ok := h.resolveStore(w, r)
	if !ok {
		return
	}

	result, err := h.service.ListSalePayments(r.Context(), scope, chi.URLParam(r, "id"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	apphttp.WriteJSON(w, http.StatusOK, result)
}
