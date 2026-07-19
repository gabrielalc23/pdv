package roles

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/gabrielalc23/pdv/internal/platform/authcontext"
	apphttp "github.com/gabrielalc23/pdv/internal/platform/http"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func principal(w http.ResponseWriter, r *http.Request) (authcontext.Principal, bool) {
	actor, err := authcontext.MustPrincipal(r.Context())
	if err != nil {
		apphttp.WriteError(w, http.StatusUnauthorized, "ACCESS_TOKEN_MISSING", "Authentication is required.", "")
		return authcontext.Principal{}, false
	}
	return actor, true
}

func (h *Handler) ListScopes(w http.ResponseWriter, r *http.Request) {
	actor, ok := principal(w, r)
	if !ok {
		return
	}
	response, err := h.service.ListScopes(r.Context(), actor)
	if err != nil {
		h.writeError(w, err)
		return
	}
	apphttp.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) ListRoles(w http.ResponseWriter, r *http.Request) {
	actor, ok := principal(w, r)
	if !ok {
		return
	}
	response, err := h.service.ListRoles(r.Context(), actor)
	if err != nil {
		h.writeError(w, err)
		return
	}
	apphttp.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) CreateRole(w http.ResponseWriter, r *http.Request) {
	actor, ok := principal(w, r)
	if !ok {
		return
	}
	var input UpsertRoleInput
	if err := apphttp.DecodeJSONBody(r, &input); err != nil {
		apphttp.WriteError(w, http.StatusBadRequest, "INVALID_JSON", "Request body must contain one valid JSON document.", "")
		return
	}
	response, err := h.service.CreateRole(r.Context(), actor, input)
	if err != nil {
		h.writeError(w, err)
		return
	}
	apphttp.WriteJSON(w, http.StatusCreated, response)
}

func (h *Handler) GetRole(w http.ResponseWriter, r *http.Request) {
	actor, ok := principal(w, r)
	if !ok {
		return
	}
	response, err := h.service.GetRole(r.Context(), actor, chi.URLParam(r, "roleId"))
	if err != nil {
		h.writeError(w, err)
		return
	}
	apphttp.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	actor, ok := principal(w, r)
	if !ok {
		return
	}
	var input UpsertRoleInput
	if err := apphttp.DecodeJSONBody(r, &input); err != nil {
		apphttp.WriteError(w, http.StatusBadRequest, "INVALID_JSON", "Request body must contain one valid JSON document.", "")
		return
	}
	response, err := h.service.UpdateRole(r.Context(), actor, chi.URLParam(r, "roleId"), input)
	if err != nil {
		h.writeError(w, err)
		return
	}
	apphttp.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) ActivateRole(w http.ResponseWriter, r *http.Request) {
	h.setRoleActive(w, r, true)
}

func (h *Handler) DeactivateRole(w http.ResponseWriter, r *http.Request) {
	h.setRoleActive(w, r, false)
}

func (h *Handler) setRoleActive(w http.ResponseWriter, r *http.Request, active bool) {
	actor, ok := principal(w, r)
	if !ok {
		return
	}
	var response RoleResponse
	var err error
	if active {
		response, err = h.service.ActivateRole(r.Context(), actor, chi.URLParam(r, "roleId"))
	} else {
		response, err = h.service.DeactivateRole(r.Context(), actor, chi.URLParam(r, "roleId"))
	}
	if err != nil {
		h.writeError(w, err)
		return
	}
	apphttp.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) CreateBinding(w http.ResponseWriter, r *http.Request) {
	actor, ok := principal(w, r)
	if !ok {
		return
	}
	var input CreateBindingInput
	if err := apphttp.DecodeJSONBody(r, &input); err != nil {
		apphttp.WriteError(w, http.StatusBadRequest, "INVALID_JSON", "Request body must contain one valid JSON document.", "")
		return
	}
	response, created, err := h.service.CreateBinding(r.Context(), actor, chi.URLParam(r, "membershipId"), input)
	if err != nil {
		h.writeError(w, err)
		return
	}
	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	apphttp.WriteJSON(w, status, response)
}

func (h *Handler) DeleteBinding(w http.ResponseWriter, r *http.Request) {
	actor, ok := principal(w, r)
	if !ok {
		return
	}
	if err := h.service.DeleteBinding(r.Context(), actor, chi.URLParam(r, "membershipId"), chi.URLParam(r, "bindingId")); err != nil {
		h.writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
