package handlers

import (
	"github.com/radiophysiker/shortener_link/internal/config"
)

type URL interface {
	CreateShortURL(fullURL string) (string, error)
	GetFullURL(shortURL string) (string, error)
}

type URLHandler struct {
	URLUseCase URL
	config     *config.Config
}

func NewURLHandler(u URL, cfg *config.Config) *URLHandler {
	return &URLHandler{
		URLUseCase: u,
		config:     cfg,
	}
}
