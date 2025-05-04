package handlers

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type Pinger interface {
	Ping(ctx context.Context) error
}

type NoOpPinger struct{}

func (n *NoOpPinger) Ping(ctx context.Context) error {
	return nil
}

type PingHandler struct {
	pinger Pinger
}

func NewPingHandler(pinger Pinger) *PingHandler {
	return &PingHandler{pinger: pinger}
}

func (h *PingHandler) Ping(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if h.pinger == nil {
		zap.L().Error("DB connection error")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("DB connection error"))
		return
	}
	if err := h.pinger.Ping(ctx); err != nil {
		zap.L().Error("DB connection error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("DB connection error"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
