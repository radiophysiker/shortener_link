package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"

	"go.uber.org/zap"

	"github.com/radiophysiker/shortener_link/internal/usecases"
	"github.com/radiophysiker/shortener_link/internal/utils"
)

type CreateShortURLEntryRequest struct {
	FullURL string `json:"url"`
}

type CreateShortURLEntryResponse struct {
	ShortURL string `json:"result"`
}

func (h *URLHandler) CreateShortURLWithJSON(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
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

	// Validate URL format
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

	shortURL, err := h.URLUseCase.CreateShortURL(fullURL)
	if err != nil {
		if errors.Is(err, usecases.ErrURLExists) {
			w.WriteHeader(http.StatusConflict)
			zap.L().Error("url already exists", zap.Error(err), zap.String("url", fullURL))
			_, err := w.Write([]byte("url already exists"))
			if err != nil {
				utils.WriteErrorWithCannotWriteResponse(w, err)
			}
			return
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
		zap.L().Error("cannot join base URL and short URL: %v", zap.Error(err))
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
		return
	}
}
