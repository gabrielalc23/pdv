package inventory

import (
	"net/http"

	"github.com/gabrielalc23/pdv/internal/platform/authcontext"
	"github.com/gabrielalc23/pdv/internal/platform/authn"
	apphttp "github.com/gabrielalc23/pdv/internal/platform/http"
	"github.com/go-chi/chi/v5"

	"github.com/gabrielalc23/pdv/internal/catalog"
)

type Handler struct {
	service        *Service
	catalogQuerier *catalog.Service
}

func NewHandler(service *Service, catalogQuerier *catalog.Service) *Handler {
	return &Handler{service: service, catalogQuerier: catalogQuerier}
}

func (h *Handler) resolveActor(w http.ResponseWriter, r *http.Request) (authn.StoreActor, bool) {
	p, err := authcontext.MustPrincipal(r.Context())
	if err != nil {
		apphttp.WriteError(w, http.StatusUnauthorized, "access_token_missing", "authentication required", "")
		return authn.StoreActor{}, false
	}
	actor, err := authn.StoreActorFromPrincipal(p)
	if err != nil {
		apphttp.WriteError(w, http.StatusBadRequest, "store_context_required", "store context is required", "")
		return authn.StoreActor{}, false
	}
	return actor, true
}

func (h *Handler) ListInventory(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.resolveActor(w, r)
	if !ok {
		return
	}
	input, err := parseListInventoryQuery(r)
	if err != nil {
		h.writeValidationError(w, err)
		return
	}
	result, err := h.service.List(r.Context(), actor, input)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	apphttp.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) GetProductInventory(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.resolveActor(w, r)
	if !ok {
		return
	}
	result, err := h.service.GetByProductID(r.Context(), actor, chi.URLParam(r, "id"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	apphttp.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) CreateEntry(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.resolveActor(w, r)
	if !ok {
		return
	}
	var input CreateInventoryEntryInput
	if err := apphttp.DecodeJSONBody(r, &input); err != nil {
		apphttp.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must contain a single valid JSON document", "")
		return
	}
	result, err := h.service.CreateEntry(r.Context(), actor, input)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	apphttp.WriteJSON(w, http.StatusCreated, result)
}

func (h *Handler) CreateAdjustment(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.resolveActor(w, r)
	if !ok {
		return
	}
	var input CreateInventoryAdjustmentInput
	if err := apphttp.DecodeJSONBody(r, &input); err != nil {
		apphttp.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must contain a single valid JSON document", "")
		return
	}
	result, err := h.service.CreateAdjustment(r.Context(), actor, input)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	apphttp.WriteJSON(w, http.StatusCreated, result)
}

func (h *Handler) ListMovements(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.resolveActor(w, r)
	if !ok {
		return
	}
	result, err := h.service.ListMovements(r.Context(), actor, chi.URLParam(r, "id"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	apphttp.WriteJSON(w, http.StatusOK, result)
}
