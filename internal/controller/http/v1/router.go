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
	userURLsHandler *handlers.UserURLsHandler,
) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestLogger())
	r.Use(middleware.GzipMiddleware)
	r.Use(middleware.AuthMiddleware)

	r.Post("/", createHandler.CreateShortURL)
	r.Get("/{id}", getHandler.GetFullURL)
	r.Post("/api/shorten", createHandler.CreateShortURLWithJSON)
	r.Post("/api/shorten/batch", createBatchURLsHandler.CreateBatchURLs)
	r.Get("/api/user/urls", userURLsHandler.GetUserURLs)
	r.Get("/ping", pingHandler.Ping)
	return r
}
