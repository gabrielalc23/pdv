package organizations

import (
	"net/http"

	"github.com/gabrielalc23/pdv/internal/platform/authcontext"
	"github.com/gabrielalc23/pdv/internal/platform/authn"
	platformhttp "github.com/gabrielalc23/pdv/internal/platform/http"
	"github.com/gabrielalc23/pdv/internal/platform/requestmeta"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.identityActor(w, r)
	if !ok {
		return
	}
	response, err := h.service.List(r.Context(), actor)
	if err != nil {
		writeError(w, err)
		return
	}
	platformhttp.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) ListStores(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.identityActor(w, r)
	if !ok {
		return
	}
	response, err := h.service.ListStores(r.Context(), actor, r.URL.Query().Get("organizationId"))
	if err != nil {
		writeError(w, err)
		return
	}
	platformhttp.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.identityActor(w, r)
	if !ok {
		return
	}
	var input CreateOrganizationRequest
	if err := platformhttp.DecodeJSONBody(r, &input); err != nil {
		writeError(w, ErrInvalidRequest)
		return
	}
	response, err := h.service.Create(r.Context(), actor, input, requestmeta.MustFromContext(r.Context()))
	if err != nil {
		writeError(w, err)
		return
	}
	platformhttp.WriteJSON(w, http.StatusCreated, response)
}

func (h *Handler) Current(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.organizationActor(w, r)
	if !ok {
		return
	}
	response, err := h.service.Current(r.Context(), actor)
	if err != nil {
		writeError(w, err)
		return
	}
	platformhttp.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.organizationActor(w, r)
	if !ok {
		return
	}
	var input UpdateOrganizationRequest
	if err := platformhttp.DecodeJSONBody(r, &input); err != nil {
		writeError(w, ErrInvalidRequest)
		return
	}
	response, err := h.service.Update(r.Context(), actor, input, requestmeta.MustFromContext(r.Context()))
	if err != nil {
		writeError(w, err)
		return
	}
	platformhttp.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) Archive(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.organizationActor(w, r)
	if !ok {
		return
	}
	var input ArchiveOrganizationRequest
	if err := platformhttp.DecodeJSONBody(r, &input); err != nil {
		writeError(w, ErrInvalidRequest)
		return
	}
	response, err := h.service.Archive(r.Context(), actor, input, requestmeta.MustFromContext(r.Context()))
	if err != nil {
		writeError(w, err)
		return
	}
	platformhttp.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) identityActor(w http.ResponseWriter, r *http.Request) (authn.IdentityActor, bool) {
	principal, ok := authcontext.PrincipalFromContext(r.Context())
	if !ok {
		writeError(w, ErrUnauthenticated)
		return authn.IdentityActor{}, false
	}
	actor, err := authn.IdentityActorFromPrincipal(principal)
	if err != nil {
		writeError(w, ErrUnauthenticated)
		return authn.IdentityActor{}, false
	}
	return actor, true
}

func (h *Handler) organizationActor(w http.ResponseWriter, r *http.Request) (authn.OrganizationActor, bool) {
	principal, ok := authcontext.PrincipalFromContext(r.Context())
	if !ok {
		writeError(w, ErrUnauthenticated)
		return authn.OrganizationActor{}, false
	}
	actor, err := authn.OrganizationActorFromPrincipal(principal)
	if err != nil {
		writeError(w, ErrUnauthenticated)
		return authn.OrganizationActor{}, false
	}
	return actor, true
}
