package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/gabrielalc23/pdv/internal/catalog"
	"github.com/gabrielalc23/pdv/internal/categories"
	"github.com/gabrielalc23/pdv/internal/checkout"
	"github.com/gabrielalc23/pdv/internal/fiscal"
	"github.com/gabrielalc23/pdv/internal/inventory"
	"github.com/gabrielalc23/pdv/internal/payments"
	"github.com/gabrielalc23/pdv/internal/platform/authn"
	"github.com/gabrielalc23/pdv/internal/platform/authz"
	"github.com/gabrielalc23/pdv/internal/platform/clock"
	"github.com/gabrielalc23/pdv/internal/platform/cookie"
	"github.com/gabrielalc23/pdv/internal/platform/csrf"
	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/gabrielalc23/pdv/internal/platform/password"
	"github.com/gabrielalc23/pdv/internal/platform/ratelimit"
	"github.com/gabrielalc23/pdv/internal/platform/requestmeta"
	"github.com/gabrielalc23/pdv/internal/platform/tenancy"
	jwt "github.com/gabrielalc23/pdv/internal/platform/token/jwt"
	"github.com/gabrielalc23/pdv/internal/platform/valkey"
	"github.com/gabrielalc23/pdv/internal/products"
	"github.com/gabrielalc23/pdv/internal/receipt"
	"github.com/gabrielalc23/pdv/internal/sales"
	"github.com/gabrielalc23/pdv/internal/sessions"
)

type Dependencies struct {
	Store              *database.PostgresStore
	Valkey             *valkey.Client
	PasswordHasher     password.Hasher
	PasswordPolicy     password.Policy
	PasswordBlocklist  password.Blocklist
	CookieManager      *cookie.Manager
	CSRFManager        *csrf.Manager
	JWTKeyring         *jwt.Keyring
	JWTIssuer          string
	JWTAudience        string
	AccessTokenTTL     time.Duration
	JWTClockSkew       time.Duration
	RequestMeta        *requestmeta.Resolver
	RateLimiter        ratelimit.Limiter
	RefreshIdleTTL     time.Duration
	SessionAbsoluteTTL time.Duration
	AuthTokenHashKey   []byte
}

type AuthComponents struct {
	AuthN        *authn.Middleware
	AuthZ        authz.Guard
	Sessions     *sessions.Service
	RefreshCodec sessions.RefreshTokenCodec
}

func buildAuthComponents(deps Dependencies, cfg configAuth) *AuthComponents {
	if deps.Store == nil || deps.Valkey == nil || deps.JWTKeyring == nil {
		return nil
	}

	validator := jwt.NewValidator(deps.JWTKeyring, cfg.JWTIssuer, cfg.JWTAudience, cfg.JWTClockSkew)

	persistence := authn.NewPersistenceStore(deps.Store.Queries)
	cache := authn.NewSessionCache(deps.Valkey, cfg.AuthSessionCacheTTL)
	touchThrottle := authn.NewTouchThrottle(deps.Valkey, cfg.AuthSessionTouchInterval)
	clk := clock.RealClock{}

	authnMiddleware := authn.NewMiddleware(validator, persistence, cache, touchThrottle, clk)
	guard := authz.NewGuard()

	refreshCodec := sessions.NewRefreshTokenCodec(deps.AuthTokenHashKey)
	sessionQuerier := sessions.NewStore(deps.Store.Queries)
	sessionProvider := sessions.NewTxProvider(deps.Store)

	sessionSvc := sessions.NewService(
		refreshCodec,
		sessionProvider,
		sessionQuerier,
		sessions.Config{
			RefreshIdleTTL:     deps.RefreshIdleTTL,
			SessionAbsoluteTTL: deps.SessionAbsoluteTTL,
		},
		clk,
	)

	return &AuthComponents{
		AuthN:        authnMiddleware,
		AuthZ:        guard,
		Sessions:     sessionSvc,
		RefreshCodec: refreshCodec,
	}
}

type configAuth struct {
	JWTIssuer                string
	JWTAudience              string
	JWTClockSkew             time.Duration
	AuthSessionCacheTTL      time.Duration
	AuthSessionTouchInterval time.Duration
}

