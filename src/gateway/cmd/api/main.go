package main

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-resty/resty/v2"
	amqp "github.com/rabbitmq/amqp091-go"

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
	fp  service.FilePublisher
}

func main() {
	cfg := getConfig()
	s3c := s3.NewFromConfig(aws.Config{
		Region: cfg.aws.s3Region,
		Credentials: credentials.NewStaticCredentialsProvider(
			cfg.aws.accessKeyId,
			cfg.aws.secretAccessKey,
			"",
		),
	})

	conn, err := amqp.Dial(cfg.rabbit.dsn())
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	defer conn.Close()

	enc, err := encryptor.NewEncryptor(cfg.encryptKey)
	if err != nil {
		slog.Error("Failed to create encryptor", "error", err)
		os.Exit(1)
	}

	fileRepository := repository.NewStore(s3c)
	fileService := service.NewFileService(enc, fileRepository)

	filePublisher, err := service.NewPublisher(conn, cfg.rabbit.queue)
	if err != nil {
		slog.Error("Failed to create publisher", "error", err)
		os.Exit(1)
	}

	app := application{
		cfg: cfg,
		rc:  resty.New(),
		fs:  fileService,
		fp:  filePublisher,
	}

	if err := app.run(); err != nil {
		slog.Error("Error running application", "error", err.Error())
		os.Exit(1)
	}
}
