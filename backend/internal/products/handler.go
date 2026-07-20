package products

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

func (h *Handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.resolveActor(w, r)
	if !ok {
		return
	}

	var input UpsertProductInput
	if err := apphttp.DecodeJSONBody(r, &input); err != nil {
		apphttp.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must contain a single valid JSON document", "")
		return
	}

	product, err := h.service.Create(r.Context(), actor, input)
	if err != nil {
		h.writeServiceError(w, err, http.StatusUnprocessableEntity)
		return
	}

	apphttp.WriteJSON(w, http.StatusCreated, product)
}

func (h *Handler) GetProduct(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.resolveActor(w, r)
	if !ok {
		return
	}

	product, err := h.service.Get(r.Context(), actor, chi.URLParam(r, "id"))
	if err != nil {
		h.writeServiceError(w, err, http.StatusUnprocessableEntity)
		return
	}

	apphttp.WriteJSON(w, http.StatusOK, product)
}

func (h *Handler) ListProducts(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.resolveActor(w, r)
	if !ok {
		return
	}

	input, err := parseListQuery(r)
	if err != nil {
		h.writeValidationError(w, err, http.StatusBadRequest)
		return
	}

	result, err := h.service.List(r.Context(), actor, input)
	if err != nil {
		h.writeServiceError(w, err, http.StatusBadRequest)
		return
	}

	apphttp.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.resolveActor(w, r)
	if !ok {
		return
	}

	var input UpsertProductInput
	if err := apphttp.DecodeJSONBody(r, &input); err != nil {
		apphttp.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must contain a single valid JSON document", "")
		return
	}

	product, err := h.service.Update(r.Context(), actor, chi.URLParam(r, "id"), input)
	if err != nil {
		h.writeServiceError(w, err, http.StatusUnprocessableEntity)
		return
	}

	apphttp.WriteJSON(w, http.StatusOK, product)
}

func (h *Handler) ActivateProduct(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.resolveActor(w, r)
	if !ok {
		return
	}

	product, err := h.service.Activate(r.Context(), actor, chi.URLParam(r, "id"))
	if err != nil {
		h.writeServiceError(w, err, http.StatusUnprocessableEntity)
		return
	}

	apphttp.WriteJSON(w, http.StatusOK, product)
}

func (h *Handler) DeactivateProduct(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.resolveActor(w, r)
	if !ok {
		return
	}

	product, err := h.service.Deactivate(r.Context(), actor, chi.URLParam(r, "id"))
	if err != nil {
		h.writeServiceError(w, err, http.StatusUnprocessableEntity)
		return
	}

	apphttp.WriteJSON(w, http.StatusOK, product)
}
