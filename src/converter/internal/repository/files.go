package repository

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

var (
	ErrNotExist = fmt.Errorf("file does not exist")
)

var partSize int64 = 10 << 20 // 10 MB

type FileWriter interface {
	// Save saves the file to S3 storage and returns the file key.
	// Filekey is the encrypted filename.
	// Types is the MIME content type of the file.
	// Bucket is the bucket name where the file will be saved.
	Save(ctx context.Context, fileKey, types, bucket string, file io.Reader) error
}

type FileReader interface {
	// Read reads the file from the bucket.
	Read(ctx context.Context, bucket string, fileKey string) (io.ReadCloser, error)
	ReadLarge(ctx context.Context, bucket string, fileKey string) (io.ReadCloser, error)
}

type FileDeleter interface {
	// Delete removes the file from the bucket.
	Delete(ctx context.Context, bucket string, fileKey string) error
}

type FileStore interface {
	FileWriter
	FileReader
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

func (s *store) Read(ctx context.Context, bucket string, fileKey string) (io.ReadCloser, error) {
	result, err := s.s3c.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(fileKey),
	})

	if err != nil {
		var noKey *types.NoSuchKey
		errors.As(err, &noKey)
		switch {
		case errors.As(err, &noKey):
			return nil, ErrNotExist
		default:
			return nil, fmt.Errorf("failed to read object %s from bucket %s: %w", fileKey, bucket, err)
		}
	}

	return result.Body, nil
}

// ReadLarge uses a download manager to download an object from a bucket.
// The download manager gets the data in parts and writes them to a buffer until all of
// the data has been downloaded.
func (s *store) ReadLarge(ctx context.Context, bucket string, fileKey string) (io.ReadCloser, error) {
	downloader := manager.NewDownloader(s.s3c, func(d *manager.Downloader) {
		d.PartSize = partSize
	})

	buffer := manager.NewWriteAtBuffer([]byte{})

	if _, err := downloader.Download(ctx, buffer, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(fileKey),
	}); err != nil {
		return nil, fmt.Errorf("failed to download object %s from bucket %s: %w", fileKey, bucket, err)
	}

	return io.NopCloser(bytes.NewReader(buffer.Bytes())), nil
}
