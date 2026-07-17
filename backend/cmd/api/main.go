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

	productService := products.NewService(store.Queries)
	productHandler := products.NewHandler(productService)
	products.RegisterRoutes(router, productHandler)

	categoryService := categories.NewService(store.Queries)
	categoryHandler := categories.NewHandler(categoryService)
	categories.RegisterRoutes(router, categoryHandler)

	inventoryService := inventory.NewService(store.Queries, inventory.NewTxManager(store))
	inventoryHandler := inventory.NewHandler(inventoryService)
	inventory.RegisterRoutes(router, inventoryHandler)

	catalogService := catalog.NewService(store.Queries)
	catalogHandler := catalog.NewHandler(catalogService)
	catalog.RegisterRoutes(router, catalogHandler)

	salesService := sales.NewService(store.Queries, sales.NewTxManager(store))
	salesHandler := sales.NewHandler(salesService)
	sales.RegisterRoutes(router, salesHandler)

	fiscalProvider := &fiscal.MockProvider{}

	checkoutService := checkout.NewService(checkout.NewTxManager(store), fiscalProvider, store)
	checkoutHandler := checkout.NewHandler(checkoutService)
	checkout.RegisterRoutes(router, checkoutHandler)

	paymentsService := payments.NewService(store.Queries)
	paymentsHandler := payments.NewHandler(paymentsService)
	payments.RegisterRoutes(router, paymentsHandler)

	fiscalService := fiscal.NewService(store.Queries, fiscalProvider)
	fiscalHandler := fiscal.NewHandler(fiscalService)
	fiscal.RegisterRoutes(router, fiscalHandler)

	receiptService := receipt.NewService(store.Queries)
	receiptHandler := receipt.NewHandler(receiptService)
	receipt.RegisterRoutes(router, receiptHandler)

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
