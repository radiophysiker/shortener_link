package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/radiophysiker/shortener_link/internal/middleware"
	"github.com/radiophysiker/shortener_link/internal/utils"
	"go.uber.org/zap"
)

type URLDeleter interface {
	DeleteBatch(ctx context.Context, shortURLs []string, userID string) error
}

type DeleteURLsHandler struct {
	deleter URLDeleter
}

func NewDeleteURLsHandler(deleter URLDeleter) *DeleteURLsHandler {
	return &DeleteURLsHandler{
		deleter: deleter,
	}
}

func (h *DeleteURLsHandler) DeleteUserURLs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var shortURLs []string
	if err := json.NewDecoder(r.Body).Decode(&shortURLs); err != nil {
		zap.L().Error("failed to decode JSON", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		_, writeErr := w.Write([]byte("invalid JSON format"))
		if writeErr != nil {
			utils.WriteErrorWithCannotWriteResponse(w, writeErr)
		}
		return
	}

	if len(shortURLs) == 0 {
		zap.L().Error("empty shortURLs array")
		w.WriteHeader(http.StatusBadRequest)
		_, writeErr := w.Write([]byte("empty URL list"))
		if writeErr != nil {
			utils.WriteErrorWithCannotWriteResponse(w, writeErr)
		}
		return
	}

	err := h.deleter.DeleteBatch(ctx, shortURLs, userID)
	if err != nil {
		zap.L().Error("failed to delete URLs", zap.Error(err), zap.String("userID", userID))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	zap.L().Info("URLs deletion accepted", zap.Strings("shortURLs", shortURLs), zap.String("userID", userID))
	w.WriteHeader(http.StatusAccepted)
}
