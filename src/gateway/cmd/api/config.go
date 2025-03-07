package main

import (
	"flag"
	"fmt"
	"os"
	"sync"
)

type AWS struct {
	s3Bucket                 string
	s3Region                 string
	s3CloudFrontDistribution string
	accessKeyId              string
	secretAccessKey          string
}

type RabbitMQ struct {
	host     string
	username string
	password string
	port     string
}

func (r RabbitMQ) dsn() string {
	return fmt.Sprintf("amqps://%s:%s@%s:%s", r.username, r.password, r.host, r.port)
}

type Address struct {
	auth string
}

type Config struct {
	port       int
	encryptKey string
	secrets    string
	addr       Address
	aws        AWS
	rabbit     RabbitMQ
}

var (
	instance Config
	once     sync.Once
)

func getConfig() Config {
	once.Do(func() {
		//if err := godotenv.Load(".env.dev"); err != nil {
		//	slog.Error("Error loading .env file", "error", err)
		//	os.Exit(1)
		//}

		instance = Config{}

		flag.IntVar(&instance.port, "port", 8080, "Server Port")

		flag.StringVar(&instance.secrets, "secrets", os.Getenv("JWT_SECRETS"), "256 bytes of secrets")
		flag.StringVar(&instance.encryptKey, "key", os.Getenv("ENCRYPT_KEY"), "Encryption key")

		flag.StringVar(&instance.addr.auth, "auth-addr", os.Getenv("JWT_SECRETS"), "Authentication Service Address")

		flag.StringVar(&instance.aws.s3Bucket, "s3-bucket", os.Getenv("S3_BUCKET"), "S3 bucket name")
		flag.StringVar(&instance.aws.s3Region, "s3-region", os.Getenv("S3_REGION"), "S3 region")
		flag.StringVar(&instance.aws.s3CloudFrontDistribution, "s3-cf", os.Getenv("S3_CLOUDFRONT_DISTRIBUTION"), "S3 CloudFront distribution ID")
		flag.StringVar(&instance.aws.accessKeyId, "aws-access-key-id", os.Getenv("AWS_ACCESS_KEY_ID"), "AWS access key ID")
		flag.StringVar(&instance.aws.secretAccessKey, "aws-secret-access-key", os.Getenv("AWS_SECRET_ACCESS_KEY"), "AWS secret access key")

		flag.StringVar(&instance.rabbit.host, "rabbit-host", os.Getenv("AMQP_HOST"), "RabbitMQ host")
		flag.StringVar(&instance.rabbit.username, "rabbit-username", os.Getenv("AMQP_USERNAME"), "RabbitMQ username")
		flag.StringVar(&instance.rabbit.password, "rabbit-password", os.Getenv("AMQP_PASSWORD"), "RabbitMQ password")
		flag.StringVar(&instance.rabbit.port, "rabbit-port", os.Getenv("AMQP_PORT"), "RabbitMQ password")

		flag.Parse()
	})

	return instance
}
