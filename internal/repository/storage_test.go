package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/radiophysiker/shortener_link/internal/entity"
	"github.com/radiophysiker/shortener_link/internal/usecases"
)

func TestNewURLRepository(t *testing.T) {
	urlStorage := NewURLRepository()
	assert.NotNil(t, urlStorage, "NewURLRepository should return a non-nil URLStorage")
}

func TestSave(t *testing.T) {
	urlStorage := NewURLRepository()
	url := entity.URL{
		ShortURL: "short",
		FullURL:  "full",
	}
	err := urlStorage.Save(url)
	assert.NoError(t, err, "Save should not return an error")
}

func TestSaveWithEmptyFullURL(t *testing.T) {
	urlStorage := NewURLRepository()
	url := entity.URL{
		ShortURL: "short",
		FullURL:  "",
	}
	err := urlStorage.Save(url)
	assert.Equal(t, usecases.ErrEmptyFullURL, err, "Save should return ErrEmptyFullURL for empty FullURL")
}

func TestGetFullURL(t *testing.T) {
	urlStorage := NewURLRepository()
	url := entity.URL{
		ShortURL: "short",
		FullURL:  "full",
	}
	err := urlStorage.Save(url)
	require.NoError(t, err, "Save should not return an error")

	fullURL, err := urlStorage.GetFullURL("short")
	require.NoError(t, err, "GetFullURL should not return an error")
	assert.Equal(t, url.FullURL, fullURL, "GetFullURL should return the correct FullURL")
}

func TestGetFullURLWithEmptyShortURL(t *testing.T) {
	urlStorage := NewURLRepository()

	fullURL, err := urlStorage.GetFullURL("")
	assert.Equal(t, usecases.ErrEmptyShortURL, err, "GetFullURL should return ErrEmptyShortURL for empty shortURL")
	assert.Empty(t, fullURL, "GetFullURL should return an empty string for empty shortURL")
}

func TestGetFullURLWithNotFoundShortURL(t *testing.T) {
	urlStorage := NewURLRepository()
	url := entity.URL{
		ShortURL: "short",
		FullURL:  "full",
	}
	err := urlStorage.Save(url)
	require.NoError(t, err, "Save should not return an error")

	fullURL, err := urlStorage.GetFullURL("not_found")
	assert.ErrorIs(t, err, usecases.ErrURLNotFound, "GetFullURL should return ErrURLNotFound for not found shortURL")
	assert.Empty(t, fullURL, "GetFullURL should return an empty string for not found shortURL")
}
