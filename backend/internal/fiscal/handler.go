package fiscal

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

func (h *Handler) GetFiscalDocument(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.resolveActor(w, r)
	if !ok {
		return
	}
	result, err := h.service.GetBySaleID(r.Context(), actor, chi.URLParam(r, "id"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	apphttp.WriteJSON(w, http.StatusOK, result)
}
