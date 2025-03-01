package db

import (
	"errors"
	"log/slog"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func AutoMigrate(dsn string) {
	m, err := migrate.New("file://migrations", dsn)
	if err != nil {
		slog.Error("Error setting up migrations", "error", err.Error())
		os.Exit(1)
	}

	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		slog.Error("Error running migrations", "error", err.Error())
		os.Exit(1)
	}

	slog.Info("Migrations applied successfully!")
}
