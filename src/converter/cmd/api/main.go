package main

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/ziliscite/video-to-mp3/converter/external/ffmpeg"
	"github.com/ziliscite/video-to-mp3/converter/internal/domain"
	"github.com/ziliscite/video-to-mp3/converter/internal/service"

	"github.com/ziliscite/video-to-mp3/converter/internal/repository"
	"github.com/ziliscite/video-to-mp3/converter/pkg/encryptor"

	"log/slog"
	"os"
)

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

	ffp, err := ffmpeg.Open()
	if err != nil {
		slog.Error("Failed to open ffmpeg", "error", err)
		os.Exit(1)
	}

	cvt := domain.NewConverter(ffp)
	fr := repository.NewStore(s3c)

	cvs := service.NewConverterService(cvt, fr, enc, cfg.aws.s3bucket.mp4, cfg.aws.s3bucket.mp3)

}
