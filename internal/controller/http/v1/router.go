package v1

import (
	"github.com/go-chi/chi"
	"github.com/radiophysiker/shortener_link/internal/config"

	"github.com/radiophysiker/shortener_link/internal/handlers"
	"github.com/radiophysiker/shortener_link/internal/middleware"
)

// NewRouter creates a new router for the v1 API.
func NewRouter(
	cfg *config.Config,
	createHandler *handlers.CreateHandler,
	createBatchURLsHandler *handlers.CreateBatchURLsHandler,
	getHandler *handlers.GetHandler,
	getUserURLsHandler *handlers.GetUserURLsHandler,
	pingHandler *handlers.PingHandler,
) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestLogger())
	r.Use(middleware.GzipMiddleware)
	r.Use(middleware.AuthMiddleware(cfg))

	r.Post("/", createHandler.CreateShortURL)
	r.Get("/{id}", getHandler.GetFullURL)
	r.Post("/api/shorten", createHandler.CreateShortURLWithJSON)
	r.Post("/api/shorten/batch", createBatchURLsHandler.CreateBatchURLs)
	r.Get("/ping", pingHandler.Ping)
	r.Route("/api/user", func(r chi.Router) {
		r.Get("/urls", getUserURLsHandler.GetUserURLs)
	})
	return r
}
