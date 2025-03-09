package main

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/ziliscite/video-to-mp3/mailer/internal"
	"log/slog"
	"os"
)

func main() {
	cfg := getConfig()

	conn, err := amqp.Dial(cfg.mq.dsn())
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	defer conn.Close()

	mailer := internal.New(cfg.mail.host, cfg.mail.port, cfg.mail.username, cfg.mail.password, cfg.mail.sender)

	msrv, err := newService(cfg, conn, mailer)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	if err = msrv.listen(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
