package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"github.com/radiophysiker/shortener_link/internal/config"
	"github.com/radiophysiker/shortener_link/internal/entity"
	"github.com/radiophysiker/shortener_link/internal/middleware"
	"github.com/radiophysiker/shortener_link/internal/utils"
)

type URLByUserFinder interface {
	GetURLsByUserID(ctx context.Context, userID string) ([]entity.URL, error)
}

type UserURLsHandler struct {
	finder URLByUserFinder
	config *config.Config
}

type UserURLResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func NewUserURLsHandler(finder URLByUserFinder, cfg *config.Config) *UserURLsHandler {
	return &UserURLsHandler{
		finder: finder,
		config: cfg,
	}
}

func (h *UserURLsHandler) GetUserURLs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Получаем UserID из контекста
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Получаем все URL пользователя
	urls, err := h.finder.GetURLsByUserID(ctx, userID)
	if err != nil {
		zap.L().Error("failed to get user URLs", zap.Error(err), zap.String("userID", userID))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Если URL нет, возвращаем 204 No Content
	if len(urls) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Формируем ответ
	response := make([]UserURLResponse, 0, len(urls))
	for _, url := range urls {
		response = append(response, UserURLResponse{
			ShortURL:    h.config.BaseURL + "/" + url.ShortURL,
			OriginalURL: url.FullURL,
		})
	}

	// Отправляем JSON ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		zap.L().Error("failed to marshal response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = w.Write(jsonResponse)
	if err != nil {
		utils.WriteErrorWithCannotWriteResponse(w, err)
	}
}
