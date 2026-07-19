package auth

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/gabrielalc23/pdv/internal/platform/authn"
)

func RegisterRoutes(r chi.Router, h *Handler, middleware *authn.Middleware) {
	r.Get("/auth/csrf", h.CSRF)
	r.Post("/auth/register", h.Register)
	r.Post("/auth/login", h.Login)
	r.Post("/auth/refresh", h.Refresh)
	r.Post("/auth/logout", h.Logout)
	r.Post("/auth/password/forgot", h.ForgotPassword)
	r.Post("/auth/password/reset", h.ResetPassword)
	r.Post("/auth/email/verify", h.VerifyEmail)
	r.Post("/auth/email/resend-verification", h.ResendVerification)
	r.Group(func(protected chi.Router) {
		protected.Use(middleware.RequireAccessToken)
		protected.Post("/auth/logout-all", h.LogoutAll)
		protected.Post("/auth/context", h.SwitchContext)
		protected.Get("/me", h.Me)
		protected.Patch("/me", h.UpdateMe)
		protected.Post("/me/password", h.ChangePassword)
		protected.Get("/me/sessions", h.Sessions)
		protected.Delete("/me/sessions/{sessionId}", h.RevokeSession)
	})
}

func SensitiveHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { setSensitiveHeaders(w); next.ServeHTTP(w, r) })
}
