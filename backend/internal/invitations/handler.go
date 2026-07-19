package invitations

import (
	"context"
	"errors"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/gabrielalc23/pdv/internal/auth"
	"github.com/gabrielalc23/pdv/internal/platform/authcontext"
	"github.com/gabrielalc23/pdv/internal/platform/cookie"
	"github.com/gabrielalc23/pdv/internal/platform/csrf"
	apphttp "github.com/gabrielalc23/pdv/internal/platform/http"
	"github.com/gabrielalc23/pdv/internal/platform/ratelimit"
	"github.com/gabrielalc23/pdv/internal/platform/requestmeta"
)

type Handler struct {
	service     *Service
	cookies     *cookie.Manager
	csrf        *csrf.Manager
	limiter     ratelimit.Limiter
	rateKey     []byte
	requestMeta *requestmeta.Resolver
}

func NewHandler(service *Service, cookies *cookie.Manager, csrfManager *csrf.Manager, limiter ratelimit.Limiter, rateKey []byte, resolver *requestmeta.Resolver) *Handler {
	return &Handler{service: service, cookies: cookies, csrf: csrfManager, limiter: limiter, rateKey: append([]byte(nil), rateKey...), requestMeta: resolver}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	principal, ok := h.principal(w, r)
	if !ok {
		return
	}
	page, err := optionalInt(r.URL.Query().Get("page"))
	if err != nil {
		h.writeError(w, validationError("page", "Paginação inválida."), true)
		return
	}
	pageSize, err := optionalInt(r.URL.Query().Get("pageSize"))
	if err != nil {
		h.writeError(w, validationError("pageSize", "Paginação inválida."), true)
		return
	}
	response, err := h.service.List(r.Context(), principal, ListInput{Status: r.URL.Query().Get("status"), Email: r.URL.Query().Get("email"), CreatedFrom: r.URL.Query().Get("createdFrom"), CreatedTo: r.URL.Query().Get("createdTo"), Page: page, PageSize: pageSize})
	if err != nil {
		h.writeError(w, err, true)
		return
	}
	apphttp.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	principal, ok := h.principal(w, r)
	if !ok {
		return
	}
	var input CreateInput
	if err := apphttp.DecodeJSONBody(r, &input); err != nil {
		h.writeError(w, ErrInvalidRequest, true)
		return
	}
	response, err := h.service.Create(r.Context(), principal, input, h.metadata(r))
	if err != nil {
		h.writeError(w, err, true)
		return
	}
	apphttp.WriteJSON(w, http.StatusCreated, response)
}

