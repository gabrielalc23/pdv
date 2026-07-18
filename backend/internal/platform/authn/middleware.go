package authn

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gabrielalc23/pdv/internal/platform/authcontext"
	"github.com/gabrielalc23/pdv/internal/platform/clock"
	platformhttp "github.com/gabrielalc23/pdv/internal/platform/http"
	jwt "github.com/gabrielalc23/pdv/internal/platform/token/jwt"
)

type Middleware struct {
	validator     *jwt.Validator
	loader        *sessionLoader
	persistence   *persistenceStore
	cache         *sessionCache
	touchThrottle *touchThrottle
	clock         clock.Clock
}

func NewMiddleware(
	validator *jwt.Validator,
	persistence *persistenceStore,
	cache *sessionCache,
	touchThrottle *touchThrottle,
	clk clock.Clock,
) *Middleware {
	loader := newSessionLoader(persistence, cache, clk.Now)
	return &Middleware{
		validator:     validator,
		loader:        loader,
		persistence:   persistence,
		cache:         cache,
		touchThrottle: touchThrottle,
		clock:         clk,
	}
}

func (m *Middleware) RequireAccessToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if strings.Count(r.Header.Get("Authorization"), ",") > 0 {
			writeAuthError(w, 401, CodeAccessTokenInvalid, "multiple authorization headers")
			return
		}

		tokenStr, err := validateBearer(authHeader)
		if err != nil {
			authErr := mapErr(err)
			writeAuthError(w, authErr.HTTPStatus, authErr.Code, authErr.Message)
			return
		}

		claims, err := m.validator.Validate(tokenStr)
		if err != nil {
			authErr := mapErr(jwtErrToAuthnErr(err))
			writeAuthError(w, authErr.HTTPStatus, authErr.Code, authErr.Message)
			return
		}

		sessionID, err := claims.ParseSessionID()
		if err != nil {
			writeAuthError(w, 401, CodeAccessTokenInvalid, "invalid session id in token")
			return
		}
		if !sessionID.Valid {
			writeAuthError(w, 401, CodeAccessTokenInvalid, "missing session id in token")
			return
		}

		userID, err := claims.ParseSubject()
		if err != nil || !userID.Valid {
			writeAuthError(w, 401, CodeAccessTokenInvalid, "invalid subject in token")
			return
		}

		state, err := m.loader.load(r.Context(), sessionID)
		if err != nil {
			if errors.Is(err, ErrDependencyUnavailable) {
				writeAuthError(w, 503, CodeDependencyUnavailable, "authentication service unavailable")
				return
			}
			authErr := mapErr(err)
			writeAuthError(w, authErr.HTTPStatus, authErr.Code, authErr.Message)
			return
		}

		if state.UserID != userID {
			writeAuthError(w, 401, CodeSessionRevoked, "session does not match user")
			return
		}

		if state.ClientID != "" && state.ClientID != claims.ClientID {
			writeAuthError(w, 401, CodeSessionRevoked, "client id mismatch")
			return
		}

		now := m.clock.Now()
		if err := validateSessionStatus(state, now); err != nil {
			authErr := mapErr(err)
			writeAuthError(w, authErr.HTTPStatus, authErr.Code, authErr.Message)
			return
		}

		if err := validateUserStatus(state.UserStatus); err != nil {
			authErr := mapErr(err)
			writeAuthError(w, authErr.HTTPStatus, authErr.Code, authErr.Message)
			return
		}

		if err := validateOrgStatus(state.OrganizationStatus); err != nil {
			authErr := mapErr(err)
			writeAuthError(w, authErr.HTTPStatus, authErr.Code, authErr.Message)
			return
		}

		if err := validateMembershipStatus(state.MembershipStatus); err != nil {
			authErr := mapErr(err)
			writeAuthError(w, authErr.HTTPStatus, authErr.Code, authErr.Message)
			return
		}

		if err := validateStoreStatus(state.StoreStatus); err != nil {
			authErr := mapErr(err)
			writeAuthError(w, authErr.HTTPStatus, authErr.Code, authErr.Message)
			return
		}

		if err := validateContext(claims, state); err != nil {
			authErr := mapErr(err)
			writeAuthError(w, authErr.HTTPStatus, authErr.Code, authErr.Message)
			return
		}

		if err := validateVersions(claims, state); err != nil {
			authErr := mapErr(err)
			writeAuthError(w, authErr.HTTPStatus, authErr.Code, authErr.Message)
			return
		}

		principal := buildPrincipal(claims, state)
		ctx := authcontext.SetPrincipal(r.Context(), principal)

		next.ServeHTTP(w, r.WithContext(ctx))

		m.tryTouch(r.Context(), state)
	})
}

