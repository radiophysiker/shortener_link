package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"github.com/radiophysiker/shortener_link/internal/config"
	"github.com/radiophysiker/shortener_link/internal/entity"
	"github.com/radiophysiker/shortener_link/internal/middleware"
)

type URLsByUserGetter interface {
	GetUserURLs(ctx context.Context, userID string) ([]entity.URL, error)
}

type GetUserURLsHandler struct {
	getter URLsByUserGetter
	config *config.Config
}

func NewGetUserURLsHandler(getter URLsByUserGetter, cfg *config.Config) *GetUserURLsHandler {
	return &GetUserURLsHandler{
		getter: getter,
		config: cfg,
	}
}

type URLResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func (h *GetUserURLsHandler) GetUserURLs(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	urls, err := h.getter.GetUserURLs(r.Context(), userID)
	if err != nil {
		zap.L().Error("failed to get user URLs", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(urls) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var response []URLResponse
	for _, url := range urls {
		response = append(response, URLResponse{
			ShortURL:    h.config.BaseURL + "/" + url.ShortURL,
			OriginalURL: url.FullURL,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		zap.L().Error("failed to encode response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
	}
}
