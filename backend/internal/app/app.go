package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/gabrielalc23/pdv/internal/audit"
	authmodule "github.com/gabrielalc23/pdv/internal/auth"
	"github.com/gabrielalc23/pdv/internal/catalog"
	"github.com/gabrielalc23/pdv/internal/categories"
	"github.com/gabrielalc23/pdv/internal/checkout"
	"github.com/gabrielalc23/pdv/internal/fiscal"
	"github.com/gabrielalc23/pdv/internal/inventory"
	"github.com/gabrielalc23/pdv/internal/invitations"
	"github.com/gabrielalc23/pdv/internal/memberships"
	"github.com/gabrielalc23/pdv/internal/organizations"
	"github.com/gabrielalc23/pdv/internal/payments"
	"github.com/gabrielalc23/pdv/internal/platform/authn"
	"github.com/gabrielalc23/pdv/internal/platform/authz"
	"github.com/gabrielalc23/pdv/internal/platform/clock"
	"github.com/gabrielalc23/pdv/internal/platform/cookie"
	"github.com/gabrielalc23/pdv/internal/platform/csrf"
	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/gabrielalc23/pdv/internal/platform/mailer"
	"github.com/gabrielalc23/pdv/internal/platform/password"
	"github.com/gabrielalc23/pdv/internal/platform/ratelimit"
	"github.com/gabrielalc23/pdv/internal/platform/requestmeta"
	jwt "github.com/gabrielalc23/pdv/internal/platform/token/jwt"
	"github.com/gabrielalc23/pdv/internal/platform/valkey"
	"github.com/gabrielalc23/pdv/internal/products"
	"github.com/gabrielalc23/pdv/internal/receipt"
	"github.com/gabrielalc23/pdv/internal/roles"
	"github.com/gabrielalc23/pdv/internal/sales"
	"github.com/gabrielalc23/pdv/internal/sessions"
	"github.com/gabrielalc23/pdv/internal/stores"
)

type Dependencies struct {
	Store                    *database.PostgresStore
	Valkey                   *valkey.Client
	PasswordHasher           password.Hasher
	PasswordPolicy           password.Policy
	PasswordBlocklist        password.Blocklist
	CookieManager            *cookie.Manager
	CSRFManager              *csrf.Manager
	JWTKeyring               *jwt.Keyring
	JWTIssuer                string
	JWTAudience              string
	AccessTokenTTL           time.Duration
	JWTClockSkew             time.Duration
	RequestMeta              *requestmeta.Resolver
	RateLimiter              ratelimit.Limiter
	RefreshIdleTTL           time.Duration
	SessionAbsoluteTTL       time.Duration
	AuthTokenHashKey         []byte
	RateLimitKeySecret       []byte
	RegistrationEnabled      bool
	TenantCreationEnabled    bool
	RequireVerifiedEmail     bool
	AppPublicURL             string
	AuthSessionCacheTTL      time.Duration
	AuthSessionTouchInterval time.Duration
	Mailer                   mailer.Mailer
	Clock                    clock.Clock
}

type AuthComponents struct {
	AuthN        *authn.Middleware
	AuthZ        authz.Guard
	Sessions     *sessions.Service
	RefreshCodec sessions.RefreshTokenCodec
	Handler      *authmodule.Handler
	Invitations  *invitations.Handler
	Invalidator  *authn.CacheInvalidator
}