func (h *Handler) Resend(w http.ResponseWriter, r *http.Request) {
	principal, ok := h.principal(w, r)
	if !ok {
		return
	}
	response, err := h.service.Resend(r.Context(), principal, chi.URLParam(r, "id"), h.metadata(r))
	if err != nil {
		h.writeError(w, err, true)
		return
	}
	apphttp.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) Revoke(w http.ResponseWriter, r *http.Request) {
	principal, ok := h.principal(w, r)
	if !ok {
		return
	}
	if err := h.service.Revoke(r.Context(), principal, chi.URLParam(r, "id"), h.metadata(r)); err != nil {
		h.writeError(w, err, true)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Inspect(w http.ResponseWriter, r *http.Request) {
	if !h.allow(w, r.Context(), "invitation:ip:"+ratelimit.Fingerprint(h.rateKey, h.metadata(r).ClientIP)) {
		return
	}
	var input InspectInput
	if err := apphttp.DecodeJSONBody(r, &input); err != nil {
		h.writeError(w, ErrInvitationNotFound, false)
		return
	}
	response, err := h.service.Inspect(r.Context(), input.Token)
	if err != nil {
		h.writeError(w, err, false)
		return
	}
	setSensitiveHeaders(w)
	apphttp.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) Accept(w http.ResponseWriter, r *http.Request) {
	if !h.allow(w, r.Context(), "invitation:ip:"+ratelimit.Fingerprint(h.rateKey, h.metadata(r).ClientIP)) {
		return
	}
	principal, authenticated := authcontext.PrincipalFromContext(r.Context())
	if authenticated {
		if !h.validateCSRF(w, r, principal.SessionID.String()) {
			return
		}
	} else if !h.validateCSRF(w, r, "") {
		return
	}
	var input AcceptInput
	if err := apphttp.DecodeJSONBody(r, &input); err != nil {
		h.writeError(w, ErrInvalidRequest, authenticated)
		return
	}
	var actor *authcontext.Principal
	if authenticated {
		actor = &principal
	}
	result, err := h.service.Accept(r.Context(), actor, input, h.metadata(r))
	if err != nil {
		h.writeError(w, err, authenticated)
		return
	}
	setSensitiveHeaders(w)
	if result.Auth != nil {
		if err := h.setAuthCookies(w, *result.Auth); err != nil {
			h.writeError(w, err, false)
			return
		}
		apphttp.WriteJSON(w, http.StatusOK, result.Auth.Response)
		return
	}
	apphttp.WriteJSON(w, http.StatusOK, result.Membership)
}

func (h *Handler) principal(w http.ResponseWriter, r *http.Request) (authcontext.Principal, bool) {
	principal, ok := authcontext.PrincipalFromContext(r.Context())
	if !ok {
		apphttp.WriteError(w, http.StatusUnauthorized, "ACCESS_TOKEN_MISSING", "Authentication is required.", "")
		return authcontext.Principal{}, false
	}
	return principal, true
}

func (h *Handler) metadata(r *http.Request) requestmeta.RequestMetadata {
	if meta, ok := requestmeta.FromContext(r.Context()); ok {
		return meta
	}
	if h.requestMeta != nil {
		return h.requestMeta.Extract(r)
	}
	return requestmeta.RequestMetadata{UserAgent: r.UserAgent()}
}

func (h *Handler) allow(w http.ResponseWriter, ctx context.Context, key string) bool {
	result, err := h.limiter.Allow(ctx, "ratelimit:"+key, 10, time.Minute)
	if err != nil {
		h.writeError(w, ErrDependencyUnavailable, false)
		return false
	}
	if result.Allowed {
		return true
	}
	retry := int(math.Ceil(time.Until(result.ResetAt).Seconds()))
	if retry < 1 {
		retry = 1
	}
	w.Header().Set("Retry-After", strconv.Itoa(retry))
	setSensitiveHeaders(w)
	apphttp.WriteError(w, http.StatusTooManyRequests, "RATE_LIMITED", "Muitas tentativas. Tente novamente mais tarde.", "")
	return false
}

func (h *Handler) validateCSRF(w http.ResponseWriter, r *http.Request, sessionID string) bool {
	if h.csrf.CheckOrigin(r) != nil || h.csrf.CheckFetchMetadata(r) != nil {
		setSensitiveHeaders(w)
		apphttp.WriteError(w, http.StatusForbidden, "CSRF_INVALID", "Token CSRF inválido.", "")
		return false
	}
	cookieValue := ""
	if csrfCookie, err := r.Cookie(h.cookies.CSRFCookieName()); err == nil {
		cookieValue = csrfCookie.Value
	}
	var err error
	if sessionID == "" {
		err = h.csrf.ValidateRequest(r, cookieValue)
	} else {
		err = h.csrf.ValidateRequestWithSession(r, cookieValue, sessionID)
	}
	if err == nil {
		return true
	}
	code := "CSRF_INVALID"
	if errors.Is(err, csrf.ErrTokenMissing) {
		code = "CSRF_TOKEN_MISSING"
	}
	setSensitiveHeaders(w)
	apphttp.WriteError(w, http.StatusForbidden, code, "Token CSRF inválido.", "")
	return false
}

func (h *Handler) setAuthCookies(w http.ResponseWriter, result auth.AuthResult) error {
	token, err := h.csrf.Generate(csrf.BindingSessionID, result.Response.Session.ID)
	if err != nil {
		return ErrDependencyUnavailable
	}
	h.cookies.SetRefreshCookie(w, result.RawRefreshToken, result.RefreshExpires)
	h.cookies.SetCSRFCookie(w, token, result.RefreshExpires)
	return nil
}

func optionalInt(value string) (int, error) {
	if value == "" {
		return 0, nil
	}
	return strconv.Atoi(value)
}

func setSensitiveHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
}
