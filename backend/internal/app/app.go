package app

import (
	"context"
	"encoding/json"
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
	"github.com/gabrielalc23/pdv/internal/platform/database"
	"github.com/gabrielalc23/pdv/internal/platform/tenancy"
	"github.com/gabrielalc23/pdv/internal/products"
	"github.com/gabrielalc23/pdv/internal/receipt"
	"github.com/gabrielalc23/pdv/internal/sales"
)

type Dependencies struct {
	Store *database.PostgresStore
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
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"status":"unavailable","detail":"database not connected"}`))
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		if err := deps.Store.Pool.Ping(ctx); err != nil {
			slog.Error("readiness ping failed", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"status":"unavailable","detail":"database ping failed"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
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

func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	_ = enc.Encode(payload)
}
