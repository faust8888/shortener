package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/faust8888/shortener/internal/app/model"
	_ "github.com/jackc/pgx/v5/stdlib"
	"time"
)

var ErrUniqueIndexConstraint = errors.New("full_url unique index constraint violation")

type Repository struct {
	db *sql.DB
}

func (r *Repository) Save(urlHash string, fullURL string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	res, err := r.db.ExecContext(ctx,
		"INSERT INTO shortener (short_url, full_url) VALUES ($1, $2) ON CONFLICT (full_url) DO NOTHING", urlHash, fullURL)
	if err != nil {
		return fmt.Errorf("repository.postgres.save: %w", err)
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return ErrUniqueIndexConstraint
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

func (r *Repository) SaveAll(batch map[string]model.CreateShortDTO) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("postgres.repository.saveAll.begin - %w", err)
	}
	for _, batchItem := range batch {
		_, err = tx.ExecContext(context.Background(),
			"INSERT INTO shortener (short_url, full_url) VALUES ($1, $2)", batchItem.HashURL, batchItem.OriginalURL)
		if err != nil {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				return fmt.Errorf("postgres.repository.saveAll.rollback - %w", rollbackErr)
			}
			return fmt.Errorf("postgres.repository.saveAll.insert: %w", err)
		}
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("postgres.repository.saveAll.commit - %w", err)
	}
	return nil
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
