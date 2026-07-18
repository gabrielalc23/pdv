package inventory

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

func (h *Handler) ListInventory(w http.ResponseWriter, r *http.Request) {
	scope, ok := h.resolveStore(w, r)
	if !ok {
		return
	}

	input, err := parseListInventoryQuery(r)
	if err != nil {
		h.writeValidationError(w, err, http.StatusBadRequest)
		return
	}

	result, err := h.service.ListInventory(r.Context(), scope, input)
	if err != nil {
		h.writeServiceError(w, err, http.StatusBadRequest)
		return
	}

	apphttp.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) GetProductInventory(w http.ResponseWriter, r *http.Request) {
	scope, ok := h.resolveStore(w, r)
	if !ok {
		return
	}

	result, err := h.service.GetProductInventory(r.Context(), scope, chi.URLParam(r, "id"))
	if err != nil {
		h.writeServiceError(w, err, http.StatusBadRequest)
		return
	}

	apphttp.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) CreateEntry(w http.ResponseWriter, r *http.Request) {
	scope, ok := h.resolveActor(w, r)
	if !ok {
		return
	}

	var input CreateInventoryEntryInput
	if err := apphttp.DecodeJSONBody(r, &input); err != nil {
		apphttp.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must contain a single valid JSON document", "")
		return
	}

	result, err := h.service.CreateEntry(r.Context(), scope, input)
	if err != nil {
		h.writeServiceError(w, err, http.StatusUnprocessableEntity)
		return
	}

	apphttp.WriteJSON(w, http.StatusCreated, result)
}

func (h *Handler) CreateAdjustment(w http.ResponseWriter, r *http.Request) {
	scope, ok := h.resolveActor(w, r)
	if !ok {
		return
	}

	var input CreateInventoryAdjustmentInput
	if err := apphttp.DecodeJSONBody(r, &input); err != nil {
		apphttp.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must contain a single valid JSON document", "")
		return
	}

	result, err := h.service.CreateAdjustment(r.Context(), scope, input)
	if err != nil {
		h.writeServiceError(w, err, http.StatusUnprocessableEntity)
		return
	}

	apphttp.WriteJSON(w, http.StatusCreated, result)
}

func (h *Handler) ListMovements(w http.ResponseWriter, r *http.Request) {
	scope, ok := h.resolveStore(w, r)
	if !ok {
		return
	}

	input, err := parseListMovementsQuery(r)
	if err != nil {
		h.writeValidationError(w, err, http.StatusBadRequest)
		return
	}

	result, err := h.service.ListMovements(r.Context(), scope, chi.URLParam(r, "id"), input)
	if err != nil {
		h.writeServiceError(w, err, http.StatusBadRequest)
		return
	}

	apphttp.WriteJSON(w, http.StatusOK, result)
}
