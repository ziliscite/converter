package main

import (
	"flag"
	"fmt"
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

type AWS struct {
	s3bucket struct {
		mp4 string
		mp3 string
	}
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
	queue    struct {
		video        string
		notification string
	}
}

func (r RabbitMQ) dsn() string {
	return fmt.Sprintf("amqps://%s:%s@%s:%s", r.username, r.password, r.host, r.port)
}

type Config struct {
	port       int
	encryptKey string
	db         DB
	aws        AWS
	rabbit     RabbitMQ
}

var (
	instance Config
	once     sync.Once
)

func getConfig() Config {
	once.Do(func() {
		instance = Config{}

		flag.IntVar(&instance.port, "port", 8080, "Server Port")

		flag.StringVar(&instance.encryptKey, "key", os.Getenv("ENCRYPT_KEY"), "Encryption key")

		flag.StringVar(&instance.db.host, "db-host", os.Getenv("POSTGRES_HOST"), "Database host")
		flag.StringVar(&instance.db.port, "db-port", os.Getenv("POSTGRES_PORT"), "Database port")
		flag.StringVar(&instance.db.user, "db-user", os.Getenv("POSTGRES_USER"), "Database user")
		flag.StringVar(&instance.db.pass, "db-pass", os.Getenv("POSTGRES_PASSWORD"), "Database password")
		flag.StringVar(&instance.db.db, "db-db", os.Getenv("POSTGRES_DB"), "Database name")

		flag.BoolVar(&instance.db.ssl, "db-ssl", false, "Database ssl")

		flag.StringVar(&instance.aws.s3bucket.mp4, "s3-mp4-bucket", os.Getenv("S3_MP4_BUCKET"), "S3 mp4 bucket name")
		flag.StringVar(&instance.aws.s3bucket.mp3, "s3-mp3-bucket", os.Getenv("S3_MP3_BUCKET"), "S3 mp3 bucket name")
		flag.StringVar(&instance.aws.s3Region, "s3-region", os.Getenv("S3_REGION"), "S3 region")
		flag.StringVar(&instance.aws.s3CloudFrontDistribution, "s3-cf", os.Getenv("S3_CLOUDFRONT_DISTRIBUTION"), "S3 CloudFront distribution ID")
		flag.StringVar(&instance.aws.accessKeyId, "aws-access-key-id", os.Getenv("AWS_ACCESS_KEY_ID"), "AWS access key ID")
		flag.StringVar(&instance.aws.secretAccessKey, "aws-secret-access-key", os.Getenv("AWS_SECRET_ACCESS_KEY"), "AWS secret access key")

		flag.StringVar(&instance.rabbit.host, "rabbit-host", os.Getenv("AMQP_HOST"), "RabbitMQ host")
		flag.StringVar(&instance.rabbit.username, "rabbit-username", os.Getenv("AMQP_USERNAME"), "RabbitMQ username")
		flag.StringVar(&instance.rabbit.password, "rabbit-password", os.Getenv("AMQP_PASSWORD"), "RabbitMQ password")
		flag.StringVar(&instance.rabbit.port, "rabbit-port", os.Getenv("AMQP_PORT"), "RabbitMQ password")
		flag.StringVar(&instance.rabbit.queue.video, "rabbit-vid-queue", os.Getenv("AMQP_VIDEO_QUEUE_NAME"), "RabbitMQ video queue")
		flag.StringVar(&instance.rabbit.queue.notification, "rabbit-notif-queue", os.Getenv("AMQP_NOTIFICATION_QUEUE_NAME"), "RabbitMQ notification queue")

		flag.Parse()
	})

	return instance
}
