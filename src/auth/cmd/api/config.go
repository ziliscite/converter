package main

import (
	"flag"
	"os"
	"sync"
)

type DB struct {
	host string
	port string
	user string
	pass string
	db   string
	ssl  bool
}

func (d DB) dsn() string {
	dsn := "postgres://" + d.user + ":" + d.pass + "@" + d.host + ":" + d.port + "/" + d.db
	if !d.ssl {
		return dsn + "?sslmode=disable"
	}
	return dsn
}

type Config struct {
	port    int
	secrets string
	db      DB
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

		flag.StringVar(&instance.db.host, "db-host", os.Getenv("POSTGRES_HOST"), "Database host")
		flag.StringVar(&instance.db.port, "db-port", os.Getenv("POSTGRES_PORT"), "Database port")
		flag.StringVar(&instance.db.user, "db-user", os.Getenv("POSTGRES_USER"), "Database user")
		flag.StringVar(&instance.db.pass, "db-pass", os.Getenv("POSTGRES_PASSWORD"), "Database password")
		flag.StringVar(&instance.db.db, "db-db", os.Getenv("POSTGRES_DB"), "Database name")

		flag.BoolVar(&instance.db.ssl, "db-ssl", false, "Database ssl")

		flag.Parse()
	})

	return instance
}
