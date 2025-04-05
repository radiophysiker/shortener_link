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

type URLStorage struct {
	urls map[ShortURL]FullURL
}

func NewURLRepository() *URLStorage {
	return &URLStorage{
		urls: make(map[ShortURL]FullURL),
	}
}

// IsShortURLExists checks if the short URL exists in memory.
func (s URLStorage) IsShortURLExists(url entity.URL) bool {
	for shortURL := range s.urls {
		if shortURL == url.ShortURL {
			return true
		}
	}
	return false
}

// Save saves the URL in memory.
func (s URLStorage) Save(url entity.URL) error {
	fullURL := url.FullURL
	if fullURL == "" {
		return usecases.ErrEmptyFullURL
	}
	if s.IsShortURLExists(url) {
		return fmt.Errorf("%w for: %s", usecases.ErrURLExists, url.ShortURL)
	}
	s.urls[url.ShortURL] = fullURL
	return nil
}

// GetFullURL returns the full URL by the short URL.
func (s URLStorage) GetFullURL(shortURL ShortURL) (FullURL, error) {
	if shortURL == "" {
		return "", usecases.ErrEmptyShortURL
	}
	fullURL, exists := s.urls[shortURL]
	if !exists {
		return "", fmt.Errorf("%w for: %s", usecases.ErrURLNotFound, shortURL)
	}
	return fullURL, nil
}
