package repository

import (
	"context"

	"github.com/radiophysiker/shortener_link/internal/config"
	"github.com/radiophysiker/shortener_link/internal/entity"
)

type Saver interface {
	Save(ctx context.Context, url entity.URL) error
	SaveBatch(ctx context.Context, urls []entity.URL) error
}

type Finder interface {
	GetFullURL(ctx context.Context, shortURL ShortURL) (FullURL, error)
}

type Closer interface {
	Close() error
}

type Storage interface {
	Saver
	Finder
	Closer
}

func NewStorage(cfg *config.Config) (Storage, error) {
	if cfg.DatabaseDSN != "" {
		pgStorage, err := NewPostgresStorage(cfg.DatabaseDSN)
		if err != nil {
			return nil, err
		}
		return pgStorage, nil
	}
	return NewGenericStorage(cfg.FileStoragePath)
}
