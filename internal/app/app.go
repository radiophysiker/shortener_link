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
	// Create logger
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
	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("cannot load config: %w", err)
	}
	logger.Info("Loaded config", zap.Any("config", cfg))
	// Create storage
	storage, err := repository.NewStorage(cfg)
	if err != nil {
		return fmt.Errorf("cannot create storage: %w", err)
	}
	defer func(storage repository.Storage) {
		logger.Info("Closing storage")
		err := storage.Close()
		if err != nil {
			logger.Error("cannot close storage", zap.Error(err))
		}
	}(storage)
	// Create use cases
	useCasesURLShortener := usecases.NewURLShortener(storage, cfg)

	// Create handlers
	createHandler := handlers.NewCreateHandler(useCasesURLShortener, cfg)
	createBatchURLsHandler := handlers.NewCreateBatchURLsHandler(useCasesURLShortener, cfg)
	getHandler := handlers.NewGetHandler(useCasesURLShortener)
	getUserURLsHandler := handlers.NewGetUserURLsHandler(useCasesURLShortener, cfg)
	pg, err := repository.NewPostgresStorage(cfg.DatabaseDSN)
	if err != nil {
		return fmt.Errorf("cannot connect to postgres: %w", err)
	}
	defer func(pg *repository.PostgresStorage) {
		err := pg.Close()
		if err != nil {
			logger.Error("cannot close postgres", zap.Error(err))
		}
	}(pg)
	pingHandler := handlers.NewPingHandler(pg)

	// Create router
	router := v1.NewRouter(cfg, createHandler, createBatchURLsHandler, getHandler, getUserURLsHandler, pingHandler)
	// Start server
	logger.Info("Starting server", zap.String("port", cfg.ServerPort))
	err = http.ListenAndServe(cfg.ServerPort, router)
	if err != nil {
		logger.Error("HTTP server has encountered an error", zap.Error(err))
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("HTTP server has encountered an error: %w", err)
		}
	}
	logger.Info("Server has been stopped")
	return nil
}
