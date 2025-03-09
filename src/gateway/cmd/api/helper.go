package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/ziliscite/video-to-mp3/gateway/internal/domain"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
)

func (app *application) serverError(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, gin.H{
		"error": "something went wrong",
	})
}

func (app *application) fileUrl(key, bucket string) string {
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucket, app.cfg.aws.s3Region, key)
}

func (app *application) extractUser(c *gin.Context) (*domain.User, error) {
	userCtx, ok := c.Get("user")
	if !ok {
		return nil, fmt.Errorf("invalid user")
	}

	user, ok := userCtx.(domain.User)
	if !ok {
		return nil, fmt.Errorf("invalid user")
	}

	return &user, nil
}

// extractFile opens the multipart header and returns the file and type. You'd have to call file.Close() the file later.
func (app *application) extractFile(header *multipart.FileHeader) (multipart.File, string, error) {
	file, err := header.Open()
	if err != nil {
		return nil, "", fmt.Errorf("failed to open file")
	}

	// read the first 512 bytes
	buffer := make([]byte, 512)
	if _, err := file.Read(buffer); err != nil {
		return nil, "", fmt.Errorf("failed to read file")
	}

	contentType := http.DetectContentType(buffer)

	// reset the file pointer so that the reader doesn't read the file from bytes 513, but from 0
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, "", fmt.Errorf("failed to reset the file pointer")
	}

	return file, contentType, nil
}

// the background() helper accepts an arbitrary function as a parameter.
func (app *application) background(fn func()) {
	// increment the WaitGroup counter.
	app.wg.Add(1)

	go func() {
		// use defer to decrement the WaitGroup counter before the goroutine returns.
		defer app.wg.Done()

		// run a deferred function which uses recover() to catch any panic
		defer func() {
			if err := recover(); err != nil {
				slog.Error("panic: %v", "error", err)
			}
		}()

		fn()
	}()
}
