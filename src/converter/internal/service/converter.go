package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/ziliscite/video-to-mp3/converter/internal/domain"
	"github.com/ziliscite/video-to-mp3/converter/internal/repository"
	"github.com/ziliscite/video-to-mp3/converter/pkg/encryptor"
	"io"
	"os"
	"path/filepath"
)

type bucket struct {
	mp4 string
	mp3 string
}

type ConverterService struct {
	cv *domain.Converter
	fr repository.FileStore
	mr repository.MetadataRepository
	en *encryptor.Encryptor
	b  bucket
}

func NewConverterService(cv *domain.Converter, fr repository.FileStore, en *encryptor.Encryptor, mp4Bucket, mp3Bucket string) *ConverterService {
	return &ConverterService{
		cv: cv,
		fr: fr,
		en: en,
		b: bucket{
			mp4: mp4Bucket,
			mp3: mp3Bucket,
		},
	}
}

func (c ConverterService) ConvertMP4(ctx context.Context, userId, filesize int64, filename, filekey string) error {
	// get the video file from S3
	video, err := c.read(ctx, filekey, filesize)
	if err != nil {
		return fmt.Errorf("failed to read video file: %v", err)
	}
	defer video.Close()

	// convert the video to mp3
	out, err := c.cv.ConvertMP4ToMP3(filename, video)
	if err != nil {
		return fmt.Errorf("failed to convert video: %v", err)
	}
	defer os.Remove(out)

	// open the converted file
	mp3, err := os.Open(out)
	if err != nil {
		return fmt.Errorf("failed to open converted file: %v", err)
	}
	defer mp3.Close()

	// encrypt the converted file
	enc, err := c.en.Encrypt(out)
	if err != nil {
		return fmt.Errorf("failed to encrypt converted file: %v", err)
	}

	ext := filepath.Ext(out)
	key := fmt.Sprintf("%s.%s", enc, ext) // it's like, something.mp3.mp3, lol

	// save the encrypted file to S3
	if err = c.fr.Save(ctx, key, c.mime(ext), c.b.mp3, mp3); err != nil {
		return fmt.Errorf("failed to save converted file: %v", err)
	}

	// if all is well, save the metadata to the database;
	//
	// I'm conflicted whether put this (and mp3 saver) in a separate service or not
	if err = c.mr.Insert(ctx, &domain.Metadata{
		UserId: userId, FileName: filename, VideoKey: filekey, AudioKey: key,
	}); err != nil {
		switch {
		case errors.Is(err, repository.ErrDuplicateEntry):
			return fmt.Errorf("metadata already exists: %v", err)
		default:
			return fmt.Errorf("failed to save metadata: %v", err)
		}
	}

	return nil
}

func (c ConverterService) read(ctx context.Context, filekey string, filesize int64) (io.ReadCloser, error) {
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

//func (c ConverterService) fileSize(file *os.File) (int64, error) {
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

func (c ConverterService) mime(ext string) string {
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
