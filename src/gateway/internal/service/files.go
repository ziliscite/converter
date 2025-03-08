package service

import (
	"context"
	"fmt"
	"github.com/ziliscite/video-to-mp3/gateway/internal/repository"
	"github.com/ziliscite/video-to-mp3/gateway/pkg/encryptor"
	"io"
)

type FileService interface {
	// UploadVideo saves video to the storage and returns the file key.
	// Filename is the original filename of the file.
	// Bucket is the bucket name where the file will be saved.
	UploadVideo(ctx context.Context, filesize int64, filename, bucket string, file io.Reader) (string, error)
	DeleteVideo(ctx context.Context, bucket, fileKey string) error
}

type fileService struct {
	en *encryptor.Encryptor
	wr repository.FileStore
}

func NewFileService(en *encryptor.Encryptor, r repository.FileStore) FileService {
	return &fileService{
		en: en,
		wr: r,
	}
}

func (u *fileService) UploadVideo(ctx context.Context, filesize int64, filename, bucket string, file io.Reader) (string, error) {
	fileKey, err := u.en.Encrypt(filename)
	if err != nil {
		return "", fmt.Errorf("cannot encrypt image url: %w", err)
	}

	const threshold = 1 << 26 // 64MB
	if filesize > threshold {
		return fileKey, u.wr.SaveLarge(ctx, fmt.Sprintf("%s.mp4", fileKey), "video/mp4", bucket, file)
	}

	return fileKey, u.wr.Save(ctx, fmt.Sprintf("%s.mp4", fileKey), "video/mp4", bucket, file)
}

func (u *fileService) DeleteVideo(ctx context.Context, bucket, fileKey string) error {
	return u.wr.Delete(ctx, bucket, fileKey)
}
