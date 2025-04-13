package repository

import (
	"github.com/radiophysiker/shortener_link/internal/config"
	"github.com/radiophysiker/shortener_link/internal/entity"
)

type Saver interface {
	Save(url entity.URL) error
}

type Finder interface {
	GetFullURL(shortURL ShortURL) (FullURL, error)
	isShortURLExists(url entity.URL) bool
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
	if cfg.FileStoragePath != "" {
		return NewFileStorage(cfg.FileStoragePath)
	}
	return NewMemoryRepository(), nil
}
