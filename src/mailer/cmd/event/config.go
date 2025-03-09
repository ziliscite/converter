package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"sync"
)

type RabbitMQ struct {
	host     string
	username string
	password string
	port     string
	queue    string
}

func (r RabbitMQ) dsn() string {
	return fmt.Sprintf("amqps://%s:%s@%s:%s", r.username, r.password, r.host, r.port)
}

type Config struct {
	mail struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
	mq RabbitMQ
}

var (
	instance Config
	once     sync.Once
)

func getConfig() Config {
	once.Do(func() {
		instance = Config{}

		flag.StringVar(&instance.mail.host, "smtp-host", os.Getenv("SMTP_HOST"), "SMTP host")

		port, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
		if err != nil {
			slog.Error("SMTP_PORT is not a valid integer", "error", err)
			os.Exit(1)
		}

		flag.IntVar(&instance.mail.port, "smtp-port", port, "SMTP port")
		flag.StringVar(&instance.mail.username, "smtp-username", os.Getenv("SMTP_USERNAME"), "SMTP username")
		flag.StringVar(&instance.mail.password, "smtp-password", os.Getenv("SMTP_PASSWORD"), "SMTP password")
		flag.StringVar(&instance.mail.sender, "smtp-sender", os.Getenv("SMTP_FROM"), "SMTP sender")

		flag.Parse()
	})

	return instance
}
