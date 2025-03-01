package db

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v4/pgxpool"
)

func Open(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	db, err := pgxpool.Connect(ctx, dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(ctx); err != nil {
		return nil, err
	}

	slog.Info("Connected to authentication database")
	return db, nil
}
