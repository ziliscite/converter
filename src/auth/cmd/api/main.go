package main

import (
	"context"
	"github.com/ziliscite/video-to-mp3/auth/internal/repository"
	"github.com/ziliscite/video-to-mp3/auth/internal/service"
	"github.com/ziliscite/video-to-mp3/auth/pkg/db"
	"github.com/ziliscite/video-to-mp3/auth/pkg/validator"
	"log/slog"
	"os"
	"time"
)

type application struct {
	cfg Config
	v   *validator.Validator
	us  service.UserService
}

func newApplication(config Config, us service.UserService) *application {
	return &application{
		cfg: config,
		v:   validator.New(),
		us:  us,
	}
}

func main() {
	slog.Info("Hallow World")

	cfg := getConfig()
	db.AutoMigrate(cfg.db.dsn())

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pool, err := db.Open(ctx, cfg.db.dsn())
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}

	userRepository := repository.NewUserRepository(pool)
	userService := service.NewUserService(userRepository)

	app := newApplication(cfg, userService)
	if err = app.run(); err != nil {
		slog.Error("Error running application", "error", err.Error())
		os.Exit(1)
	}
}
