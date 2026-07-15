package sales

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

func (h *Handler) CreateSale(w http.ResponseWriter, r *http.Request) {
	var input CreateSaleInput
	if err := apphttp.DecodeJSONBody(r, &input); err != nil {
		apphttp.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must contain a single valid JSON document", "")
		return
	}

	result, err := h.service.Create(r.Context(), input)
	if err != nil {
		h.writeServiceError(w, err, http.StatusUnprocessableEntity)
		return
	}

	apphttp.WriteJSON(w, http.StatusCreated, result)
}

func (h *Handler) ListSales(w http.ResponseWriter, r *http.Request) {
	input, err := parseListQuery(r)
	if err != nil {
		h.writeValidationError(w, err, http.StatusBadRequest)
		return
	}

	result, err := h.service.List(r.Context(), input)
	if err != nil {
		h.writeServiceError(w, err, http.StatusBadRequest)
		return
	}

	apphttp.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) GetSale(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.Get(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		h.writeServiceError(w, err, http.StatusBadRequest)
		return
	}

	apphttp.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) AddItem(w http.ResponseWriter, r *http.Request) {
	var input AddSaleItemInput
	if err := apphttp.DecodeJSONBody(r, &input); err != nil {
		apphttp.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must contain a single valid JSON document", "")
		return
	}

	result, err := h.service.AddItem(r.Context(), chi.URLParam(r, "id"), input)
	if err != nil {
		h.writeServiceError(w, err, http.StatusUnprocessableEntity)
		return
	}

	apphttp.WriteJSON(w, http.StatusCreated, result)
}

func (h *Handler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	var input UpdateSaleItemInput
	if err := apphttp.DecodeJSONBody(r, &input); err != nil {
		apphttp.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must contain a single valid JSON document", "")
		return
	}

	result, err := h.service.UpdateItem(r.Context(), chi.URLParam(r, "id"), chi.URLParam(r, "itemId"), input)
	if err != nil {
		h.writeServiceError(w, err, http.StatusUnprocessableEntity)
		return
	}

	apphttp.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) RemoveItem(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.RemoveItem(r.Context(), chi.URLParam(r, "id"), chi.URLParam(r, "itemId"))
	if err != nil {
		h.writeServiceError(w, err, http.StatusUnprocessableEntity)
		return
	}

	apphttp.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) CancelSale(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.Cancel(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		h.writeServiceError(w, err, http.StatusUnprocessableEntity)
		return
	}

	apphttp.WriteJSON(w, http.StatusOK, result)
}
