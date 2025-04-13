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

func TestSaveExistingURL(t *testing.T) {
	urlStorage := NewURLRepository()
	url := entity.URL{
		ShortURL: "short",
		FullURL:  "full",
	}

	err := urlStorage.Save(url)
	require.NoError(t, err, "First save should not return an error")

	// Try to save the same short URL with different full URL
	urlDuplicate := entity.URL{
		ShortURL: "short",
		FullURL:  "another_full",
	}
	err = urlStorage.Save(urlDuplicate)
	assert.ErrorIs(t, err, usecases.ErrURLExists, "Save should return ErrURLExists for duplicate shortURL")
}

func TestIsShortURLExists(t *testing.T) {
	urlStorage := NewURLRepository()
	url := entity.URL{
		ShortURL: "short",
		FullURL:  "full",
	}

	// Test non-existent URL
	exists := urlStorage.IsShortURLExists(url)
	assert.False(t, exists, "Should return false for non-existent URL")

	// Save URL and test again
	err := urlStorage.Save(url)
	require.NoError(t, err)

	exists = urlStorage.IsShortURLExists(url)
	assert.True(t, exists, "Should return true for existing URL")
}

func TestMultipleSave(t *testing.T) {
	urlStorage := NewURLRepository()
	urls := []entity.URL{
		{ShortURL: "short1", FullURL: "full1"},
		{ShortURL: "short2", FullURL: "full2"},
		{ShortURL: "short3", FullURL: "full3"},
	}

	// Save all URLs
	for _, url := range urls {
		err := urlStorage.Save(url)
		require.NoError(t, err)
	}

	// Verify each URL
	for _, url := range urls {
		fullURL, err := urlStorage.GetFullURL(url.ShortURL)
		require.NoError(t, err)
		assert.Equal(t, url.FullURL, fullURL)
	}
}
