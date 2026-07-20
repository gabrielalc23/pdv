package memberships

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/gabrielalc23/pdv/internal/platform/authcontext"
	platformhttp "github.com/gabrielalc23/pdv/internal/platform/http"
	"github.com/gabrielalc23/pdv/internal/platform/requestmeta"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	input, err := parseListInput(r)
	if err != nil {
		writeError(w, err)
		return
	}
	actor, err := actorFromRequest(r)
	if err != nil {
		writeError(w, err)
		return
	}
	result, err := h.service.List(r.Context(), actor, input)
	if err != nil {
		writeError(w, err)
		return
	}
	platformhttp.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	actor, err := actorFromRequest(r)
	if err != nil {
		writeError(w, err)
		return
	}
	result, err := h.service.Get(r.Context(), actor, chi.URLParam(r, "membershipId"))
	if err != nil {
		writeError(w, err)
		return
	}
	platformhttp.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) UpdateDefaultStore(w http.ResponseWriter, r *http.Request) {
	var input UpdateDefaultStoreInput
	if err := platformhttp.DecodeJSONBody(r, &input); err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Request body must contain one valid JSON document.", "")
		return
	}
	actor, err := actorFromRequest(r)
	if err != nil {
		writeError(w, err)
		return
	}
	result, err := h.service.UpdateDefaultStore(r.Context(), actor, chi.URLParam(r, "membershipId"), input)
	if err != nil {
		writeError(w, err)
		return
	}
	platformhttp.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	var input UpdateStatusInput
	if err := platformhttp.DecodeJSONBody(r, &input); err != nil {
		platformhttp.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Request body must contain one valid JSON document.", "")
		return
	}
	actor, err := actorFromRequest(r)
	if err != nil {
		writeError(w, err)
		return
	}
	result, err := h.service.UpdateStatus(r.Context(), actor, chi.URLParam(r, "membershipId"), input)
	if err != nil {
		writeError(w, err)
		return
	}
	platformhttp.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) Suspend(w http.ResponseWriter, r *http.Request) {
	h.updateStatus(w, r, "SUSPENDED")
}

func (h *Handler) Reactivate(w http.ResponseWriter, r *http.Request) {
	h.updateStatus(w, r, "ACTIVE")
}

func (h *Handler) Remove(w http.ResponseWriter, r *http.Request) {
	h.updateStatus(w, r, "REMOVED")
}

func (h *Handler) updateStatus(w http.ResponseWriter, r *http.Request, status string) {
	actor, err := actorFromRequest(r)
	if err != nil {
		writeError(w, err)
		return
	}
	result, err := h.service.UpdateStatus(r.Context(), actor, chi.URLParam(r, "membershipId"), UpdateStatusInput{Status: status})
	if err != nil {
		writeError(w, err)
		return
	}
	platformhttp.WriteJSON(w, http.StatusOK, result)
}

func actorFromRequest(r *http.Request) (Actor, error) {
	principal, err := authcontext.MustPrincipal(r.Context())
	if err != nil || !principal.HasOrganizationScope() {
		return Actor{}, validationError("context", "organization context is required")
	}
	return Actor{
		UserID:         principal.UserID,
		SessionID:      principal.SessionID,
		OrganizationID: principal.OrganizationID,
		MembershipID:   principal.MembershipID,
		Scopes:         principal.Scopes,
		RequestMeta:    requestmeta.MustFromContext(r.Context()),
	}, nil
}

func parseListInput(r *http.Request) (ListInput, error) {
	input := ListInput{Search: r.URL.Query().Get("search"), Status: r.URL.Query().Get("status")}
	if value := r.URL.Query().Get("page"); value != "" {
		page, err := strconv.Atoi(value)
		if err != nil {
			return ListInput{}, validationError("page", "must be an integer")
		}
		input.Page = &page
	}
	if value := r.URL.Query().Get("pageSize"); value != "" {
		pageSize, err := strconv.Atoi(value)
		if err != nil {
			return ListInput{}, validationError("pageSize", "must be an integer")
		}
		input.PageSize = &pageSize
	}
	return input, nil
}

func writeError(w http.ResponseWriter, err error) {
	var validation *ValidationError
	switch {
	case errors.As(err, &validation):
		status, code := http.StatusUnprocessableEntity, "VALIDATION_ERROR"
		if validation.Field == "context" {
			status, code = http.StatusBadRequest, "ORGANIZATION_CONTEXT_REQUIRED"
		}
		platformhttp.WriteError(w, status, code, validation.Message, validation.Field)
	case errors.Is(err, ErrInsufficientScope):
		platformhttp.WriteError(w, http.StatusForbidden, "INSUFFICIENT_SCOPE", "Insufficient permissions.", "")
	case errors.Is(err, ErrMembershipNotFound):
		platformhttp.WriteError(w, http.StatusNotFound, "MEMBERSHIP_NOT_FOUND", "Membership not found.", "")
	case errors.Is(err, ErrStoreNotAvailable):
		platformhttp.WriteError(w, http.StatusNotFound, "STORE_NOT_FOUND", "Store not found or unavailable to this membership.", "defaultStoreId")
	case errors.Is(err, ErrLastOwnerRequired):
		platformhttp.WriteError(w, http.StatusConflict, "LAST_OWNER_REQUIRED", "The organization must retain at least one active owner.", "")
	case errors.Is(err, ErrInvalidStatusTransition):
		platformhttp.WriteError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Invalid membership status transition.", "status")
	default:
		platformhttp.WriteError(w, http.StatusServiceUnavailable, "AUTH_DEPENDENCY_UNAVAILABLE", "Authentication service is unavailable.", "")
	}
}
