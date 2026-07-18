package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gabrielalc23/pdv/config"
	"github.com/gabrielalc23/pdv/internal/catalog"
	"github.com/gabrielalc23/pdv/internal/categories"
	"github.com/gabrielalc23/pdv/internal/checkout"
	"github.com/gabrielalc23/pdv/internal/fiscal"
	"github.com/gabrielalc23/pdv/internal/inventory"
	"github.com/gabrielalc23/pdv/internal/payments"
	"github.com/gabrielalc23/pdv/internal/platform/database"
	apphttp "github.com/gabrielalc23/pdv/internal/platform/http"
	"github.com/gabrielalc23/pdv/internal/platform/tenancy"
	"github.com/go-chi/chi/v5"
	"github.com/gabrielalc23/pdv/internal/products"
	"github.com/gabrielalc23/pdv/internal/receipt"
	"github.com/gabrielalc23/pdv/internal/sales"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	store, err := database.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	router := apphttp.NewRouter(apphttp.Dependencies{
		HealthHandler: func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		},
	})

	resolver := tenancy.NewContextResolver()

	router.Group(func(r chi.Router) {
		r.Use(tenancy.Middleware)

		productStore := products.NewStore(store.Queries)
		productService := products.NewService(productStore)
		productHandler := products.NewHandler(productService, resolver)
		products.RegisterRoutes(r, productHandler)

		categoryStore := categories.NewStore(store.Queries)
		categoryService := categories.NewService(categoryStore)
		categoryHandler := categories.NewHandler(categoryService, resolver)
		categories.RegisterRoutes(r, categoryHandler)

		inventoryReadStore := inventory.NewReadStore(store.Queries)
		inventoryTxManager := inventory.NewTxManager(store)
		inventoryService := inventory.NewService(inventoryReadStore, inventoryTxManager)
		inventoryHandler := inventory.NewHandler(inventoryService, resolver)
		inventory.RegisterRoutes(r, inventoryHandler)

		catalogStore := catalog.NewStore(store.Queries)
		catalogService := catalog.NewService(catalogStore)
		catalogHandler := catalog.NewHandler(catalogService, resolver)
		catalog.RegisterRoutes(r, catalogHandler)

		salesReadStore := sales.NewReadStore(store.Queries)
		salesTxManager := sales.NewTxManager(store)
		salesService := sales.NewService(salesReadStore, salesTxManager)
		salesHandler := sales.NewHandler(salesService, resolver)
		sales.RegisterRoutes(r, salesHandler)

		fiscalProvider := &fiscal.MockProvider{}

		checkoutTxManager := checkout.NewTxManager(store)
		checkoutService := checkout.NewService(checkoutTxManager, fiscalProvider)
		checkoutHandler := checkout.NewHandler(checkoutService, resolver)
		checkout.RegisterRoutes(r, checkoutHandler)

		paymentsStore := payments.NewStore(store.Queries)
		paymentsService := payments.NewService(paymentsStore)
		paymentsHandler := payments.NewHandler(paymentsService, resolver)
		payments.RegisterRoutes(r, paymentsHandler)

		fiscalService := fiscal.NewService(fiscal.NewStore(store.Queries), fiscalProvider)
		fiscalHandler := fiscal.NewHandler(fiscalService, resolver)
		fiscal.RegisterRoutes(r, fiscalHandler)

		receiptStore := receipt.NewStore(store.Queries)
		receiptService := receipt.NewService(receiptStore)
		receiptHandler := receipt.NewHandler(receiptService, resolver)
		receipt.RegisterRoutes(r, receiptHandler)
	})

	server := &http.Server{
		Addr:    cfg.Address,
		Handler: router,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	log.Printf("listening on %s", cfg.Address)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
