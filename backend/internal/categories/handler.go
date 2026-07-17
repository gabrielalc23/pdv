package categories

import (
	"net/http"

	apphttp "github.com/gabrielalc23/pdv/internal/platform/http"
	"github.com/go-chi/chi/v5"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (h *Handler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var input UpsertCategoryInput
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

func (h *Handler) ListCategories(w http.ResponseWriter, r *http.Request) {
	input, err := parseListQuery(r)
	if err != nil {
		h.writeValidationError(w, err)
		return
	}
	result, err := h.service.List(r.Context(), input)
	if err != nil {
		h.writeServiceError(w, err, http.StatusBadRequest)
		return
	}
	apphttp.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) GetCategory(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.Get(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		h.writeServiceError(w, err, http.StatusBadRequest)
		return
	}
	apphttp.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	var input UpsertCategoryInput
	if err := apphttp.DecodeJSONBody(r, &input); err != nil {
		apphttp.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must contain a single valid JSON document", "")
		return
	}
	result, err := h.service.Update(r.Context(), chi.URLParam(r, "id"), input)
	if err != nil {
		h.writeServiceError(w, err, http.StatusUnprocessableEntity)
		return
	}
	apphttp.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) ActivateCategory(w http.ResponseWriter, r *http.Request) {
	h.setActive(w, r, true)
}

func (h *Handler) DeactivateCategory(w http.ResponseWriter, r *http.Request) {
	h.setActive(w, r, false)
}

func (h *Handler) setActive(w http.ResponseWriter, r *http.Request, active bool) {
	var result CategoryResponse
	var err error
	if active {
		result, err = h.service.Activate(r.Context(), chi.URLParam(r, "id"))
	} else {
		result, err = h.service.Deactivate(r.Context(), chi.URLParam(r, "id"))
	}
	if err != nil {
		h.writeServiceError(w, err, http.StatusUnprocessableEntity)
		return
	}
	apphttp.WriteJSON(w, http.StatusOK, result)
}
