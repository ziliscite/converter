package main

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-resty/resty/v2"
	"github.com/ziliscite/video-to-mp3/gateway/internal/repository"
	"github.com/ziliscite/video-to-mp3/gateway/internal/service"
	"github.com/ziliscite/video-to-mp3/gateway/pkg/encryptor"
	"log/slog"
	"os"
)

type application struct {
	cfg Config
	rc  *resty.Client
	fs  service.FileService
}

func main() {
	cfg := getConfig()
	client := s3.NewFromConfig(aws.Config{
		Region: cfg.aws.s3Region,
		Credentials: credentials.NewStaticCredentialsProvider(
			cfg.aws.accessKeyId,
			cfg.aws.secretAccessKey,
			"",
		),
	})

	enc, err := encryptor.NewEncryptor(cfg.encryptKey)
	if err != nil {
		slog.Error("Failed to create encryptor", "error", err)
		os.Exit(1)
	}

	fileRepository := repository.NewStore(client)
	fileService := service.NewUploadService(enc, fileRepository)

	app := application{
		cfg: cfg,
		rc:  resty.New(),
		fs:  fileService,
	}

	if err := app.run(); err != nil {
		slog.Error("Error running application", "error", err.Error())
		os.Exit(1)
	}
}
