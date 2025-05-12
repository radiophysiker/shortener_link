package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-chi/chi"
	"go.uber.org/zap"

	"github.com/radiophysiker/shortener_link/internal/usecases"
	"github.com/radiophysiker/shortener_link/internal/utils"
)

type URLGetter interface {
	GetFullURL(ctx context.Context, shortURL string) (string, error)
}

type GetHandler struct {
	getter URLGetter
}

func NewGetHandler(getter URLGetter) *GetHandler {
	return &GetHandler{
		getter: getter,
	}
}

func (h *GetHandler) GetFullURL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	shortURL := chi.URLParam(r, "id")
	fullURL, err := h.getter.GetFullURL(ctx, shortURL)
	if err != nil {
		if errors.Is(err, usecases.ErrEmptyShortURL) {
			zap.L().Error("short url is empty", zap.Error(err), zap.String("shortURL", shortURL))
			w.WriteHeader(http.StatusBadRequest)
			_, err := w.Write([]byte("short url is empty"))
			if err != nil {
				utils.WriteErrorWithCannotWriteResponse(w, err)
			}
			return
		}
		if errors.Is(err, usecases.ErrURLNotFound) {
			zap.L().Error("url is not found for shortURL", zap.Error(err), zap.String("shortURL", shortURL))
			w.WriteHeader(http.StatusNotFound)
			_, err := w.Write([]byte("url is not found for " + shortURL))
			if err != nil {
				utils.WriteErrorWithCannotWriteResponse(w, err)
			}
			return
		}
		zap.L().Error("cannot get full URL: %v", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	zap.L().Info("get full URL", zap.String("shortURL", shortURL), zap.String("fullURL", fullURL))
	w.Header().Set("Location", fullURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
