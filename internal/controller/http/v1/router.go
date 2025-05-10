package v1

import (
	"github.com/go-chi/chi"

	"github.com/radiophysiker/shortener_link/internal/handlers"
	"github.com/radiophysiker/shortener_link/internal/middleware"
)

// NewRouter creates a new router for the v1 API.
func NewRouter(
	createHandler *handlers.CreateHandler,
	createBatchURLsHandler *handlers.CreateBatchURLsHandler,
	getHandler *handlers.GetHandler,
	pingHandler *handlers.PingHandler,
) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestLogger())
	r.Use(middleware.GzipMiddleware)

	r.Post("/", createHandler.CreateShortURL)
	r.Get("/{id}", getHandler.GetFullURL)
	r.Post("/api/shorten", createHandler.CreateShortURLWithJSON)
	r.Post("/api/shorten/batch", createBatchURLsHandler.CreateBatchURLs)
	r.Get("/ping", pingHandler.Ping)
	return r
}
