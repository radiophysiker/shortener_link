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

const UnknownUserID = "unknown"

type GenericStorage struct {
	filePath string
	urls     map[ShortURL]entity.URL
	count    int64
	file     *os.File
}

type FileRecord struct {
	UUID        int64  `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	UserID      string `json:"user_id"`
	IsDeleted   bool   `json:"is_deleted"`
}

func NewGenericStorage(filePath string) (*GenericStorage, error) {
	fs := &GenericStorage{
		urls:     make(map[ShortURL]entity.URL),
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
		userID := record.UserID
		if userID == "" {
			userID = UnknownUserID
		}

		fs.urls[record.ShortURL] = entity.URL{
			ShortURL:  record.ShortURL,
			FullURL:   record.OriginalURL,
			UserID:    userID,
			IsDeleted: record.IsDeleted,
		}
		fs.count++
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to scan file: %w", err)
	}

	return nil
}

// checkURLExists проверяет, существует ли URL в хранилище
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
	for _, existingURL := range fs.urls {
		if existingURL.FullURL == url.FullURL && !existingURL.IsDeleted {
			return usecases.ErrURLConflict
		}
	}
	return nil
}

// getCount возвращает следующий уникальный идентификатор
func (fs *GenericStorage) getCount() int64 {
	fs.count++
	return fs.count
}

// Save сохраняет URL в хранилище
func (fs *GenericStorage) Save(ctx context.Context, url entity.URL) error {
	if url.FullURL == "" {
		return usecases.ErrEmptyFullURL
	}

	if url.UserID == "" {
		url.UserID = UnknownUserID
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
			IsDeleted:   url.IsDeleted,
		}
		data, err := json.Marshal(record)
		if err != nil {
			return fmt.Errorf("failed to marshal record: %w", err)
		}

		if _, err := fs.file.Write(append(data, '\n')); err != nil {
			return fmt.Errorf("failed to write to file: %w", err)
		}
	}

	fs.urls[url.ShortURL] = url
	return nil
}

// GetFullURL возвращает полный URL по короткому URL
func (fs *GenericStorage) GetFullURL(ctx context.Context, shortURL ShortURL) (FullURL, error) {
	if shortURL == "" {
		return "", usecases.ErrEmptyShortURL
	}
	url, exists := fs.urls[shortURL]
	if !exists {
		return "", fmt.Errorf("%w for: %s", usecases.ErrURLNotFound, shortURL)
	}

	// Проверяем, удален ли URL
	if url.IsDeleted {
		return "", fmt.Errorf("%w: %s", usecases.ErrURLDeleted, shortURL)
	}

	return url.FullURL, nil
}

// Close закрывает файл хранилища
func (fs *GenericStorage) Close() error {
	if fs.file != nil {
		return fs.file.Close()
	}
	return nil
}

// SaveBatch сохраняет несколько URL в хранилище
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
		if url.UserID == "" {
			url.UserID = UnknownUserID
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
				IsDeleted:   url.IsDeleted,
			}
			data, err := json.Marshal(record)
			if err != nil {
				return fmt.Errorf("failed to marshal record: %w", err)
			}

			if _, err := fs.file.Write(append(data, '\n')); err != nil {
				return fmt.Errorf("failed to write to file: %w", err)
			}
		}
		fs.urls[url.ShortURL] = url
	}

	return nil
}

// GetUserURLs возвращает все URL для указанного пользователя
func (fs *GenericStorage) GetUserURLs(ctx context.Context, userID string) ([]entity.URL, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID cannot be empty")
	}

	var userURLs []entity.URL
	for _, url := range fs.urls {
		// Возвращаем только неудаленные URL пользователя
		if url.UserID == userID && !url.IsDeleted {
			userURLs = append(userURLs, url)
		}
	}
	return userURLs, nil
}

// DeleteBatch помечает URL как удаленные для указанного пользователя
func (fs *GenericStorage) DeleteBatch(ctx context.Context, shortURLs []string, userID string) error {
	if len(shortURLs) == 0 {
		return nil
	}

	if userID == "" {
		return usecases.ErrEmptyUserID
	}

	var updatedRecords []FileRecord
	for _, shortURL := range shortURLs {
		url, exists := fs.urls[shortURL]
		if !exists {
			continue
		}
		if url.UserID != userID {
			continue
		}
		if url.IsDeleted {
			continue
		}
		url.IsDeleted = true
		fs.urls[shortURL] = url
		if fs.filePath != "" {
			uuid := fs.getCount()
			record := FileRecord{
				UUID:        uuid,
				ShortURL:    url.ShortURL,
				OriginalURL: url.FullURL,
				UserID:      url.UserID,
				IsDeleted:   true,
			}
			updatedRecords = append(updatedRecords, record)
		}
	}
	if fs.filePath != "" && len(updatedRecords) > 0 {
		for _, record := range updatedRecords {
			data, err := json.Marshal(record)
			if err != nil {
				return fmt.Errorf("failed to marshal updated record: %w", err)
			}

			if _, err := fs.file.Write(append(data, '\n')); err != nil {
				return fmt.Errorf("failed to write updated record to file: %w", err)
			}
		}
	}

	return nil
}
