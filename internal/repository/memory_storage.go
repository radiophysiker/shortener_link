package repository

import (
	"fmt"

	"github.com/radiophysiker/shortener_link/internal/entity"
	"github.com/radiophysiker/shortener_link/internal/usecases"
)

type (
	ShortURL = string
	FullURL  = string
)

type MemoryStorage struct {
	urls map[ShortURL]FullURL
}

func NewMemoryRepository() *MemoryStorage {
	return &MemoryStorage{
		urls: make(map[ShortURL]FullURL),
	}
}

// isShortURLExists checks if the short URL exists in memory.
func (s *MemoryStorage) isShortURLExists(url entity.URL) (bool, error) {
	for shortURL := range s.urls {
		if shortURL == url.ShortURL {
			return true, nil
		}
	}
	return false, nil
}

// Save saves the URL in memory.
func (s *MemoryStorage) Save(url entity.URL) error {
	fullURL := url.FullURL
	if fullURL == "" {
		return usecases.ErrEmptyFullURL
	}
	exists, err := s.isShortURLExists(url)
	if err != nil {
		return fmt.Errorf("failed to check if short URL exists: %w", err)
	}
	if exists {
		return fmt.Errorf("%w for: %s", usecases.ErrURLExists, url.ShortURL)
	}
	s.urls[url.ShortURL] = fullURL
	return nil
}

// GetFullURL returns the full URL by the short URL.
func (s *MemoryStorage) GetFullURL(shortURL ShortURL) (FullURL, error) {
	if shortURL == "" {
		return "", usecases.ErrEmptyShortURL
	}
	fullURL, exists := s.urls[shortURL]
	if !exists {
		return "", fmt.Errorf("%w for: %s", usecases.ErrURLNotFound, shortURL)
	}
	return fullURL, nil
}

func (s *MemoryStorage) Close() error {
	s.urls = nil
	return nil
}
func (s *MemoryStorage) SaveBatch(urls []entity.URL) error {
	if len(urls) == 0 {
		return nil
	}

	for _, url := range urls {
		if url.FullURL == "" {
			return usecases.ErrEmptyFullURL
		}

		s.urls[url.ShortURL] = url.FullURL
	}

	return nil
}