func buildAuthComponents(deps Dependencies) *AuthComponents {
	if deps.Store == nil || deps.Valkey == nil || deps.JWTKeyring == nil {
		return nil
	}
	clk := deps.Clock
	if clk == nil {
		clk = clock.RealClock{}
	}

	validator := jwt.NewValidator(deps.JWTKeyring, deps.JWTIssuer, deps.JWTAudience, deps.JWTClockSkew)

	persistence := authn.NewPersistenceStore(deps.Store.Queries)
	cache := authn.NewSessionCache(deps.Valkey, deps.AuthSessionCacheTTL)
	touchThrottle := authn.NewTouchThrottle(deps.Valkey, deps.AuthSessionTouchInterval)

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
	invalidator := authn.NewCacheInvalidator(cache)
	sessionSvc.SetCacheInvalidator(invalidator)
	signer := jwt.NewSigner(deps.JWTKeyring, deps.JWTIssuer, deps.JWTAudience, deps.AccessTokenTTL, deps.JWTClockSkew)
	authService, err := authmodule.NewService(deps.Store, sessionSvc, deps.PasswordHasher, deps.PasswordPolicy, deps.PasswordBlocklist, signer, audit.NewWriter(), deps.Mailer, clk, invalidator, authmodule.Config{
		RegistrationEnabled:  deps.RegistrationEnabled,
		RequireVerifiedEmail: deps.RequireVerifiedEmail,
		AccessTokenTTL:       deps.AccessTokenTTL,
		TokenHashKey:         deps.AuthTokenHashKey,
		RateLimitKey:         deps.RateLimitKeySecret,
		PublicURL:            deps.AppPublicURL,
	})
	if err != nil {
		slog.Error("failed to build auth service", "error", err)
		return nil
	}
	authHandler := authmodule.NewHandler(authService, deps.CookieManager, deps.CSRFManager, deps.RateLimiter, deps.RateLimitKeySecret, deps.RequestMeta, validator)

	invitationService, err := invitations.NewService(
		invitations.NewStore(deps.Store.Queries),
		invitations.NewTxProvider(deps.Store, audit.NewWriter(), sessionSvc),
		deps.PasswordHasher,
		deps.PasswordPolicy,
		deps.PasswordBlocklist,
		signer,
		clk,
		deps.Mailer,
		invalidator,
		invitations.Config{
			TokenHashKey:       deps.AuthTokenHashKey,
			PublicURL:          deps.AppPublicURL,
			InvitationTTL:      7 * 24 * time.Hour,
			OwnerInvitationTTL: 7 * 24 * time.Hour,
			AccessTokenTTL:     deps.AccessTokenTTL,
		},
	)
	if err != nil {
		slog.Error("failed to build invitation service", "error", err)
		return nil
	}
	invitationHandler := invitations.NewHandler(invitationService, deps.CookieManager, deps.CSRFManager, deps.RateLimiter, deps.RateLimitKeySecret, deps.RequestMeta)

	return &AuthComponents{
		AuthN:        authnMiddleware,
		AuthZ:        guard,
		Sessions:     sessionSvc,
		RefreshCodec: refreshCodec,
		Handler:      authHandler,
		Invitations:  invitationHandler,
		Invalidator:  invalidator,
	}
}

