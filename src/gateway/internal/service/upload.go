package service

import (
	"context"
	"fmt"
	"github.com/ziliscite/video-to-mp3/gateway/internal/repository"
	"github.com/ziliscite/video-to-mp3/gateway/pkg/encryptor"
	"mime/multipart"
)

type FileService interface {
	// UploadVideo saves video to the storage and returns the file key.
	// Filename is the original filename of the file.
	// Bucket is the bucket name where the file will be saved.
	UploadVideo(ctx context.Context, filename, bucket string, file multipart.File) (string, error)
}

type fileService struct {
	en *encryptor.Encryptor
	r  repository.Write
}

func NewUploadService(en *encryptor.Encryptor, r repository.Write) FileService {
	return &fileService{
		en: en,
		r:  r,
	}
}

func (u *fileService) UploadVideo(ctx context.Context, filename, bucket string, file multipart.File) (string, error) {
	fileKey, err := u.en.Encrypt(filename)
	if err != nil {
		return "", fmt.Errorf("cannot encrypt image url: %w", err)
	}

	// Maybe saves the file metadata to mongodb

	return u.r.Save(ctx, fmt.Sprintf("%s.mp4", fileKey), "video/mp4", bucket, file)
}
