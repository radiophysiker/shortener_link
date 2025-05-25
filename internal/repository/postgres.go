package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"

	"github.com/radiophysiker/shortener_link/internal/entity"
	"github.com/radiophysiker/shortener_link/internal/usecases"
)

type PostgresStorage struct {
	pool *pgxpool.Pool
}

func NewPostgresStorage(dsn string) (*PostgresStorage, error) {
	if dsn == "" {
		return nil, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.Connect(ctx, dsn)
	if err != nil {
		return nil, err
	}
	ps := &PostgresStorage{
		pool: pool,
	}
	if err := ps.createTable(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to create table: %w", err)
	}
	return ps, nil
}

func (p *PostgresStorage) createTable(ctx context.Context) error {
	query := `
	CREATE TABLE IF NOT EXISTS shortened_urls (
		id SERIAL PRIMARY KEY,
		short_url VARCHAR(10) NOT NULL UNIQUE,
		full_url TEXT NOT NULL,
		user_id VARCHAR(36) NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_short_url ON shortened_urls(short_url);
	CREATE INDEX IF NOT EXISTS idx_user_id ON shortened_urls(user_id);
	CREATE UNIQUE INDEX IF NOT EXISTS idx_full_url_user_id ON shortened_urls(full_url, user_id);
	`

	_, err := p.pool.Exec(ctx, query)
	return err
}

func (p *PostgresStorage) Ping(ctx context.Context) error {
	return p.pool.Ping(ctx)
}

func (p *PostgresStorage) Close() error {
	p.pool.Close()
	return nil
}

func (p *PostgresStorage) isShortURLExists(url entity.URL) (bool, error) {
	const query = `
	SELECT EXISTS (
		SELECT 1
		FROM shortened_urls
		WHERE short_url = $1
	);
`

	var exists bool
	err := p.pool.QueryRow(context.Background(), query, url.ShortURL).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (p *PostgresStorage) Save(ctx context.Context, url entity.URL) error {
	fullURL := url.FullURL
	if fullURL == "" {
		return usecases.ErrEmptyFullURL
	}

	// First try to get the existing short URL for this full URL and user
	existingShortURL, err := p.GetShortURLByFullURLAndUserID(ctx, fullURL, url.UserID)
	if err == nil {
		// If we found an existing short URL, return it with a conflict error
		return fmt.Errorf("%w: %s", usecases.ErrURLConflict, existingShortURL)
	}

	// If no existing URL found, proceed with saving
	query := `
	INSERT INTO shortened_urls (short_url, full_url, user_id)
	VALUES ($1, $2, $3);
	`
	_, err = p.pool.Exec(ctx, query, url.ShortURL, url.FullURL, url.UserID)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == pgerrcode.UniqueViolation {
			// If we get a unique violation, try to get the existing short URL again
			existingShortURL, err = p.GetShortURLByFullURLAndUserID(ctx, fullURL, url.UserID)
			if err != nil {
				return fmt.Errorf("failed to get existing short URL: %w", err)
			}
			return fmt.Errorf("%w: %s", usecases.ErrURLConflict, existingShortURL)
		}
		return fmt.Errorf("failed to save URL: %w", err)
	}
	return nil
}

func (p *PostgresStorage) GetFullURL(ctx context.Context, shortURL ShortURL) (FullURL, error) {
	if shortURL == "" {
		return "", usecases.ErrEmptyShortURL
	}
	query := `
	SELECT full_url
	FROM shortened_urls
	WHERE short_url = $1;
	`
	var fullURL FullURL
	err := p.pool.QueryRow(ctx, query, shortURL).Scan(&fullURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", fmt.Errorf("%w: %s", usecases.ErrURLNotFound, shortURL)
		}
		return "", fmt.Errorf("couldn't get full URL for %s: %w", shortURL, err)
	}
	return fullURL, nil
}

func (p *PostgresStorage) GetShortURLByFullURLAndUserID(ctx context.Context, fullURL string, userID string) (string, error) {
	if fullURL == "" {
		return "", usecases.ErrEmptyFullURL
	}

	query := `
	SELECT short_url
	FROM shortened_urls
	WHERE full_url = $1 AND user_id = $2;
	`

	var shortURL string
	err := p.pool.QueryRow(ctx, query, fullURL, userID).Scan(&shortURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", fmt.Errorf("%w for URL: %s and user: %s", usecases.ErrURLNotFound, fullURL, userID)
		}
		return "", fmt.Errorf("failed to get short URL for %s: %w", fullURL, err)
	}

	return shortURL, nil
}

func (p *PostgresStorage) GetURLsByUserID(ctx context.Context, userID string) ([]entity.URL, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID cannot be empty")
	}

	query := `
	SELECT short_url, full_url, user_id
	FROM shortened_urls
	WHERE user_id = $1
	ORDER BY created_at DESC;
	`

	rows, err := p.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query URLs for user %s: %w", userID, err)
	}
	defer rows.Close()

	var urls []entity.URL
	for rows.Next() {
		var url entity.URL
		err := rows.Scan(&url.ShortURL, &url.FullURL, &url.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan URL row: %w", err)
		}
		urls = append(urls, url)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return urls, nil
}

func (p *PostgresStorage) SaveBatch(ctx context.Context, urls []entity.URL) error {
	if len(urls) == 0 {
		return usecases.ErrEmptyBatch
	}
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			err := tx.Rollback(ctx)
			if err != nil {
				zap.L().Error("failed to rollback transaction", zap.Error(err))
			}
		}
	}()

	for _, url := range urls {
		query := `
		INSERT INTO shortened_urls (short_url, full_url, user_id)
		VALUES ($1, $2, $3);
		`
		_, err = tx.Exec(ctx, query, url.ShortURL, url.FullURL, url.UserID)
		if err != nil {
			zap.L().Error("failed to save URL in batch", zap.Error(err))
			if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == pgerrcode.UniqueViolation {
				// If we get a unique violation, try to get the existing short URL again
				existingShortURL, err := p.GetShortURLByFullURLAndUserID(ctx, url.FullURL, url.UserID)
				if err != nil {
					return fmt.Errorf("failed to get existing short URL: %w", err)
				}
				return fmt.Errorf("%w: %s", usecases.ErrURLConflict, existingShortURL)
			}
			return fmt.Errorf("failed to save URL: %w", err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
