package main

import (
	"flag"
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

type Address struct {
	auth string
}

type Config struct {
	port    int
	secrets string
	addr    Address
	aws     AWS
}

var (
	instance Config
	once     sync.Once
)

func getConfig() Config {
	once.Do(func() {
		instance = Config{}

		flag.IntVar(&instance.port, "port", 8080, "Server Port")

		flag.StringVar(&instance.secrets, "secrets", os.Getenv("JWT_SECRETS"), "256 bytes of secrets")

		flag.StringVar(&instance.addr.auth, "auth-addr", os.Getenv("JWT_SECRETS"), "Authentication Service Address")

		flag.StringVar(&instance.aws.s3Bucket, "s3-bucket", os.Getenv("S3_BUCKET"), "S3 bucket name")
		flag.StringVar(&instance.aws.s3Region, "s3-region", os.Getenv("S3_REGION"), "S3 region")
		flag.StringVar(&instance.aws.s3CloudFrontDistribution, "s3-cf", os.Getenv("S3_CLOUDFRONT_DISTRIBUTION"), "S3 CloudFront distribution ID")
		flag.StringVar(&instance.aws.accessKeyId, "aws-access-key-id", os.Getenv("AWS_ACCESS_KEY_ID"), "AWS access key ID")
		flag.StringVar(&instance.aws.secretAccessKey, "aws-secret-access-key", os.Getenv("AWS_SECRET_ACCESS_KEY"), "AWS secret access key")

		flag.Parse()
	})

	return instance
}