func (m *Middleware) tryTouch(ctx context.Context, state sessionState) {
	if !m.touchThrottle.tryTouch(ctx, state.SessionID) {
		return
	}

	idleExpiresAt := state.IdleExpiresAt
	if idleExpiresAt.IsZero() {
		return
	}

	go func() {
		if err := m.persistence.touchSession(ctx, state.SessionID, state.UserID, idleExpiresAt); err != nil {
			slog.Warn("authn: failed to touch session", "error", err)
		}
	}()

	go func() {
		m.cache.setUserPasswordVersion(ctx, state.UserID, state.PasswordVersion)
	}()
}

func buildPrincipal(claims *jwt.Claims, state sessionState) authcontext.Principal {
	p := authcontext.Principal{
		ClientID:        claims.ClientID,
		RoleKeys:        copySlice(claims.Roles),
		PasswordVersion: claims.PV,
		AuthTime:        time.Unix(claims.AuthTime, 0),
	}

	if sessionID, err := claims.ParseSessionID(); err == nil {
		p.SessionID = sessionID
	}
	if userID, err := claims.ParseSubject(); err == nil {
		p.UserID = userID
	}
	if jti, err := claims.ParseJTI(); err == nil {
		p.TokenID = jti
	}

	switch claims.Ctx {
	case jwt.ContextIdentity:
		p.ContextKind = authcontext.ContextIdentity
	case jwt.ContextOrganization:
		p.ContextKind = authcontext.ContextOrganization
		if orgID, err := claims.ParseOrgID(); err == nil {
			p.OrganizationID = orgID
		}
		if memID, err := claims.ParseMembershipID(); err == nil {
			p.MembershipID = memID
		}
		p.OrgAuthzVersion = copyInt64Ptr(claims.OAV)
		p.MemberAuthzVersion = copyInt64Ptr(claims.MAV)
	case jwt.ContextStore:
		p.ContextKind = authcontext.ContextStore
		if orgID, err := claims.ParseOrgID(); err == nil {
			p.OrganizationID = orgID
		}
		if memID, err := claims.ParseMembershipID(); err == nil {
			p.MembershipID = memID
		}
		if storeID, err := claims.ParseStoreID(); err == nil {
			p.StoreID = storeID
		}
		p.OrgAuthzVersion = copyInt64Ptr(claims.OAV)
		p.MemberAuthzVersion = copyInt64Ptr(claims.MAV)
	}

	if claims.Scope != "" {
		scopes := strings.Split(claims.Scope, " ")
		scopeSet := make(authcontext.ScopeSet, len(scopes))
		for _, s := range scopes {
			if s != "" {
				scopeSet[authcontext.Scope(s)] = struct{}{}
			}
		}
		p.Scopes = scopeSet
	}

	return p
}

func jwtErrToAuthnErr(err error) error {
	switch {
	case errors.Is(err, jwt.ErrTokenExpired):
		return ErrAccessTokenExpired
	case errors.Is(err, jwt.ErrTokenSignature), errors.Is(err, jwt.ErrTokenSize),
		errors.Is(err, jwt.ErrTokenAlgorithm), errors.Is(err, jwt.ErrTokenTyp),
		errors.Is(err, jwt.ErrTokenKID), errors.Is(err, jwt.ErrTokenIssuer),
		errors.Is(err, jwt.ErrTokenAudience), errors.Is(err, jwt.ErrClaimsInvalid),
		errors.Is(err, jwt.ErrClaimsIncoherent):
		return ErrAccessTokenInvalid
	default:
		return ErrAccessTokenInvalid
	}
}

func writeAuthError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")

	if status == 401 {
		w.Header().Set("WWW-Authenticate", "Bearer")
	}

	platformhttp.WriteJSON(w, status, platformhttp.ErrorResponse{
		Error: platformhttp.ErrorDetails{
			Code:    code,
			Message: message,
		},
	})
}

func copySlice(s []string) []string {
	if s == nil {
		return nil
	}
	cp := make([]string, len(s))
	copy(cp, s)
	return cp
}

func copyInt64Ptr(p *int64) *int64 {
	if p == nil {
		return nil
	}
	v := *p
	return &v
}
