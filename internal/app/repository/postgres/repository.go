package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"time"
)

type Repository struct {
	db *sql.DB
}

func (r *Repository) Save(urlHash string, fullURL string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO shortener (short_url, full_url) VALUES ($1, $2)", urlHash, fullURL)
	if err != nil {
		return fmt.Errorf("repository.postgres.save: %w", err)
	}
	return nil
}

func (r *Repository) FindByHash(hash string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	query := `
        SELECT full_url
        FROM shortener
        WHERE short_url = $1
    `
	var fullURL string
	err := r.db.QueryRowContext(ctx, query, hash).Scan(&fullURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("short url not found for %s", hash)
		}
		return "", fmt.Errorf("failed to find record by hash: %v", err)
	}
	return fullURL, nil
}

func (r *Repository) Ping() (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := r.db.PingContext(ctx); err != nil {
		return false, fmt.Errorf("couldn't ping the PostgreSQL server: %s", err.Error())
	}
	return true, nil
}

func NewPostgresRepository(dataSourceName string) *Repository {
	db, err := sql.Open("pgx", dataSourceName)
	if err != nil {
		panic(err)
	}
	return &Repository{db: db}
}
