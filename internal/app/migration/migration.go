package migration

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/faust8888/shortener/internal/middleware/logger"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func Run(dataSourceName string) error {
	db, err := sql.Open("pgx", dataSourceName)
	if err != nil {
		return fmt.Errorf("migration.run: oppening conection - %w", err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("migration.run: creating migration driver - %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://internal/app/migration/sql",
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("migration.run: creating migration instance - %w", err)
	}

	if err = m.Up(); !errors.Is(err, migrate.ErrNoChange) && err != nil {
		return fmt.Errorf("migration.run: applying migrations - %w", err)
	}
	logger.Log.Info("migration.run: all sql scripts applied successfully")
	return nil
}
