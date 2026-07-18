package checkout

import (
	"net/http"

	"github.com/gabrielalc23/pdv/internal/platform/tenancy"
	apphttp "github.com/gabrielalc23/pdv/internal/platform/http"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service  *Service
	resolver tenancy.Resolver
}

func NewHandler(service *Service, resolver tenancy.Resolver) *Handler {
	return &Handler{service: service, resolver: resolver}
}

func (h *Handler) resolveActor(w http.ResponseWriter, r *http.Request) (tenancy.ActorScope, bool) {
	if h.resolver == nil {
		apphttp.WriteError(w, http.StatusInternalServerError, "tenant_context_unavailable", "tenant resolver not configured", "")
		return tenancy.ActorScope{}, false
	}
	scope, err := h.resolver.Actor(r.Context())
	if err != nil {
		apphttp.WriteError(w, http.StatusUnauthorized, "tenant_context_unavailable", "actor scope is required", "")
		return tenancy.ActorScope{}, false
	}
	return scope, true
}

func (h *Handler) CheckoutSale(w http.ResponseWriter, r *http.Request) {
	scope, ok := h.resolveActor(w, r)
	if !ok {
		return
	}

	var input CheckoutInput
	if err := apphttp.DecodeJSONBody(r, &input); err != nil {
		apphttp.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must contain a single valid JSON document", "")
		return
	}

	result, err := h.service.Checkout(r.Context(), scope, chi.URLParam(r, "id"), input)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	apphttp.WriteJSON(w, http.StatusOK, result)
}
