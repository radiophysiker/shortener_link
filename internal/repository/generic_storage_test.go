package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/radiophysiker/shortener_link/internal/entity"
	"github.com/radiophysiker/shortener_link/internal/usecases"
)

func TestNewGenericStorage(t *testing.T) {
	// Тест хранилища в памяти
	urlStorage, err := NewGenericStorage("")
	assert.NoError(t, err, "NewGenericStorage should not return an error")
	assert.NotNil(t, urlStorage, "NewGenericStorage should return a non-nil GenericStorage")
}

func TestSave(t *testing.T) {
	urlStorage, err := NewGenericStorage("")
	require.NoError(t, err, "NewGenericStorage should not return an error")

	url := entity.URL{
		ShortURL: "short",
		FullURL:  "full",
	}
	err = urlStorage.Save(context.Background(), url)
	assert.NoError(t, err, "Save should not return an error")
}

func TestSaveWithEmptyFullURL(t *testing.T) {
	urlStorage, err := NewGenericStorage("")
	require.NoError(t, err, "NewGenericStorage should not return an error")

	url := entity.URL{
		ShortURL: "short",
		FullURL:  "",
	}
	err = urlStorage.Save(context.Background(), url)
	assert.Equal(t, usecases.ErrEmptyFullURL, err, "Save should return ErrEmptyFullURL for empty FullURL")
}

func TestGetFullURL(t *testing.T) {
	urlStorage, err := NewGenericStorage("")
	require.NoError(t, err, "NewGenericStorage should not return an error")

	url := entity.URL{
		ShortURL: "short",
		FullURL:  "full",
	}
	err = urlStorage.Save(context.Background(), url)
	require.NoError(t, err, "Save should not return an error")

	fullURL, err := urlStorage.GetFullURL(context.Background(), "short")
	require.NoError(t, err, "GetFullURL should not return an error")
	assert.Equal(t, url.FullURL, fullURL, "GetFullURL should return the correct FullURL")
}

func TestGetFullURLWithEmptyShortURL(t *testing.T) {
	urlStorage, err := NewGenericStorage("")
	require.NoError(t, err, "NewGenericStorage should not return an error")

	_, err = urlStorage.GetFullURL(context.Background(), "")
	assert.Equal(t, usecases.ErrEmptyShortURL, err, "GetFullURL should return ErrEmptyShortURL for empty shortURL")
	//assert.Empty(t, fullURL, "GetFullURL should return an empty string for empty shortURL")
}

func TestGetFullURLWithNotFoundShortURL(t *testing.T) {
	urlStorage, err := NewGenericStorage("")
	require.NoError(t, err, "NewGenericStorage should not return an error")

	url := entity.URL{
		ShortURL: "short",
		FullURL:  "full",
	}
	err = urlStorage.Save(context.Background(), url)
	require.NoError(t, err, "Save should not return an error")

	fullURL, err := urlStorage.GetFullURL(context.Background(), "not_found")
	assert.ErrorIs(t, err, usecases.ErrURLNotFound, "GetFullURL should return ErrURLNotFound for not found shortURL")
	assert.Empty(t, fullURL, "GetFullURL should return an empty string for not found shortURL")
}

func TestSaveExistingURL(t *testing.T) {
	urlStorage, err := NewGenericStorage("")
	require.NoError(t, err, "NewGenericStorage should not return an error")

	url := entity.URL{
		ShortURL: "short",
		FullURL:  "full",
	}

	err = urlStorage.Save(context.Background(), url)
	require.NoError(t, err, "First save should not return an error")

	// Try to save the same short URL with a different full URL
	urlDuplicate := entity.URL{
		ShortURL: "short",
		FullURL:  "another_full",
	}
	err = urlStorage.Save(context.Background(), urlDuplicate)
	assert.ErrorIs(t, err, usecases.ErrURLGeneratedBefore, "Save should return ErrURLGeneratedBefore for duplicate shortURL")
}

func TestIsShortURLExists(t *testing.T) {
	urlStorage, err := NewGenericStorage("")
	require.NoError(t, err, "NewGenericStorage should not return an error")

	url := entity.URL{
		ShortURL: "short",
		FullURL:  "full",
	}

	// Test non-existent URL
	err = urlStorage.checkURLExists(url, true)
	if err != nil {
		require.NoError(t, err)
	}

	// Save URL and test again
	err = urlStorage.checkURLExists(url, true)
	require.NoError(t, err)
}

func TestMultipleSave(t *testing.T) {
	urlStorage, err := NewGenericStorage("")
	require.NoError(t, err, "NewGenericStorage should not return an error")

	urls := []entity.URL{
		{ShortURL: "short1", FullURL: "full1"},
		{ShortURL: "short2", FullURL: "full2"},
		{ShortURL: "short3", FullURL: "full3"},
	}

	// Save all URLs
	for _, url := range urls {
		err := urlStorage.Save(context.Background(), url)
		require.NoError(t, err)
	}

	// Verify each URL
	for _, url := range urls {
		fullURL, err := urlStorage.GetFullURL(context.Background(), url.ShortURL)
		require.NoError(t, err)
		assert.Equal(t, url.FullURL, fullURL)
	}
}

func TestClose(t *testing.T) {
	urlStorage, err := NewGenericStorage("")
	require.NoError(t, err, "NewGenericStorage should not return an error")

	err = urlStorage.Close()
	assert.NoError(t, err, "Close should not return an error")
}
