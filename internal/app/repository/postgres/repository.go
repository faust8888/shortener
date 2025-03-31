package postgres

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"time"
)

type Repository struct {
	db *sql.DB
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