func New(deps Dependencies) http.Handler {
	router := chi.NewRouter()

	router.Use(chimw.RequestID)
	if deps.RequestMeta != nil {
		router.Use(deps.RequestMeta.Middleware)
	}
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

	authComponents := buildAuthComponents(deps)
	if authComponents != nil {
		authmodule.RegisterRoutes(router, authComponents.Handler, authComponents.AuthN)
		invitations.RegisterPublicRoutes(router, authComponents.Invitations, authComponents.AuthN)
		registerAdministrationRoutes(router, deps, authComponents)
	}

	if authComponents != nil {
		router.Group(func(business chi.Router) {
			business.Use(authComponents.AuthN.RequireAccessToken)

			productStore := products.NewStore(deps.Store.Queries)
			productService := products.NewService(productStore)
			productHandler := products.NewHandler(productService)
			products.RegisterRoutes(business, productHandler, authComponents.AuthZ)

			categoryStore := categories.NewStore(deps.Store.Queries)
			categoryService := categories.NewService(categoryStore)
			categoryHandler := categories.NewHandler(categoryService)
			categories.RegisterRoutes(business, categoryHandler, authComponents.AuthZ)

			catalogStore := catalog.NewStore(deps.Store.Queries)
			catalogService := catalog.NewService(catalogStore)

			inventoryReadStore := inventory.NewReadStore(deps.Store.Queries)
			inventoryTxManager := inventory.NewTxManager(deps.Store)
			inventoryService := inventory.NewService(inventoryReadStore, inventoryTxManager)
			inventoryHandler := inventory.NewHandler(inventoryService, catalogService)
			inventory.RegisterRoutes(business, inventoryHandler, authComponents.AuthZ)

			catalogHandler := catalog.NewHandler(catalogService)
			catalog.RegisterRoutes(business, catalogHandler, authComponents.AuthZ)

			salesReadStore := sales.NewReadStore(deps.Store.Queries)
			salesTxManager := sales.NewTxManager(deps.Store)
			salesService := sales.NewService(salesReadStore, salesTxManager)
			salesHandler := sales.NewHandler(salesService)
			sales.RegisterRoutes(business, salesHandler, authComponents.AuthZ)

			fiscalProvider := &fiscal.MockProvider{}

			checkoutTxManager := checkout.NewTxManager(deps.Store)
			checkoutService := checkout.NewService(checkoutTxManager, fiscalProvider)
			checkoutHandler := checkout.NewHandler(checkoutService)
			checkout.RegisterRoutes(business, checkoutHandler, authComponents.AuthZ)

			paymentsStore := payments.NewStore(deps.Store.Queries)
			paymentsService := payments.NewService(paymentsStore)
			paymentsHandler := payments.NewHandler(paymentsService)
			payments.RegisterRoutes(business, paymentsHandler, authComponents.AuthZ)

			fiscalService := fiscal.NewService(fiscal.NewStore(deps.Store.Queries), fiscalProvider)
			fiscalHandler := fiscal.NewHandler(fiscalService)
			fiscal.RegisterRoutes(business, fiscalHandler, authComponents.AuthZ)

			receiptStore := receipt.NewStore(deps.Store.Queries)
			receiptService := receipt.NewService(receiptStore)
			receiptHandler := receipt.NewHandler(receiptService)
			receipt.RegisterRoutes(business, receiptHandler, authComponents.AuthZ)
		})
	}

	return router
}

func registerAdministrationRoutes(router chi.Router, deps Dependencies, components *AuthComponents) {
	writer := audit.NewWriter()

	organizationService, err := organizations.NewService(deps.Store, writer, deps.TenantCreationEnabled, components.Invalidator)
	if err != nil {
		slog.Error("failed to build organizations service", "error", err)
		return
	}
	storeService, err := stores.NewService(deps.Store, writer, components.Invalidator)
	if err != nil {
		slog.Error("failed to build stores service", "error", err)
		return
	}
	roleService, err := roles.NewService(
		roles.NewStore(deps.Store.Queries),
		roles.NewTxProvider(deps.Store, writer),
		components.Invalidator,
		deps.Clock,
	)
	if err != nil {
		slog.Error("failed to build roles service", "error", err)
		return
	}

	membershipService := memberships.NewService(
		memberships.NewStore(deps.Store.Queries),
		memberships.NewTxManager(deps.Store, writer),
		memberships.NewSessionRevoker(),
		components.Invalidator,
	)
	auditService := audit.NewService(audit.NewReadStore(deps.Store.Queries))
	organizationHandler := organizations.NewHandler(organizationService)
	organizations.RegisterSelfServiceRoutes(router, organizationHandler, components.AuthN)

	router.Route("/v1", func(versioned chi.Router) {
		organizations.RegisterRoutes(versioned, organizationHandler, components.AuthN, components.AuthZ)
		memberships.RegisterRoutes(versioned, memberships.NewHandler(membershipService), components.AuthN, components.AuthZ)
		invitations.RegisterAdminRoutes(versioned, components.Invitations, components.AuthN, components.AuthZ)
		versioned.Group(func(protected chi.Router) {
			protected.Use(components.AuthN.RequireAccessToken)
			stores.RegisterRoutes(protected, stores.NewHandler(storeService), components.AuthZ)
		})
	})

	// These packages register their own /v1 prefix and share one authentication boundary.
	router.Group(func(protected chi.Router) {
		protected.Use(components.AuthN.RequireAccessToken)
		roles.RegisterRoutes(protected, roles.NewHandler(roleService), components.AuthZ)
		audit.RegisterRoutes(protected, audit.NewHandler(auditService), components.AuthZ)
	})
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
	_, _ = w.Write(fmt.Appendf(nil, `{"status":"unavailable","detail":%q}`, detail))
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
