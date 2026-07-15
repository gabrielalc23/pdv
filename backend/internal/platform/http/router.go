package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Dependencies struct {
	HealthHandler http.HandlerFunc
}

func NewRouter(deps Dependencies) chi.Router {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.ClientIPFromXFF())
	router.Use(middleware.Recoverer)

	router.Get("/health", deps.HealthHandler)

	return router
}
