package audit

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gabrielalc23/pdv/internal/platform/authcontext"
	apphttp "github.com/gabrielalc23/pdv/internal/platform/http"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	input, err := parseListQuery(r.URL.Query())
	if err != nil {
		writeReadError(w, err)
		return
	}
	principal, err := authcontext.MustPrincipal(r.Context())
	if err != nil {
		apphttp.WriteError(w, http.StatusUnauthorized, "ACCESS_TOKEN_MISSING", "Authentication is required.", "")
		return
	}
	response, err := h.service.List(r.Context(), principal, input)
	if err != nil {
		writeReadError(w, err)
		return
	}
	apphttp.WriteJSON(w, http.StatusOK, response)
}

func parseListQuery(query url.Values) (ListInput, error) {
	input := ListInput{
		StoreID:           firstValue(query, "storeId", "store_id"),
		EventType:         firstValue(query, "eventType", "event_type"),
		Outcome:           query.Get("outcome"),
		ActorUserID:       firstValue(query, "actorUserId", "actor_user_id"),
		ActorMembershipID: firstValue(query, "actorMembershipId", "actor_membership_id"),
		EntityType:        firstValue(query, "entityType", "entity_type"),
		EntityID:          firstValue(query, "entityId", "entity_id"),
		OccurredFrom:      firstValue(query, "occurredFrom", "occurred_from", "from"),
		OccurredTo:        firstValue(query, "occurredTo", "occurred_to", "to"),
		Sort:              query.Get("sort"), Order: query.Get("order"),
	}
	page, err := optionalPositiveInt(query, "page")
	if err != nil {
		return ListInput{}, err
	}
	pageSize, err := optionalPositiveIntAliases(query, "pageSize", "page_size")
	if err != nil {
		return ListInput{}, err
	}
	input.Page, input.PageSize = page, pageSize
	return input, nil
}

func firstValue(query url.Values, names ...string) string {
	for _, name := range names {
		if value := query.Get(name); value != "" {
			return value
		}
	}
	return ""
}

func optionalPositiveInt(query url.Values, name string) (*int, error) {
	return optionalPositiveIntAliases(query, name)
}

func optionalPositiveIntAliases(query url.Values, names ...string) (*int, error) {
	raw := firstValue(query, names...)
	if raw == "" {
		return nil, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return nil, validationError(names[0], "must be an integer")
	}
	return &value, nil
}

func writeReadError(w http.ResponseWriter, err error) {
	var validation *ValidationError
	switch {
	case errors.As(err, &validation):
		apphttp.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", validation.Message, validation.Field)
	case errors.Is(err, ErrOrganizationContext):
		apphttp.WriteError(w, http.StatusBadRequest, "ORGANIZATION_CONTEXT_REQUIRED", "Organization context is required.", "")
	case errors.Is(err, ErrInsufficientScope):
		apphttp.WriteError(w, http.StatusForbidden, "INSUFFICIENT_SCOPE", "Insufficient permissions.", "")
	case errors.Is(err, ErrReadFailed):
		apphttp.WriteError(w, http.StatusServiceUnavailable, "AUDIT_DEPENDENCY_UNAVAILABLE", "Audit service is unavailable.", "")
	default:
		apphttp.WriteError(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "An unexpected error occurred.", "")
	}
}
