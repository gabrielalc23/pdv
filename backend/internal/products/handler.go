package products

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

func (h *Handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	scope, ok := h.resolveOrg(w, r)
	if !ok {
		return
	}

	var input UpsertProductInput
	if err := apphttp.DecodeJSONBody(r, &input); err != nil {
		apphttp.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must contain a single valid JSON document", "")
		return
	}

	product, err := h.service.Create(r.Context(), scope, input)
	if err != nil {
		h.writeServiceError(w, err, http.StatusUnprocessableEntity)
		return
	}

	apphttp.WriteJSON(w, http.StatusCreated, product)
}

func (h *Handler) GetProduct(w http.ResponseWriter, r *http.Request) {
	scope, ok := h.resolveOrg(w, r)
	if !ok {
		return
	}

	product, err := h.service.Get(r.Context(), scope, chi.URLParam(r, "id"))
	if err != nil {
		h.writeServiceError(w, err, http.StatusUnprocessableEntity)
		return
	}

	apphttp.WriteJSON(w, http.StatusOK, product)
}

func (h *Handler) ListProducts(w http.ResponseWriter, r *http.Request) {
	scope, ok := h.resolveOrg(w, r)
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

func (h *Handler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	scope, ok := h.resolveOrg(w, r)
	if !ok {
		return
	}

	var input UpsertProductInput
	if err := apphttp.DecodeJSONBody(r, &input); err != nil {
		apphttp.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must contain a single valid JSON document", "")
		return
	}

	product, err := h.service.Update(r.Context(), scope, chi.URLParam(r, "id"), input)
	if err != nil {
		h.writeServiceError(w, err, http.StatusUnprocessableEntity)
		return
	}

	apphttp.WriteJSON(w, http.StatusOK, product)
}

func (h *Handler) ActivateProduct(w http.ResponseWriter, r *http.Request) {
	scope, ok := h.resolveOrg(w, r)
	if !ok {
		return
	}

	product, err := h.service.Activate(r.Context(), scope, chi.URLParam(r, "id"))
	if err != nil {
		h.writeServiceError(w, err, http.StatusUnprocessableEntity)
		return
	}

	apphttp.WriteJSON(w, http.StatusOK, product)
}

func (h *Handler) DeactivateProduct(w http.ResponseWriter, r *http.Request) {
	scope, ok := h.resolveOrg(w, r)
	if !ok {
		return
	}

	product, err := h.service.Deactivate(r.Context(), scope, chi.URLParam(r, "id"))
	if err != nil {
		h.writeServiceError(w, err, http.StatusUnprocessableEntity)
		return
	}

	apphttp.WriteJSON(w, http.StatusOK, product)
}
