package main

import (
	"flag"
	"os"
	"sync"
)

type Config struct {
	port    int
	secrets string
	db      struct {
		dsn string
	}
}

var (
	instance Config
	once     sync.Once
)

func getConfig() Config {
	once.Do(func() {
		instance = Config{}

		flag.IntVar(&instance.port, "port", 80, "Server Port")
		flag.StringVar(&instance.secrets, "secrets", os.Getenv("JWT_SECRETS"), "256 bytes of secrets")
		flag.StringVar(&instance.db.dsn, "db-dsn", os.Getenv("DB_DSN"), "PostgreSQL DSN")

		flag.Parse()
	})

	return instance
}
