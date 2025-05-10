package usecases

import (
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
)

type BatchItem struct {
	CorrelationID string
	OriginalURL   string
	ShortURL      string
}

type URLRepository interface {
	Save(url entity.URL) error
	GetFullURL(shortURL string) (string, error)
	SaveBatch(urls []entity.URL) error
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
func (us URLUseCase) CreateShortURL(fullURL string) (string, error) {
	return us.retryCreateShortURL(1, fullURL)
}

// retryCreateShortURL is a recursive function that tries to create a short URL.
func (us URLUseCase) retryCreateShortURL(numberAttempts int, fullURL string) (string, error) {
	shortURL := utils.GetShortRandomString(lenShortenedURL)
	url := entity.URL{
		ShortURL: shortURL,
		FullURL:  fullURL,
	}
	err := us.urlRepository.Save(url)
	if err != nil {
		if errors.Is(err, ErrEmptyFullURL) {
			return "", ErrEmptyFullURL
		}
		if errors.Is(err, ErrURLGeneratedBefore) {
			if numberAttempts >= maxNumberAttempts {
				// if we have reached the maximum number of attempts, we return an error
				return "", ErrFailedToGenerateShortURL
			} else {
				return us.retryCreateShortURL(numberAttempts+1, fullURL)
			}
		}
		return "", fmt.Errorf("failed to save URL: %w", err)
	}
	return shortURL, nil
}

// CreateBatchURLs creates multiple short URLs in a batch.
func (us URLUseCase) CreateBatchURLs(items []BatchItem) ([]BatchItem, error) {
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
		})

		items[i].ShortURL = shortURL
		resultItems = append(resultItems, items[i])
	}

	if err := us.urlRepository.SaveBatch(urls); err != nil {
		return nil, fmt.Errorf("failed to save batch of URLs: %w", err)
	}

	return resultItems, nil
}

// GetFullURL returns the full URL by the short URL.
func (us URLUseCase) GetFullURL(shortURL string) (string, error) {
	fullURL, err := us.urlRepository.GetFullURL(shortURL)
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
