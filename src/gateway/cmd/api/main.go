package main

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-resty/resty/v2"
	"log/slog"
	"net/http"
	"os"
	"time"
)

type application struct {
	cfg Config
	s3c *s3.Client
	rc  *resty.Client
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cfg := getConfig()

	awsCfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		slog.Error("Failed to load default AWS config", "error", err.Error())
		os.Exit(1)
	}
	client := s3.NewFromConfig(awsCfg)

	app := application{
		cfg: cfg,
		s3c: client,
		rc:  resty.New(),
	}

	app.run()
}
