package auth

import (
	"context"
	"errors"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gabrielalc23/pdv/internal/platform/authcontext"
	"github.com/gabrielalc23/pdv/internal/platform/cookie"
	"github.com/gabrielalc23/pdv/internal/platform/csrf"
	platformhttp "github.com/gabrielalc23/pdv/internal/platform/http"
	"github.com/gabrielalc23/pdv/internal/platform/ratelimit"
	"github.com/gabrielalc23/pdv/internal/platform/requestmeta"
	jwt "github.com/gabrielalc23/pdv/internal/platform/token/jwt"
	"github.com/gabrielalc23/pdv/internal/sessions"
)

type Handler struct {
	service     *Service
	cookies     *cookie.Manager
	csrf        *csrf.Manager
	limiter     ratelimit.Limiter
	rateKey     []byte
	requestMeta *requestmeta.Resolver
	validator   *jwt.Validator
}

func NewHandler(service *Service, cookies *cookie.Manager, csrfManager *csrf.Manager, limiter ratelimit.Limiter, rateKey []byte, resolver *requestmeta.Resolver, validator *jwt.Validator) *Handler {
	return &Handler{service: service, cookies: cookies, csrf: csrfManager, limiter: limiter, rateKey: rateKey, requestMeta: resolver, validator: validator}
}

func (h *Handler) CSRF(w http.ResponseWriter, r *http.Request) {
	token, err := h.csrf.Generate(csrf.BindingPreauth, "preauth")
	if err != nil {
		writeError(w, ErrDependencyUnavailable, false)
		return
	}
	h.cookies.SetCSRFCookie(w, token, time.Now().Add(24*time.Hour))
	setSensitiveHeaders(w)
	platformhttp.WriteJSON(w, http.StatusOK, CSRFResponse{CSRFToken: token})
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if !h.service.RegistrationEnabled() {
		writeError(w, ErrRegistrationDisabled, false)
		return
	}
	meta := h.metadata(r)
	if !h.allow(w, r.Context(), "register:ip:"+ratelimit.Fingerprint(h.rateKey, meta.ClientIP), 3, time.Hour) {
		return
	}
	if !h.validateCSRF(w, r, "") {
		return
	}
	var input RegisterRequest
	if err := platformhttp.DecodeJSONBody(r, &input); err != nil {
		writeError(w, ErrInvalidRequest, false)
		return
	}
	result, err := h.service.Register(r.Context(), input, meta)
	if err != nil {
		writeError(w, err, false)
		return
	}
	setSensitiveHeaders(w)
	if result.VerificationRequired {
		platformhttp.WriteJSON(w, http.StatusAccepted, VerificationRequiredResponse{Status: "VERIFICATION_REQUIRED"})
		return
	}
	if err := h.setAuthCookies(w, *result.Auth); err != nil {
		writeError(w, err, false)
		return
	}
	platformhttp.WriteJSON(w, http.StatusCreated, result.Auth.Response)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if !h.validateCSRF(w, r, "") {
		return
	}
	meta := h.metadata(r)
	if !h.allow(w, r.Context(), "login:ip:"+ratelimit.Fingerprint(h.rateKey, meta.ClientIP), 5, time.Minute) {
		return
	}
	var input LoginRequest
	if err := platformhttp.DecodeJSONBody(r, &input); err != nil {
		writeError(w, ErrInvalidRequest, false)
		return
	}
	normalized := normalizeEmail(input.Email)
	if !h.allow(w, r.Context(), "login:email:"+ratelimit.Fingerprint(h.rateKey, normalized), 10, 15*time.Minute) {
		return
	}
	result, err := h.service.Login(r.Context(), input, meta)
	if err != nil {
		writeError(w, err, false)
		return
	}
	if err := h.setAuthCookies(w, result); err != nil {
		writeError(w, err, false)
		return
	}
	setSensitiveHeaders(w)
	platformhttp.WriteJSON(w, http.StatusOK, result.Response)
}

