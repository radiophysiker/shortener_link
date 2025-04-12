package v1

import (
	"github.com/go-chi/chi"

	"github.com/radiophysiker/shortener_link/internal/handlers"
	"github.com/radiophysiker/shortener_link/internal/middleware"
)

// NewRouter creates a new router for the v1 API.
func NewRouter(h *handlers.URLHandler) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestLogger())

	r.Post("/", h.CreateShortURL)
	r.Get("/{id}", h.GetFullURL)
	r.Post("/api/shorten", h.CreateShortURLWithJSON)
	return r
}
