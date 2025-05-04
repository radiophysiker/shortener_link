package repository

import (
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

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
	pool, err := pgxpool.Connect(context.Background(), dsn)
	if err != nil {
		return nil, err
	}
	ps := &PostgresStorage{
		pool: pool,
	}
	if err := ps.createTable(); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to create table: %w", err)
	}
	return ps, nil
}

func (p *PostgresStorage) createTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS shortened_urls (
		id SERIAL PRIMARY KEY,
		short_url VARCHAR(10) NOT NULL UNIQUE,
		full_url TEXT NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_short_url ON shortened_urls(short_url);
	`

	_, err := p.pool.Exec(context.Background(), query)
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

func (p *PostgresStorage) Save(url entity.URL) error {
	fullURL := url.FullURL
	if fullURL == "" {
		return usecases.ErrEmptyFullURL
	}
	exists, err := p.isShortURLExists(url)
	if err != nil {
		return fmt.Errorf("failed to check if short URL exists: %w", err)
	}
	if exists {
		return fmt.Errorf("%w for: %s", usecases.ErrURLExists, url.ShortURL)
	}
	query := `
	INSERT INTO shortened_urls (short_url, full_url)
	VALUES ($1, $2);
	`
	_, err = p.pool.Exec(context.Background(), query, url.ShortURL, url.FullURL)
	if err != nil {
		return fmt.Errorf("failed to save URL: %w", err)
	}
	return nil

}

func (p *PostgresStorage) GetFullURL(shortURL ShortURL) (FullURL, error) {
	if shortURL == "" {
		return "", usecases.ErrEmptyShortURL
	}
	query := `
	SELECT full_url
	FROM shortened_urls
	WHERE short_url = $1;
	`
	var fullURL FullURL
	err := p.pool.QueryRow(context.Background(), query, shortURL).Scan(&fullURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", fmt.Errorf("%w: %s", usecases.ErrURLNotFound, shortURL) // Используем ошибку из usecases
		}
		return "", fmt.Errorf("couldn't get full URL for %s: %w", shortURL, err)
	}
	return fullURL, nil
}

func (p *PostgresStorage) SaveBatch(urls []entity.URL) error {
	if len(urls) == 0 {
		return usecases.ErrEmptyBatch
	}
	tx, err := p.pool.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			err := tx.Rollback(context.Background())
			if err != nil {
				zap.L().Error("failed to rollback transaction", zap.Error(err))
			}
		}
	}()

	for _, url := range urls {
		query := `
		INSERT INTO shortened_urls (short_url, full_url)
		VALUES ($1, $2);
		`
		_, err = tx.Exec(context.Background(), query, url.ShortURL, url.FullURL)
		if err != nil {
			return fmt.Errorf("failed to save URL in batch: %w", err)
		}
	}

	if err = tx.Commit(context.Background()); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
