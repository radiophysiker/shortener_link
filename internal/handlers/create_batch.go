package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"go.uber.org/zap"

	"github.com/radiophysiker/shortener_link/internal/config"
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

type CreateBatchURLs interface {
	CreateBatchURLs(items []usecases.BatchItem) ([]usecases.BatchItem, error)
}

type CreateBatchURLsHandler struct {
	creator CreateBatchURLs
	config  *config.Config
}

func NewCreateBatchURLsHandler(createBatchURLs CreateBatchURLs, cfg *config.Config) *CreateBatchURLsHandler {
	return &CreateBatchURLsHandler{creator: createBatchURLs, config: cfg}
}

func (h *CreateBatchURLsHandler) CreateBatchURLs(w http.ResponseWriter, r *http.Request) {
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
			_, err := w.Write([]byte("url is empty"))
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

	resultItems, err := h.creator.CreateBatchURLs(batchItems)
	if err != nil {
		zap.L().Error("cannot create batch of short URLs", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	responseItems := make([]BatchURLResponse, 0, len(resultItems))
	baseURL := h.config.BaseURL
	for _, item := range resultItems {
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
