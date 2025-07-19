package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/radiophysiker/shortener_link/internal/config"
	"github.com/radiophysiker/shortener_link/internal/entity"
	"github.com/radiophysiker/shortener_link/internal/utils"
)

const lenShortenedURL = 6
const maxNumberAttempts = 5

var (
	ErrURLGeneratedBefore       = errors.New("shortURL already generated before")
	ErrFailedToGenerateShortURL = errors.New("failed to generate short URL")
	ErrEmptyFullURL             = errors.New("empty full URL")
	ErrEmptyShortURL            = errors.New("empty short URL")
	ErrURLNotFound              = errors.New("URL not found")
	ErrEmptyBatch               = errors.New("empty batch")
	ErrURLConflict              = errors.New("URL already exists in the database")
	ErrEmptyUserID              = errors.New("empty user ID")
)

type BatchItem struct {
	CorrelationID string
	OriginalURL   string
	ShortURL      string
}

type URLRepository interface {
	Save(ctx context.Context, url entity.URL) error
	GetFullURL(ctx context.Context, shortURL string) (string, error)
	SaveBatch(ctx context.Context, urls []entity.URL) error
	GetUserURLs(ctx context.Context, userID string) ([]entity.URL, error)
}

type URLUseCase struct {
	urlRepository URLRepository
	config        *config.Config
}

func NewURLShortener(re URLRepository, cfg *config.Config) *URLUseCase {
	return &URLUseCase{
		urlRepository: re,
		config:        cfg,
	}
}

// CreateShortURL creates a short URL.
func (us URLUseCase) CreateShortURL(ctx context.Context, fullURL string, userID string) (string, error) {
	return us.retryCreateShortURL(ctx, 1, fullURL, userID)
}

// retryCreateShortURL is a recursive function that tries to create a short URL.
func (us URLUseCase) retryCreateShortURL(ctx context.Context, numberAttempts int, fullURL string, userID string) (string, error) {
	shortURL := utils.GetShortRandomString(lenShortenedURL)
	url := entity.URL{
		ShortURL: shortURL,
		FullURL:  fullURL,
		UserID:   userID,
	}
	err := us.urlRepository.Save(ctx, url)
	if err != nil {
		if errors.Is(err, ErrEmptyFullURL) {
			return "", ErrEmptyFullURL
		}
		if errors.Is(err, ErrURLGeneratedBefore) {
			if numberAttempts >= maxNumberAttempts {
				return "", ErrFailedToGenerateShortURL
			} else {
				return us.retryCreateShortURL(ctx, numberAttempts+1, fullURL, userID)
			}
		}
		if errors.Is(err, ErrURLConflict) {
			errStr := err.Error()
			if len(errStr) > len(ErrURLConflict.Error())+2 { // +2 for ": "
				existingShortURL := errStr[len(ErrURLConflict.Error())+2:]
				return existingShortURL, ErrURLConflict
			}
		}
		return "", fmt.Errorf("failed to save URL: %w", err)
	}
	return shortURL, nil
}

// CreateBatchURLs creates multiple short URLs in a batch.
func (us URLUseCase) CreateBatchURLs(ctx context.Context, items []BatchItem, userID string) ([]BatchItem, error) {
	if len(items) == 0 {
		return nil, ErrEmptyBatch
	}
	urls := make([]entity.URL, 0, len(items))
	resultItems := make([]BatchItem, 0, len(items))

	for i := range items {
		if items[i].OriginalURL == "" {
			return nil, ErrEmptyFullURL
		}
		if items[i].CorrelationID == "" {
			return nil, ErrEmptyShortURL
		}

		shortURL := utils.GetShortRandomString(lenShortenedURL)
		urls = append(urls, entity.URL{
			ShortURL: shortURL,
			FullURL:  items[i].OriginalURL,
			UserID:   userID,
		})

		items[i].ShortURL = shortURL
		resultItems = append(resultItems, items[i])
	}

	if err := us.urlRepository.SaveBatch(ctx, urls); err != nil {
		return nil, fmt.Errorf("failed to save batch of URLs: %w", err)
	}

	return resultItems, nil
}

// GetFullURL returns the full URL by the short URL.
func (us URLUseCase) GetFullURL(ctx context.Context, shortURL string) (string, error) {
	fullURL, err := us.urlRepository.GetFullURL(ctx, shortURL)
	if err != nil {
		if errors.Is(err, ErrEmptyShortURL) {
			return "", ErrEmptyShortURL
		}
		if errors.Is(err, ErrURLNotFound) {
			return "", fmt.Errorf("%w for: %s", ErrURLNotFound, shortURL)
		}
		return "", fmt.Errorf("failed to get full URL: %w", err)
	}
	return fullURL, nil
}

func (us URLUseCase) GetUserURLs(ctx context.Context, userID string) ([]entity.URL, error) {
	if userID == "" {
		return nil, ErrEmptyUserID
	}

	urls, err := us.urlRepository.GetUserURLs(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user URLs: %w", err)
	}

	return urls, nil
}