func New(deps Dependencies) http.Handler {
	router := chi.NewRouter()

	router.Use(chimw.RequestID)
	router.Use(chimw.Recoverer)
	router.Use(requestLogger)
	router.Use(securityHeaders)

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	router.Get("/ready", func(w http.ResponseWriter, r *http.Request) {
		if deps.Store == nil || deps.Store.Pool == nil {
			writeReadyError(w, "database not connected")
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		if err := deps.Store.Pool.Ping(ctx); err != nil {
			slog.Error("readiness ping failed", "error", err)
			writeReadyError(w, "database ping failed")
			return
		}
		if deps.Valkey != nil {
			if err := deps.Valkey.Ping(ctx); err != nil {
				slog.Error("readiness valkey ping failed", "error", err)
				writeReadyError(w, "valkey ping failed")
				return
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	if deps.JWTKeyring != nil {
		jwksService := jwt.NewJWKSService(deps.JWTKeyring)
		router.Get("/.well-known/jwks.json", jwksService.ServeHTTP)
	}

	router.Get("/auth/csrf", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "use GET")
			return
		}
		if csrf.IsUnsafeMethod(r.Method) {
			if err := deps.CSRFManager.CheckOrigin(r); err != nil {
				writeError(w, http.StatusForbidden, "CSRF_INVALID", "origin not allowed")
				return
			}
			if err := deps.CSRFManager.CheckFetchMetadata(r); err != nil {
				writeError(w, http.StatusForbidden, "CSRF_INVALID", "cross-site request rejected")
				return
			}
		}

		token, err := deps.CSRFManager.Generate(csrf.BindingPreauth, "preauth")
		if err != nil {
			slog.Error("failed to generate csrf token", "error", err)
			writeError(w, http.StatusInternalServerError, "AUTH_DEPENDENCY_UNAVAILABLE", "failed to generate csrf token")
			return
		}

		deps.CookieManager.SetCSRFCookie(w, token, time.Now().Add(24*time.Hour))

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"csrfToken": token})
	})

	router.Group(func(r chi.Router) {
		r.Use(tenancy.Middleware)

		productStore := products.NewStore(deps.Store.Queries)
		productService := products.NewService(productStore)
		productHandler := products.NewHandler(productService, tenancy.NewContextResolver())
		products.RegisterRoutes(r, productHandler)

		categoryStore := categories.NewStore(deps.Store.Queries)
		categoryService := categories.NewService(categoryStore)
		categoryHandler := categories.NewHandler(categoryService, tenancy.NewContextResolver())
		categories.RegisterRoutes(r, categoryHandler)

		inventoryReadStore := inventory.NewReadStore(deps.Store.Queries)
		inventoryTxManager := inventory.NewTxManager(deps.Store)
		inventoryService := inventory.NewService(inventoryReadStore, inventoryTxManager)
		inventoryHandler := inventory.NewHandler(inventoryService, tenancy.NewContextResolver())
		inventory.RegisterRoutes(r, inventoryHandler)

		catalogStore := catalog.NewStore(deps.Store.Queries)
		catalogService := catalog.NewService(catalogStore)
		catalogHandler := catalog.NewHandler(catalogService, tenancy.NewContextResolver())
		catalog.RegisterRoutes(r, catalogHandler)

		salesReadStore := sales.NewReadStore(deps.Store.Queries)
		salesTxManager := sales.NewTxManager(deps.Store)
		salesService := sales.NewService(salesReadStore, salesTxManager)
		salesHandler := sales.NewHandler(salesService, tenancy.NewContextResolver())
		sales.RegisterRoutes(r, salesHandler)

		fiscalProvider := &fiscal.MockProvider{}

		checkoutTxManager := checkout.NewTxManager(deps.Store)
		checkoutService := checkout.NewService(checkoutTxManager, fiscalProvider)
		checkoutHandler := checkout.NewHandler(checkoutService, tenancy.NewContextResolver())
		checkout.RegisterRoutes(r, checkoutHandler)

		paymentsStore := payments.NewStore(deps.Store.Queries)
		paymentsService := payments.NewService(paymentsStore)
		paymentsHandler := payments.NewHandler(paymentsService, tenancy.NewContextResolver())
		payments.RegisterRoutes(r, paymentsHandler)

		fiscalService := fiscal.NewService(fiscal.NewStore(deps.Store.Queries), fiscalProvider)
		fiscalHandler := fiscal.NewHandler(fiscalService, tenancy.NewContextResolver())
		fiscal.RegisterRoutes(r, fiscalHandler)

		receiptStore := receipt.NewStore(deps.Store.Queries)
		receiptService := receipt.NewService(receiptStore)
		receiptHandler := receipt.NewHandler(receiptService, tenancy.NewContextResolver())
		receipt.RegisterRoutes(r, receiptHandler)
	})

	return router
}

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := chimw.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)
		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", ww.Status(),
			"duration", time.Since(start).String(),
			"request_id", chimw.GetReqID(r.Context()),
		)
	})
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(w, r)
	})
}

func writeReadyError(w http.ResponseWriter, detail string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusServiceUnavailable)
	_, _ = w.Write([]byte(fmt.Sprintf(`{"status":"unavailable","detail":%q}`, detail)))
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}

func NewHTTPServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
}