func (h *Handler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	if !h.validateCSRF(w, r, "") {
		return
	}
	var input EmailActionRequest
	if err := decodeSmallJSON(w, r, &input); err != nil {
		writeError(w, ErrInvalidRequest, false)
		return
	}
	meta := h.metadata(r)
	normalizedEmail := normalizeEmail(input.Email)
	if !h.allow(w, r.Context(), "password-forgot:ip:"+ratelimit.Fingerprint(h.rateKey, meta.ClientIP), 10, time.Hour) ||
		!h.allow(w, r.Context(), "password-forgot:email:"+ratelimit.Fingerprint(h.rateKey, normalizedEmail), 3, time.Hour) {
		return
	}
	if err := h.service.ForgotPassword(r.Context(), input, meta); err != nil {
		writeError(w, err, false)
		return
	}
	setSensitiveHeaders(w)
	platformhttp.WriteJSON(w, http.StatusAccepted, AcceptedResponse{Status: "ACCEPTED"})
}

func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	if !h.validateCSRF(w, r, "") {
		return
	}
	var input PasswordResetRequest
	if err := decodeSmallJSON(w, r, &input); err != nil {
		writeError(w, ErrInvalidRequest, false)
		return
	}
	if err := h.service.ResetPassword(r.Context(), input, h.metadata(r)); err != nil {
		writeError(w, err, false)
		return
	}
	h.cookies.ClearAuthCookies(w)
	setSensitiveHeaders(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	if !h.validateCSRF(w, r, "") {
		return
	}
	var input EmailActionRequest
	if err := decodeSmallJSON(w, r, &input); err != nil {
		writeError(w, ErrInvalidRequest, false)
		return
	}
	meta := h.metadata(r)
	normalizedEmail := normalizeEmail(input.Email)
	if !h.allow(w, r.Context(), "resend-verification:ip:"+ratelimit.Fingerprint(h.rateKey, meta.ClientIP), 10, time.Hour) ||
		!h.allow(w, r.Context(), "resend-verification:email:"+ratelimit.Fingerprint(h.rateKey, normalizedEmail), 3, time.Hour) {
		return
	}
	if err := h.service.ResendVerification(r.Context(), input, meta); err != nil {
		writeError(w, err, false)
		return
	}
	setSensitiveHeaders(w)
	platformhttp.WriteJSON(w, http.StatusAccepted, AcceptedResponse{Status: "ACCEPTED"})
}

