package repository

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"mime/multipart"
)

type Write interface {
	// Save saves the file to S3 storage and returns the file key.
	// Filekey is the encrypted filename.
	// Types is the MIME content type of the file.
	// Bucket is the bucket name where the file will be saved.
	Save(ctx context.Context, fileKey, types, bucket string, file multipart.File) (string, error)
	Delete(signedUrl string) error
}

type Read interface {
	Get(signedUrl string) (string, error)
}

type ImageStore interface {
	Write
	Read
}

type store struct {
	s3c *s3.Client
}

func NewStore(s3c *s3.Client) ImageStore {
	return &store{
		s3c: s3c,
	}
}

func (s *store) Save(ctx context.Context, fileKey, types, bucket string, file multipart.File) (string, error) {
	if _, err := s.s3c.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(fileKey),
		Body:        file,
		ContentType: aws.String(types),
	}); err != nil {
		return "", fmt.Errorf("cannot upload image: %w", err)
	}

	return fileKey, nil
}

func (s *store) Get(signedUrl string) (string, error) {
	return "", nil
}

func (s *store) Delete(signedUrl string) error {
	return nil
}
