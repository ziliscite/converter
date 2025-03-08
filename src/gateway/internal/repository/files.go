package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"io"
	"time"
)

var (
	ErrNotExist = fmt.Errorf("file does not exist")
)

type FileWriter interface {
	// Save saves the file to S3 storage and returns the file key.
	// Filekey is the encrypted filename.
	// Types is the MIME content type of the file.
	// Bucket is the bucket name where the file will be saved.
	Save(ctx context.Context, fileKey, types, bucket string, file io.Reader) error
	SaveLarge(ctx context.Context, fileKey, types, bucket string, file io.Reader) error
}

type FileDeleter interface {
	// Delete removes the file from the bucket.
	Delete(ctx context.Context, bucket string, fileKey string) error
}

type FileStore interface {
	FileWriter
	FileDeleter
}

type store struct {
	s3c *s3.Client
}

func NewStore(s3c *s3.Client) FileStore {
	return &store{
		s3c: s3c,
	}
}

// Save saves the file to an object in a bucket.
func (s *store) Save(ctx context.Context, fileKey, types, bucket string, file io.Reader) error {
	if _, err := s.s3c.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(fileKey),
		Body:        file,
		ContentType: aws.String(types),
	}); err != nil {
		return fmt.Errorf("failed to upload file %s to bucket %s: %w", fileKey, bucket, err)
	}

	if err := s3.NewObjectExistsWaiter(s.s3c).Wait(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(fileKey),
	}, time.Minute); err != nil {
		return fmt.Errorf("failed to confirm existence of uploaded file %s in bucket %s: %w", fileKey, bucket, err)
	}

	return nil
}

// SaveLarge uses an upload manager to upload data to an object in a bucket.
// The upload manager breaks large data into parts and uploads the parts concurrently.
func (s *store) SaveLarge(ctx context.Context, fileKey, types, bucket string, file io.Reader) error {
	var size int64 = 10 << 20 // 10 MB
	uploader := manager.NewUploader(s.s3c, func(u *manager.Uploader) {
		u.PartSize = size
	})

	if _, err := uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(fileKey),
		Body:        file,
		ContentType: aws.String(types),
	}); err != nil {
		var apiErr smithy.APIError
		errors.As(err, &apiErr)

		switch {
		case apiErr.ErrorCode() == "EntityTooLarge":
			return fmt.Errorf("file exceeds maximum size of 5TB for multipart upload to bucket %s: %w", bucket, err)
		default:
			return fmt.Errorf("failed to upload file %s to bucket %s: %w", fileKey, bucket, err)
		}
	}

	if err := s3.NewObjectExistsWaiter(s.s3c).Wait(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(fileKey),
	}, time.Minute); err != nil {
		return fmt.Errorf("failed to confirm existence of uploaded file %s in bucket %s: %w", fileKey, bucket, err)
	}

	return nil
}

func (s *store) Delete(ctx context.Context, bucket string, fileKey string) error {
	if _, err := s.s3c.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(fileKey),
	}); err != nil {
		var noKey *types.NoSuchKey
		errors.As(err, &noKey)
		switch {
		case errors.As(err, &noKey):
			return ErrNotExist
		default:
			return fmt.Errorf("failed to delete object %s from bucket %s: %w", fileKey, bucket, err)
		}
	}

	if err := s3.NewObjectExistsWaiter(s.s3c).Wait(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(fileKey),
	}, time.Minute); err != nil {
		return fmt.Errorf("failed attempt to wait for object %s in bucket %s to be deleted", fileKey, bucket)
	}

	return nil
}
