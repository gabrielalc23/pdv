package sales

import (
	"net/http"

	"github.com/gabrielalc23/pdv/internal/platform/authcontext"
	"github.com/gabrielalc23/pdv/internal/platform/authn"
	apphttp "github.com/gabrielalc23/pdv/internal/platform/http"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
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

func (h *Handler) CreateSale(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.resolveActor(w, r)
	if !ok {
		return
	}

	var input CreateSaleInput
	if err := apphttp.DecodeJSONBody(r, &input); err != nil {
		apphttp.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must contain a single valid JSON document", "")
		return
	}
	result, err := h.service.Create(r.Context(), actor, input)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	apphttp.WriteJSON(w, http.StatusCreated, result)
}

func (h *Handler) ListSales(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.resolveActor(w, r)
	if !ok {
		return
	}
	input, err := parseListQuery(r)
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

func (h *Handler) GetSale(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.resolveActor(w, r)
	if !ok {
		return
	}
	result, err := h.service.Get(r.Context(), actor, chi.URLParam(r, "id"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	apphttp.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) AddItem(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.resolveActor(w, r)
	if !ok {
		return
	}
	var input AddSaleItemInput
	if err := apphttp.DecodeJSONBody(r, &input); err != nil {
		apphttp.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must contain a single valid JSON document", "")
		return
	}
	result, err := h.service.AddItem(r.Context(), actor, chi.URLParam(r, "id"), input)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	apphttp.WriteJSON(w, http.StatusCreated, result)
}

func (h *Handler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.resolveActor(w, r)
	if !ok {
		return
	}
	var input UpdateSaleItemInput
	if err := apphttp.DecodeJSONBody(r, &input); err != nil {
		apphttp.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must contain a single valid JSON document", "")
		return
	}
	result, err := h.service.UpdateItem(r.Context(), actor, chi.URLParam(r, "id"), chi.URLParam(r, "itemId"), input)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	apphttp.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) RemoveItem(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.resolveActor(w, r)
	if !ok {
		return
	}
	result, err := h.service.RemoveItem(r.Context(), actor, chi.URLParam(r, "id"), chi.URLParam(r, "itemId"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	apphttp.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) CancelSale(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.resolveActor(w, r)
	if !ok {
		return
	}
	result, err := h.service.Cancel(r.Context(), actor, chi.URLParam(r, "id"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	apphttp.WriteJSON(w, http.StatusOK, result)
}
