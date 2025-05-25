package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"

	"go.uber.org/zap"

	"github.com/radiophysiker/shortener_link/internal/config"
	"github.com/radiophysiker/shortener_link/internal/middleware"
	"github.com/radiophysiker/shortener_link/internal/usecases"
	"github.com/radiophysiker/shortener_link/internal/utils"
)

type BatchURLRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type BatchURLResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type BatchURLCreator interface {
	CreateBatchURLs(ctx context.Context, items []usecases.BatchItem, userID string) ([]usecases.BatchItem, error)
}

type CreateBatchURLsHandler struct {
	creator BatchURLCreator
	config  *config.Config
}

func NewCreateBatchURLsHandler(createBatchURLs BatchURLCreator, cfg *config.Config) *CreateBatchURLsHandler {
	return &CreateBatchURLsHandler{creator: createBatchURLs, config: cfg}
}

func (h *CreateBatchURLsHandler) CreateBatchURLs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Получаем UserID из контекста
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok || userID == "" {
		w.WriteHeader(http.StatusInternalServerError)
		zap.L().Error("userID not found in context")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		zap.L().Error("cannot read request body", zap.Error(err))
		return
	}

	if len(body) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte("empty request body"))
		if err != nil {
			utils.WriteErrorWithCannotWriteResponse(w, err)
		}
		return
	}

	var requestItems []BatchURLRequest
	if err := json.Unmarshal(body, &requestItems); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte("invalid json format"))
		if err != nil {
			utils.WriteErrorWithCannotWriteResponse(w, err)
		}
		return
	}

	if len(requestItems) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte("empty batch"))
		if err != nil {
			utils.WriteErrorWithCannotWriteResponse(w, err)
		}
		return
	}

	batchItems := make([]usecases.BatchItem, 0, len(requestItems))
	for _, item := range requestItems {
		if item.OriginalURL == "" {
			w.WriteHeader(http.StatusBadRequest)
			_, err := w.Write([]byte("original_url is empty"))
			if err != nil {
				utils.WriteErrorWithCannotWriteResponse(w, err)
			}
			return
		}

		if item.CorrelationID == "" {
			w.WriteHeader(http.StatusBadRequest)
			_, err := w.Write([]byte("correlation_id is empty"))
			if err != nil {
				utils.WriteErrorWithCannotWriteResponse(w, err)
			}
			return
		}

		// Validate URL format
		parsedURL, err := url.Parse(item.OriginalURL)
		if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
			w.WriteHeader(http.StatusBadRequest)
			zap.L().Error("invalid url format", zap.Error(err), zap.String("url", item.OriginalURL))
			_, err := w.Write([]byte("invalid url format"))
			if err != nil {
				utils.WriteErrorWithCannotWriteResponse(w, err)
			}
			return
		}

		batchItems = append(batchItems, usecases.BatchItem{
			CorrelationID: item.CorrelationID,
			OriginalURL:   item.OriginalURL,
		})
	}

	items, err := h.creator.CreateBatchURLs(ctx, batchItems, userID)
	if err != nil {
		if errors.Is(err, usecases.ErrURLConflict) {
			zap.L().Warn("some URLs in batch already exist", zap.Error(err))
			if len(items) == 0 {
				w.WriteHeader(http.StatusConflict)
				_, err := w.Write([]byte("all URLs in batch already exist"))
				if err != nil {
					utils.WriteErrorWithCannotWriteResponse(w, err)
				}
				return
			}
		} else {
			// Для других ошибок возвращаем 500
			zap.L().Error("cannot create batch of short URLs", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	responseItems := make([]BatchURLResponse, 0, len(items))
	baseURL := h.config.BaseURL
	for _, item := range items {
		shortURLPath, err := url.JoinPath(baseURL, item.ShortURL)
		if err != nil {
			zap.L().Error("cannot join base URL and short URL", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		responseItems = append(responseItems, BatchURLResponse{
			CorrelationID: item.CorrelationID,
			ShortURL:      shortURLPath,
		})
	}

	jsonResp, err := json.Marshal(responseItems)
	if err != nil {
		utils.WriteErrorWithCannotWriteResponse(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(jsonResp)
	if err != nil {
		utils.WriteErrorWithCannotWriteResponse(w, err)
	}
}
