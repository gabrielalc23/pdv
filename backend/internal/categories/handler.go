package categories

import (
	"errors"
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

func (h *Handler) resolveActor(w http.ResponseWriter, r *http.Request) (authn.OrganizationActor, bool) {
	p, err := authcontext.MustPrincipal(r.Context())
	if err != nil {
		apphttp.WriteError(w, http.StatusUnauthorized, "access_token_missing", "authentication required", "")
		return authn.OrganizationActor{}, false
	}
	actor, err := authn.OrganizationActorFromPrincipal(p)
	if err != nil {
		apphttp.WriteError(w, http.StatusBadRequest, "organization_context_required", "organization context is required", "")
		return authn.OrganizationActor{}, false
	}
	return actor, true
}

func (h *Handler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.resolveActor(w, r)
	if !ok {
		return
	}

	var input UpsertCategoryInput
	if err := apphttp.DecodeJSONBody(r, &input); err != nil {
		apphttp.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must contain a single valid JSON document", "")
		return
	}
	result, err := h.service.Create(r.Context(), actor, input)
	if err != nil {
		h.writeServiceError(w, err, http.StatusUnprocessableEntity)
		return
	}
	apphttp.WriteJSON(w, http.StatusCreated, result)
}

func (h *Handler) ListCategories(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.resolveActor(w, r)
	if !ok {
		return
	}

	input, err := parseListQuery(r)
	if err != nil {
		var validationErr *ValidationError
		if errors.As(err, &validationErr) {
			apphttp.WriteError(w, http.StatusBadRequest, "validation_error", validationErr.Message, validationErr.Field)
			return
		}
		apphttp.WriteError(w, http.StatusBadRequest, "validation_error", err.Error(), "")
		return
	}
	result, err := h.service.List(r.Context(), actor, input)
	if err != nil {
		h.writeServiceError(w, err, http.StatusBadRequest)
		return
	}
	apphttp.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) GetCategory(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.resolveActor(w, r)
	if !ok {
		return
	}

	result, err := h.service.Get(r.Context(), actor, chi.URLParam(r, "id"))
	if err != nil {
		h.writeServiceError(w, err, http.StatusBadRequest)
		return
	}
	apphttp.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.resolveActor(w, r)
	if !ok {
		return
	}

	var input UpsertCategoryInput
	if err := apphttp.DecodeJSONBody(r, &input); err != nil {
		apphttp.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must contain a single valid JSON document", "")
		return
	}
	result, err := h.service.Update(r.Context(), actor, chi.URLParam(r, "id"), input)
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
	actor, ok := h.resolveActor(w, r)
	if !ok {
		return
	}

	var result CategoryResponse
	var err error
	if active {
		result, err = h.service.Activate(r.Context(), actor, chi.URLParam(r, "id"))
	} else {
		result, err = h.service.Deactivate(r.Context(), actor, chi.URLParam(r, "id"))
	}
	if err != nil {
		h.writeServiceError(w, err, http.StatusUnprocessableEntity)
		return
	}
	apphttp.WriteJSON(w, http.StatusOK, result)
}