func (h *Handler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	if !h.validateCSRF(w, r, "") {
		return
	}
	var input EmailVerifyRequest
	if err := decodeSmallJSON(w, r, &input); err != nil {
		writeError(w, ErrInvalidRequest, false)
		return
	}
	if err := h.service.VerifyEmail(r.Context(), input, h.metadata(r)); err != nil {
		writeError(w, err, false)
		return
	}
	setSensitiveHeaders(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	if err := requireEmptyBody(r); err != nil {
		writeError(w, ErrInvalidRequest, false)
		return
	}
	refreshCookie, err := r.Cookie(h.cookies.RefreshCookieName())
	if err != nil || refreshCookie.Value == "" {
		writeError(w, sessions.ErrRefreshTokenMissing, false)
		return
	}
	if err := h.csrf.CheckOrigin(r); err != nil {
		writeCSRFError(w, "CSRF_INVALID", "Origem inválida.")
		return
	}
	if err := h.csrf.CheckFetchMetadata(r); err != nil {
		writeCSRFError(w, "CSRF_INVALID", "Requisição cross-site rejeitada.")
		return
	}
	meta := h.metadata(r)
	if !h.allow(w, r.Context(), "refresh:ip:"+ratelimit.Fingerprint(h.rateKey, meta.ClientIP), 60, time.Minute) {
		return
	}
	sessionID, err := h.service.ResolveRefreshSessionID(r.Context(), refreshCookie.Value)
	if err != nil {
		if errors.Is(err, sessions.ErrRefreshTokenReused) {
			h.cookies.ClearAuthCookies(w)
		}
		writeError(w, err, false)
		return
	}
	if !h.validateCSRF(w, r, uuidString(sessionID)) {
		return
	}
	if !h.allow(w, r.Context(), "refresh:session:"+ratelimit.Fingerprint(h.rateKey, uuidString(sessionID)), 30, time.Minute) {
		return
	}
	result, err := h.service.Refresh(r.Context(), refreshCookie.Value, meta)
	if err != nil {
		if errors.Is(err, sessions.ErrRefreshTokenReused) {
			h.cookies.ClearAuthCookies(w)
		}
		writeError(w, err, false)
		return
	}
	if err := h.setAuthCookies(w, result); err != nil {
		writeError(w, err, false)
		return
	}
	setSensitiveHeaders(w)
	platformhttp.WriteJSON(w, http.StatusOK, result.Response)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if err := requireEmptyBody(r); err != nil {
		writeError(w, ErrInvalidRequest, false)
		return
	}
	defer h.cookies.ClearAuthCookies(w)
	meta := h.metadata(r)
	if userID, sessionID, ok := h.optionalBearer(r); ok {
		if err := h.service.Logout(r.Context(), userID, sessionID, meta); err != nil {
			writeError(w, err, true)
			return
		}
		setSensitiveHeaders(w)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	refreshCookie, cookieErr := r.Cookie(h.cookies.RefreshCookieName())
	if cookieErr == nil && refreshCookie.Value != "" {
		sessionID, err := h.service.ResolveRefreshSessionID(r.Context(), refreshCookie.Value)
		if err == nil {
			if !h.validateCSRF(w, r, uuidString(sessionID)) {
				return
			}
			state, stateErr := h.service.loadState(r.Context(), sessionID)
			if stateErr == nil {
				if err := h.service.Logout(r.Context(), state.UserID, sessionID, meta); err != nil {
					writeError(w, err, false)
					return
				}
			} else if !errors.Is(stateErr, sessions.ErrSessionRevoked) && !errors.Is(stateErr, sessions.ErrSessionExpired) {
				writeError(w, stateErr, false)
				return
			}
		}
	}
	setSensitiveHeaders(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) LogoutAll(w http.ResponseWriter, r *http.Request) {
	if err := requireEmptyBody(r); err != nil {
		writeError(w, ErrInvalidRequest, true)
		return
	}
	principal, ok := authcontext.PrincipalFromContext(r.Context())
	if !ok {
		writeError(w, ErrInvalidCredentials, true)
		return
	}
	if err := h.service.LogoutAll(r.Context(), principal.UserID, principal.SessionID, h.metadata(r)); err != nil {
		writeError(w, err, true)
		return
	}
	h.cookies.ClearAuthCookies(w)
	setSensitiveHeaders(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) SwitchContext(w http.ResponseWriter, r *http.Request) {
	principal, ok := authcontext.PrincipalFromContext(r.Context())
	if !ok {
		writeError(w, ErrInvalidCredentials, true)
		return
	}
	var input ContextRequest
	if err := platformhttp.DecodeJSONBody(r, &input); err != nil {
		writeError(w, ErrInvalidRequest, true)
		return
	}
	result, err := h.service.SwitchContext(r.Context(), principal.UserID, principal.SessionID, input, principal.AuthTime, h.metadata(r))
	if err != nil {
		writeError(w, err, true)
		return
	}
	setSensitiveHeaders(w)
	platformhttp.WriteJSON(w, http.StatusOK, result.Response)
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	principal, ok := authcontext.PrincipalFromContext(r.Context())
	if !ok {
		writeError(w, ErrInvalidCredentials, true)
		return
	}
	response, err := h.service.Me(r.Context(), principal.SessionID)
	if err != nil {
		writeError(w, err, true)
		return
	}
	setSensitiveHeaders(w)
	platformhttp.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	principal, ok := authcontext.PrincipalFromContext(r.Context())
	if !ok {
		writeError(w, ErrInvalidCredentials, true)
		return
	}
	var input UpdateMeRequest
	if err := decodeSmallJSON(w, r, &input); err != nil {
		writeError(w, ErrInvalidRequest, true)
		return
	}
	response, err := h.service.UpdateMe(r.Context(), principal.UserID, input)
	if err != nil {
		writeError(w, err, true)
		return
	}
	setSensitiveHeaders(w)
	platformhttp.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	principal, ok := authcontext.PrincipalFromContext(r.Context())
	if !ok {
		writeError(w, ErrInvalidCredentials, true)
		return
	}
	var input ChangePasswordRequest
	if err := decodeSmallJSON(w, r, &input); err != nil {
		writeError(w, ErrInvalidRequest, true)
		return
	}
	if err := h.service.ChangePassword(r.Context(), principal.UserID, principal.SessionID, input, h.metadata(r)); err != nil {
		writeError(w, err, true)
		return
	}
	h.cookies.ClearAuthCookies(w)
	setSensitiveHeaders(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Sessions(w http.ResponseWriter, r *http.Request) {
	principal, ok := authcontext.PrincipalFromContext(r.Context())
	if !ok {
		writeError(w, ErrInvalidCredentials, true)
		return
	}
	response, err := h.service.ListSessions(r.Context(), principal.UserID, principal.SessionID)
	if err != nil {
		writeError(w, err, true)
		return
	}
	setSensitiveHeaders(w)
	platformhttp.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) RevokeSession(w http.ResponseWriter, r *http.Request) {
	principal, ok := authcontext.PrincipalFromContext(r.Context())
	if !ok {
		writeError(w, ErrInvalidCredentials, true)
		return
	}
	raw := chi.URLParam(r, "sessionId")
	var target pgtype.UUID
	if err := target.Scan(raw); err != nil || !target.Valid || uuidString(target) != strings.ToLower(raw) {
		writeError(w, validationError("sessionId", "UUID inválido."), true)
		return
	}
	clear, err := h.service.RevokeOwnSession(r.Context(), principal.UserID, principal.SessionID, target, h.metadata(r))
	if err != nil {
		writeError(w, err, true)
		return
	}
	if clear {
		h.cookies.ClearAuthCookies(w)
	}
	setSensitiveHeaders(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) setAuthCookies(w http.ResponseWriter, result AuthResult) error {
	csrfToken, err := h.csrf.Generate(csrf.BindingSessionID, result.Response.Session.ID)
	if err != nil {
		return ErrDependencyUnavailable
	}
	h.cookies.SetRefreshCookie(w, result.RawRefreshToken, result.RefreshExpires)
	h.cookies.SetCSRFCookie(w, csrfToken, result.RefreshExpires)
	return nil
}

func (h *Handler) validateCSRF(w http.ResponseWriter, r *http.Request, sessionID string) bool {
	if err := h.csrf.CheckOrigin(r); err != nil {
		writeCSRFError(w, "CSRF_INVALID", "Origem inválida.")
		return false
	}
	if err := h.csrf.CheckFetchMetadata(r); err != nil {
		writeCSRFError(w, "CSRF_INVALID", "Requisição cross-site rejeitada.")
		return false
	}
	cookieValue := ""
	if c, err := r.Cookie(h.cookies.CSRFCookieName()); err == nil {
		cookieValue = c.Value
	}
	var err error
	if sessionID == "" {
		err = h.csrf.ValidateRequest(r, cookieValue)
	} else {
		err = h.csrf.ValidateRequestWithSession(r, cookieValue, sessionID)
	}
	if err != nil {
		code := "CSRF_INVALID"
		message := "Token CSRF inválido."
		if errors.Is(err, csrf.ErrTokenMissing) {
			code, message = "CSRF_TOKEN_MISSING", "Token CSRF ausente."
		}
		writeCSRFError(w, code, message)
		return false
	}
	return true
}

func writeCSRFError(w http.ResponseWriter, code, message string) {
	setSensitiveHeaders(w)
	platformhttp.WriteError(w, http.StatusForbidden, code, message, "")
}

func (h *Handler) allow(w http.ResponseWriter, ctx context.Context, key string, limit int, window time.Duration) bool {
	result, err := h.limiter.Allow(ctx, "ratelimit:"+key, limit, window)
	if err != nil {
		writeError(w, ErrDependencyUnavailable, false)
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
	platformhttp.WriteError(w, http.StatusTooManyRequests, "RATE_LIMITED", "Muitas tentativas. Tente novamente mais tarde.", "")
	return false
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

func (h *Handler) optionalBearer(r *http.Request) (pgtype.UUID, pgtype.UUID, bool) {
	header := r.Header.Get("Authorization")
	if !strings.HasPrefix(header, "Bearer ") || strings.Contains(header, ",") || h.validator == nil {
		return pgtype.UUID{}, pgtype.UUID{}, false
	}
	claims, err := h.validator.Validate(strings.TrimSpace(strings.TrimPrefix(header, "Bearer ")))
	if err != nil {
		return pgtype.UUID{}, pgtype.UUID{}, false
	}
	userID, userErr := claims.ParseSubject()
	sessionID, sessionErr := claims.ParseSessionID()
	return userID, sessionID, userErr == nil && sessionErr == nil && userID.Valid && sessionID.Valid
}

func requireEmptyBody(r *http.Request) error {
	data, err := io.ReadAll(io.LimitReader(r.Body, 1025))
	if err != nil {
		return err
	}
	if len(data) > 1024 {
		return ErrInvalidRequest
	}
	if strings.TrimSpace(string(data)) != "" {
		return ErrInvalidRequest
	}
	return nil
}

func decodeSmallJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 8<<10)
	return platformhttp.DecodeJSONBody(r, dst)
}
