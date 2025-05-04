package repository

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/radiophysiker/shortener_link/internal/entity"
	"github.com/radiophysiker/shortener_link/internal/usecases"
)

type FileStorage struct {
	filePath string
	urls     map[ShortURL]FullURL
	count    int64
	file     *os.File
}

type FileRecord struct {
	UUID        int64  `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func NewFileStorage(filePath string) (*FileStorage, error) {
	fs := &FileStorage{
		urls:     make(map[ShortURL]FullURL),
		filePath: filePath,
		count:    0,
	}

	err := fs.init()
	if err != nil {
		return nil, err
	}

	return fs, nil
}

// init initializes the FileStorage by loading URL mapping data from the specified file.
func (fs *FileStorage) init() error {
	file, err := os.OpenFile(fs.filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	fs.file = file

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var record entity.URL
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			return err
		}
		fs.urls[record.ShortURL] = record.FullURL
		fs.count++
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to scan file: %w", err)
	}

	return nil
}

func (fs *FileStorage) isShortURLExists(url entity.URL) (bool, error) {
	_, exists := fs.urls[url.ShortURL]
	return exists, nil
}

func (fs *FileStorage) getCount() int64 {
	fs.count++
	return fs.count
}

func (fs *FileStorage) Save(url entity.URL) error {
	uuid := fs.getCount()
	exists, err := fs.isShortURLExists(url)
	if err != nil {
		return fmt.Errorf("failed to check if short URL exists: %w", err)
	}
	if exists {
		return fmt.Errorf("%w for: %s", usecases.ErrURLExists, url.ShortURL)
	}
	record := FileRecord{
		UUID:        uuid,
		ShortURL:    url.ShortURL,
		OriginalURL: url.FullURL,
	}
	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("failed to marshal record: %w", err)
	}

	if _, err := fs.file.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	fs.urls[url.ShortURL] = url.FullURL
	return nil
}

func (fs *FileStorage) GetFullURL(shortURL ShortURL) (FullURL, error) {
	fullURL, exists := fs.urls[shortURL]
	if !exists {
		return "", fmt.Errorf("%w for: %s", usecases.ErrURLNotFound, shortURL)
	}
	return fullURL, nil
}

func (fs *FileStorage) Close() error {
	if fs.file != nil {
		return fs.file.Close()
	}
	return nil
}

func (fs *FileStorage) SaveBatch(urls []entity.URL) error {
	if len(urls) == 0 {
		return nil
	}

	for _, url := range urls {
		if url.FullURL == "" {
			return usecases.ErrEmptyFullURL
		}

		uuid := fs.getCount()
		record := FileRecord{
			UUID:        uuid,
			ShortURL:    url.ShortURL,
			OriginalURL: url.FullURL,
		}
		data, err := json.Marshal(record)
		if err != nil {
			return fmt.Errorf("failed to marshal record: %w", err)
		}

		if _, err := fs.file.Write(append(data, '\n')); err != nil {
			return fmt.Errorf("failed to write to file: %w", err)
		}

		fs.urls[url.ShortURL] = url.FullURL
	}

	return nil
}
