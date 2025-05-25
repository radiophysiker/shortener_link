package repository

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/radiophysiker/shortener_link/internal/entity"
	"github.com/radiophysiker/shortener_link/internal/usecases"
)

type (
	ShortURL = string
	FullURL  = string
)

type FileRecord struct {
	UUID        int64  `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	UserID      string `json:"user_id"`
}

type GenericStorage struct {
	filePath string
	urls     map[ShortURL]FullURL
	userURLs map[string][]entity.URL // userID -> URLs
	count    int64
	file     *os.File
}

func NewGenericStorage(filePath string) (*GenericStorage, error) {
	fs := &GenericStorage{
		urls:     make(map[ShortURL]FullURL),
		userURLs: make(map[string][]entity.URL),
		filePath: filePath,
		count:    0,
	}
	if filePath != "" {
		err := fs.init()
		if err != nil {
			return nil, err
		}
	}
	return fs, nil
}

// init initializes the GenericStorage by loading URL mapping data from the specified file.
func (fs *GenericStorage) init() error {
	file, err := os.OpenFile(fs.filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	fs.file = file

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var record FileRecord
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			return err
		}

		url := entity.URL{
			ShortURL: record.ShortURL,
			FullURL:  record.OriginalURL,
			UserID:   record.UserID,
		}

		fs.urls[record.ShortURL] = record.OriginalURL
		fs.userURLs[record.UserID] = append(fs.userURLs[record.UserID], url)
		fs.count++
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to scan file: %w", err)
	}

	return nil
}

func (fs *GenericStorage) checkURLExists(url entity.URL, isGeneratedShortURL bool) error {
	if isGeneratedShortURL {
		_, exists := fs.urls[url.ShortURL]
		if exists {
			return usecases.ErrURLGeneratedBefore
		}
	}
	_, exists := fs.urls[url.ShortURL]
	if exists {
		return usecases.ErrURLConflict
	}
	for _, fullURL := range fs.urls {
		if fullURL == url.FullURL {
			return usecases.ErrURLConflict
		}
	}
	return nil
}

func (fs *GenericStorage) getCount() int64 {
	fs.count++
	return fs.count
}

func (fs *GenericStorage) Save(ctx context.Context, url entity.URL) error {
	if url.FullURL == "" {
		return usecases.ErrEmptyFullURL
	}
	err := fs.checkURLExists(url, true)
	if err != nil {
		return err
	}

	if fs.filePath != "" {
		uuid := fs.getCount()
		record := FileRecord{
			UUID:        uuid,
			ShortURL:    url.ShortURL,
			OriginalURL: url.FullURL,
			UserID:      url.UserID,
		}
		data, err := json.Marshal(record)
		if err != nil {
			return fmt.Errorf("failed to marshal record: %w", err)
		}

		if _, err := fs.file.Write(append(data, '\n')); err != nil {
			return fmt.Errorf("failed to write to file: %w", err)
		}
	}

	fs.urls[url.ShortURL] = url.FullURL
	fs.userURLs[url.UserID] = append(fs.userURLs[url.UserID], url)
	return nil
}

func (fs *GenericStorage) GetFullURL(ctx context.Context, shortURL ShortURL) (FullURL, error) {
	if shortURL == "" {
		return "", usecases.ErrEmptyShortURL
	}
	fullURL, exists := fs.urls[shortURL]
	if !exists {
		return "", fmt.Errorf("%w for: %s", usecases.ErrURLNotFound, shortURL)
	}
	return fullURL, nil
}

func (fs *GenericStorage) GetURLsByUserID(ctx context.Context, userID string) ([]entity.URL, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID cannot be empty")
	}

	urls, exists := fs.userURLs[userID]
	if !exists {
		return []entity.URL{}, nil
	}

	return urls, nil
}

func (fs *GenericStorage) Close() error {
	if fs.file != nil {
		return fs.file.Close()
	}
	return nil
}

func (fs *GenericStorage) SaveBatch(ctx context.Context, urls []entity.URL) error {
	if len(urls) == 0 {
		return nil
	}

	for _, url := range urls {
		if url.FullURL == "" {
			return usecases.ErrEmptyFullURL
		}
		if url.ShortURL == "" {
			return usecases.ErrEmptyShortURL
		}
		err := fs.checkURLExists(url, false)
		if err != nil {
			return fmt.Errorf("failed to check if URL exists: %w", err)
		}
		if fs.filePath != "" {
			uuid := fs.getCount()
			record := FileRecord{
				UUID:        uuid,
				ShortURL:    url.ShortURL,
				OriginalURL: url.FullURL,
				UserID:      url.UserID,
			}
			data, err := json.Marshal(record)
			if err != nil {
				return fmt.Errorf("failed to marshal record: %w", err)
			}

			if _, err := fs.file.Write(append(data, '\n')); err != nil {
				return fmt.Errorf("failed to write to file: %w", err)
			}
		}
		fs.urls[url.ShortURL] = url.FullURL
		fs.userURLs[url.UserID] = append(fs.userURLs[url.UserID], url)
	}

	return nil
}
