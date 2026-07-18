package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gabrielalc23/pdv/config"
	"github.com/gabrielalc23/pdv/internal/app"
	"github.com/gabrielalc23/pdv/internal/platform/database"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	store, err := database.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer store.Close()

	handler := app.New(app.Dependencies{
		Store: store,
	})

	server := app.NewHTTPServer(cfg.Address, handler)

	go func() {
		<-ctx.Done()
		slog.Info("shutting down server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			slog.Error("shutdown error", "error", err)
		}
	}()

	slog.Info(fmt.Sprintf("listening on %s", cfg.Address))
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("server: %v", err)
	}
}
