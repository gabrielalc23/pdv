package inventory

import (
	"net/http"

	apphttp "github.com/gabrielalc23/pdv/internal/platform/http"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) ListInventory(w http.ResponseWriter, r *http.Request) {
	input, err := parseListInventoryQuery(r)
	if err != nil {
		h.writeValidationError(w, err, http.StatusBadRequest)
		return
	}

	result, err := h.service.ListInventory(r.Context(), input)
	if err != nil {
		h.writeServiceError(w, err, http.StatusBadRequest)
		return
	}

	apphttp.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) GetProductInventory(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.GetProductInventory(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		h.writeServiceError(w, err, http.StatusBadRequest)
		return
	}

	apphttp.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) CreateEntry(w http.ResponseWriter, r *http.Request) {
	var input CreateInventoryEntryInput
	if err := apphttp.DecodeJSONBody(r, &input); err != nil {
		apphttp.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must contain a single valid JSON document", "")
		return
	}

	result, err := h.service.CreateEntry(r.Context(), input)
	if err != nil {
		h.writeServiceError(w, err, http.StatusUnprocessableEntity)
		return
	}

	apphttp.WriteJSON(w, http.StatusCreated, result)
}

func (h *Handler) CreateAdjustment(w http.ResponseWriter, r *http.Request) {
	var input CreateInventoryAdjustmentInput
	if err := apphttp.DecodeJSONBody(r, &input); err != nil {
		apphttp.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must contain a single valid JSON document", "")
		return
	}

	result, err := h.service.CreateAdjustment(r.Context(), input)
	if err != nil {
		h.writeServiceError(w, err, http.StatusUnprocessableEntity)
		return
	}

	apphttp.WriteJSON(w, http.StatusCreated, result)
}

func (h *Handler) ListMovements(w http.ResponseWriter, r *http.Request) {
	input, err := parseListMovementsQuery(r)
	if err != nil {
		h.writeValidationError(w, err, http.StatusBadRequest)
		return
	}

	result, err := h.service.ListMovements(r.Context(), chi.URLParam(r, "id"), input)
	if err != nil {
		h.writeServiceError(w, err, http.StatusBadRequest)
		return
	}

	apphttp.WriteJSON(w, http.StatusOK, result)
}
