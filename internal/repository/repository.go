package repository

import (
	"github.com/radiophysiker/shortener_link/internal/config"
	"github.com/radiophysiker/shortener_link/internal/entity"
)

type Saver interface {
	Save(url entity.URL) error
	SaveBatch(urls []entity.URL) error
}

type Finder interface {
	GetFullURL(shortURL ShortURL) (FullURL, error)
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
