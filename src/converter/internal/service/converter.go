package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ziliscite/video-to-mp3/converter/internal/domain"
	"github.com/ziliscite/video-to-mp3/converter/internal/repository"
	"github.com/ziliscite/video-to-mp3/converter/pkg/encryptor"
)

type ConverterMP4 interface {
	// ConvertMP4 converts the video to mp3 format.
	// takes the user id, file size, filename, and file key as arguments.
	// returns the audio key and an error if any.
	ConvertMP4(ctx context.Context, userId, filesize int64, filename, filekey string) (string, error)
}

type ConverterService interface {
	ConverterMP4
	// ConverterMP3
	// ConverterText
}

type bucket struct {
	mp4 string
	mp3 string
}

type converterService struct {
	cv *domain.Converter
	fr repository.FileStore
	mr repository.MetadataRepository
	en *encryptor.Encryptor
	b  bucket
}

func NewConverterService(cv *domain.Converter, fr repository.FileStore, en *encryptor.Encryptor, mp4Bucket, mp3Bucket string) ConverterService {
	return &converterService{
		cv: cv,
		fr: fr,
		en: en,
		b: bucket{
			mp4: mp4Bucket,
			mp3: mp3Bucket,
		},
	}
}

func (c *converterService) ConvertMP4(ctx context.Context, userId, filesize int64, filename, filekey string) (string, error) {
	// get the video file from S3
	video, err := c.read(ctx, filekey, filesize)
	if err != nil {
		return "", fmt.Errorf("failed to read video file: %v", err)
	}
	defer video.Close()

	// convert the video to mp3
	out, err := c.cv.ConvertMP4ToMP3(filename, video)
	if err != nil {
		return "", fmt.Errorf("failed to convert video: %v", err)
	}
	defer os.Remove(out)

	// encrypt and store the mp3
	audioKey, err := c.storeMP3(ctx, out)
	if err != nil {
		return "", fmt.Errorf("failed to process and store mp3: %v", err)
	}

	// if all is well, save the metadata to the database;
	if err = c.saveMetadata(ctx, &domain.Metadata{
		UserId: userId, FileName: filename, VideoKey: filekey, AudioKey: audioKey,
	}); err != nil {
		return "", fmt.Errorf("failed to save metadata: %v", err)
	}

	return audioKey, nil
}

func (c *converterService) storeMP3(ctx context.Context, mp3Path string) (string, error) {
	// open the converted file
	mp3, err := os.Open(mp3Path)
	if err != nil {
		return "", fmt.Errorf("failed to open converted file: %v", err)
	}
	defer mp3.Close()

	// encrypt the converted file
	enc, err := c.en.Encrypt(mp3Path)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt converted file: %v", err)
	}

	ext := filepath.Ext(mp3Path)
	key := fmt.Sprintf("%s.%s", enc, ext)

	// save the encrypted file to S3
	if err = c.fr.Save(ctx, key, c.mime(ext), c.b.mp3, mp3); err != nil {
		return "", fmt.Errorf("failed to save converted file: %v", err)
	}

	return key, nil
}

func (c *converterService) saveMetadata(ctx context.Context, data *domain.Metadata) error {
	if err := c.mr.Insert(ctx, data); err != nil {
		switch {
		case errors.Is(err, repository.ErrDuplicateEntry):
			return fmt.Errorf("metadata already exists: %v", err)
		default:
			return fmt.Errorf("failed to save metadata: %v", err)
		}
	}

	return nil
}

func (c *converterService) read(ctx context.Context, filekey string, filesize int64) (io.ReadCloser, error) {
	const threshold = 1 << 26 // 64MB

	var (
		video io.ReadCloser
		err   error
	)

	if filesize > threshold {
		video, err = c.fr.ReadLarge(ctx, c.b.mp4, filekey)
	} else {
		video, err = c.fr.Read(ctx, c.b.mp4, filekey)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to read video file: %v", err)
	}

	return video, nil
}

//func (c *converterService) fileSize(file *os.File) (int64, error) {
//	fileInfo, err := file.Stat()
//	if err != nil {
//		return 0, fmt.Errorf("failed to get file info: %v", err)
//	}
//
//	size := fileInfo.Size()
//	if size == 0 {
//		return 0, fmt.Errorf("converted file is empty")
//	}
//
//	return size, nil
//}

func (c *converterService) mime(ext string) string {
	switch ext {
	case ".aac":
		return "audio/aac"
	case ".mp3":
		return "audio/mpeg"
	case ".wav":
		return "audio/wav"
	// should not happen. as in convert, we only support these 3 formats
	default:
		return ""
	}
}
