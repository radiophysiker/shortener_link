package app

import (
	"errors"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"github.com/radiophysiker/shortener_link/internal/config"
	v1 "github.com/radiophysiker/shortener_link/internal/controller/http/v1"
	"github.com/radiophysiker/shortener_link/internal/handlers"
	"github.com/radiophysiker/shortener_link/internal/repository"
	"github.com/radiophysiker/shortener_link/internal/usecases"
)

func Run() error {
	logger, err := zap.NewProduction()
	if err != nil {
		return fmt.Errorf("cannot create logger: %w", err)
	}
	zap.ReplaceGlobals(logger)
	defer func(logger *zap.Logger) {
		err := logger.Sync()
		if err != nil {
			logger.Error("cannot sync logger", zap.Error(err))
		}
	}(logger)

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("cannot load config: %w", err)
	}
	urlRepository := repository.NewURLRepository()
	useCasesURLShortener := usecases.NewURLShortener(urlRepository, cfg)
	urlHandler := handlers.NewURLHandler(useCasesURLShortener, cfg)

	router := v1.NewRouter(urlHandler)

	logger.Info("Starting server", zap.String("port", cfg.ServerPort))

	err = http.ListenAndServe(cfg.ServerPort, router)
	if err != nil {
		logger.Error("HTTP server has encountered an error", zap.Error(err))
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("HTTP server has encountered an error: %w", err)
		}
	}
	return nil
}
