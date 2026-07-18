package sales

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

func (h *Handler) resolveActor(w http.ResponseWriter, r *http.Request) (tenancy.ActorScope, bool) {
	if h.resolver == nil {
		apphttp.WriteError(w, http.StatusInternalServerError, "tenant_context_unavailable", "tenant resolver not configured", "")
		return tenancy.ActorScope{}, false
	}
	scope, err := h.resolver.Actor(r.Context())
	if err != nil {
		apphttp.WriteError(w, http.StatusUnauthorized, "tenant_context_unavailable", "actor scope is required", "")
		return tenancy.ActorScope{}, false
	}
	return scope, true
}

func (h *Handler) CreateSale(w http.ResponseWriter, r *http.Request) {
	scope, ok := h.resolveActor(w, r)
	if !ok {
		return
	}

	var input CreateSaleInput
	if err := apphttp.DecodeJSONBody(r, &input); err != nil {
		apphttp.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must contain a single valid JSON document", "")
		return
	}

	result, err := h.service.Create(r.Context(), scope, input)
	if err != nil {
		h.writeServiceError(w, err, http.StatusUnprocessableEntity)
		return
	}

	apphttp.WriteJSON(w, http.StatusCreated, result)
}

func (h *Handler) ListSales(w http.ResponseWriter, r *http.Request) {
	scope, ok := h.resolveActor(w, r)
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

func (h *Handler) GetSale(w http.ResponseWriter, r *http.Request) {
	scope, ok := h.resolveActor(w, r)
	if !ok {
		return
	}

	result, err := h.service.Get(r.Context(), scope, chi.URLParam(r, "id"))
	if err != nil {
		h.writeServiceError(w, err, http.StatusBadRequest)
		return
	}

	apphttp.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) AddItem(w http.ResponseWriter, r *http.Request) {
	scope, ok := h.resolveActor(w, r)
	if !ok {
		return
	}

	var input AddSaleItemInput
	if err := apphttp.DecodeJSONBody(r, &input); err != nil {
		apphttp.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must contain a single valid JSON document", "")
		return
	}

	result, err := h.service.AddItem(r.Context(), scope, chi.URLParam(r, "id"), input)
	if err != nil {
		h.writeServiceError(w, err, http.StatusUnprocessableEntity)
		return
	}

	apphttp.WriteJSON(w, http.StatusCreated, result)
}

func (h *Handler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	scope, ok := h.resolveActor(w, r)
	if !ok {
		return
	}

	var input UpdateSaleItemInput
	if err := apphttp.DecodeJSONBody(r, &input); err != nil {
		apphttp.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must contain a single valid JSON document", "")
		return
	}

	result, err := h.service.UpdateItem(r.Context(), scope, chi.URLParam(r, "id"), chi.URLParam(r, "itemId"), input)
	if err != nil {
		h.writeServiceError(w, err, http.StatusUnprocessableEntity)
		return
	}

	apphttp.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) RemoveItem(w http.ResponseWriter, r *http.Request) {
	scope, ok := h.resolveActor(w, r)
	if !ok {
		return
	}

	result, err := h.service.RemoveItem(r.Context(), scope, chi.URLParam(r, "id"), chi.URLParam(r, "itemId"))
	if err != nil {
		h.writeServiceError(w, err, http.StatusUnprocessableEntity)
		return
	}

	apphttp.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) CancelSale(w http.ResponseWriter, r *http.Request) {
	scope, ok := h.resolveActor(w, r)
	if !ok {
		return
	}

	result, err := h.service.Cancel(r.Context(), scope, chi.URLParam(r, "id"))
	if err != nil {
		h.writeServiceError(w, err, http.StatusUnprocessableEntity)
		return
	}

	apphttp.WriteJSON(w, http.StatusOK, result)
}
