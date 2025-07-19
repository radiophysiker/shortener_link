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

type URLCreator interface {
	CreateShortURL(ctx context.Context, fullURL string, userID string) (string, error)
}

type CreateHandler struct {
	creator URLCreator
	config  *config.Config
}

func NewCreateHandler(creator URLCreator, cfg *config.Config) *CreateHandler {
	return &CreateHandler{
		creator: creator,
		config:  cfg,
	}
}

func (h *CreateHandler) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	ctx := r.Context()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		zap.L().Error("cannot read request body: %v", zap.Error(err))
		return
	}

	zap.L().Info("Creating short url for", zap.String("body", string(body)))
	fullURL := string(body)

	if fullURL == "" {
		w.WriteHeader(http.StatusBadRequest)
		zap.L().Error("url is empty")
		_, err := w.Write([]byte("url is empty"))
		if err != nil {
			utils.WriteErrorWithCannotWriteResponse(w, err)
		}
		return
	}

	parsedURL, err := url.Parse(fullURL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		w.WriteHeader(http.StatusBadRequest)
		zap.L().Error("invalid url format", zap.Error(err), zap.String("url", fullURL))
		_, err := w.Write([]byte("invalid url format"))
		if err != nil {
			utils.WriteErrorWithCannotWriteResponse(w, err)
		}
		return
	}

	// Получаем userID из контекста
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		zap.L().Error("userID not found in context")
		return
	}

	shortURL, err := h.creator.CreateShortURL(ctx, fullURL, userID)
	if err != nil {
		if errors.Is(err, usecases.ErrURLConflict) {
			w.WriteHeader(http.StatusConflict)
			baseURL := h.config.BaseURL
			shortURLPath, err := url.JoinPath(baseURL, shortURL)
			if err != nil {
				zap.L().Error("cannot join base URL and short URL", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			_, err = w.Write([]byte(shortURLPath))
			if err != nil {
				utils.WriteErrorWithCannotWriteResponse(w, err)
			}
			return
		}
		if errors.Is(err, usecases.ErrFailedToGenerateShortURL) {
			w.WriteHeader(http.StatusInternalServerError)
			zap.L().Error("failed to generate short URL", zap.Error(err), zap.String("url", fullURL))
			_, err := w.Write([]byte("failed to generate short URL"))
			if err != nil {
				utils.WriteErrorWithCannotWriteResponse(w, err)
			}
		}
		if errors.Is(err, usecases.ErrEmptyFullURL) {
			w.WriteHeader(http.StatusBadRequest)
			zap.L().Error("url is empty", zap.Error(err), zap.String("url", fullURL))
			_, err := w.Write([]byte("url is empty"))
			if err != nil {
				utils.WriteErrorWithCannotWriteResponse(w, err)
			}
			return
		}
		zap.L().Error("cannot create short URL: %v", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	baseURL := h.config.BaseURL
	shortURLPath, err := url.JoinPath(baseURL, shortURL)
	if err != nil {
		zap.L().Error("cannot join base URL and short URL", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, err = w.Write([]byte(shortURLPath))
	if err != nil {
		utils.WriteErrorWithCannotWriteResponse(w, err)
	}
}

type CreateShortURLEntryRequest struct {
	FullURL string `json:"url"`
}

type CreateShortURLEntryResponse struct {
	ShortURL string `json:"result"`
}

func (h *CreateHandler) CreateShortURLWithJSON(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	ctx := r.Context()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		zap.L().Error("cannot read request body: %v", zap.Error(err))
		return
	}

	var request CreateShortURLEntryRequest
	err = json.Unmarshal(body, &request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte("invalid json format"))
		if err != nil {
			utils.WriteErrorWithCannotWriteResponse(w, err)
		}
		return
	}

	var fullURL = request.FullURL
	if fullURL == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, err = w.Write([]byte("fullURL is empty"))
		if err != nil {
			utils.WriteErrorWithCannotWriteResponse(w, err)
			return
		}
		return
	}

	parsedURL, err := url.Parse(fullURL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		w.WriteHeader(http.StatusBadRequest)
		zap.L().Error("invalid url format", zap.Error(err), zap.String("url", fullURL))
		_, err := w.Write([]byte("invalid url format"))
		if err != nil {
			utils.WriteErrorWithCannotWriteResponse(w, err)
			return
		}
		return
	}

	// Получаем userID из контекста
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		zap.L().Error("userID not found in context")
		return
	}

	shortURL, err := h.creator.CreateShortURL(ctx, fullURL, userID)
	if err != nil {
		if errors.Is(err, usecases.ErrURLConflict) {
			baseURL := h.config.BaseURL
			shortURLPath, err := url.JoinPath(baseURL, shortURL)
			if err != nil {
				zap.L().Error("cannot join base URL and short URL", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			resp := CreateShortURLEntryResponse{ShortURL: shortURLPath}
			jsonResp, err := json.Marshal(resp)
			if err != nil {
				utils.WriteErrorWithCannotWriteResponse(w, err)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			_, err = w.Write(jsonResp)
			if err != nil {
				utils.WriteErrorWithCannotWriteResponse(w, err)
			}
			return
		}
		if errors.Is(err, usecases.ErrFailedToGenerateShortURL) {
			w.WriteHeader(http.StatusInternalServerError)
			zap.L().Error("failed to generate short URL", zap.Error(err), zap.String("url", fullURL))
			_, err := w.Write([]byte("failed to generate short URL"))
			if err != nil {
				utils.WriteErrorWithCannotWriteResponse(w, err)
			}
		}
		if errors.Is(err, usecases.ErrEmptyFullURL) {
			w.WriteHeader(http.StatusBadRequest)
			zap.L().Error("url is empty", zap.Error(err), zap.String("url", fullURL))
			_, err := w.Write([]byte("url is empty"))
			if err != nil {
				utils.WriteErrorWithCannotWriteResponse(w, err)
			}
			return
		}
		zap.L().Error("cannot create short URL: %v", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	baseURL := h.config.BaseURL
	shortURLPath, err := url.JoinPath(baseURL, shortURL)
	if err != nil {
		zap.L().Error("cannot join base URL and short URL", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	resp := CreateShortURLEntryResponse{ShortURL: shortURLPath}
	jsonResp, err := json.Marshal(resp)
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
