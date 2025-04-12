package handlers

import (
	"github.com/radiophysiker/shortener_link/internal/config"
)

type URLUseCase interface {
	URLCreator
	URLGetter
}

type URLHandler struct {
    URLUseCase
	config *config.Config
}

func NewURLHandler(u URLUseCase, cfg *config.Config) *URLHandler {
	return &URLHandler{
		URLUseCase: u,
		config:     cfg,
	}
}
