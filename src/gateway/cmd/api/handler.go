package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/ziliscite/video-to-mp3/gateway/internal/domain"
	"io"
	"net/http"
)

const maxSize = 2 << 28 // 500~MB

func (app *application) login(c *gin.Context) {
	var request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	resp, err := app.rc.R().SetHeader("Content-Type", "application/json").
		SetBody(request).
		Post(fmt.Sprintf("%s/v1/login", app.cfg.addr))
	if err != nil {
		// Network/client-side error (e.g., timeout, DNS failure).
		app.serverError(c)
		return
	}

	// Server returned a 4xx/5xx HTTP status (e.g., 400 Bad Request).
	// Resty doesn’t treat this as an `err`, so check the status code.
	if resp.IsError() {
		// Forward the remote server’s status code (e.g., 400) and body.
		c.JSON(resp.StatusCode(), gin.H{"error": resp.String()})
		return
	}

	var auth domain.Auth
	if err = json.Unmarshal(resp.Body(), &auth); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, auth)
}

func (app *application) register(c *gin.Context) {
	var request struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	resp, err := app.rc.R().SetHeader("Content-Type", "application/json").
		SetBody(request).
		Post(fmt.Sprintf("%s/v1/register", app.cfg.addr))
	if err != nil {
		app.serverError(c)
		return
	}

	if resp.IsError() {
		c.JSON(resp.StatusCode(), gin.H{"error": resp.String()})
		return
	}

	var user domain.User
	if err = json.Unmarshal(resp.Body(), &user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (app *application) upload(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)

	file, err := c.FormFile("video")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "video file is required"})
		return
	}

	video, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open file"})
		return
	}
	defer video.Close()

	// read the first 512 bytes
	buffer := make([]byte, 512)
	if _, err = video.Read(buffer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read file"})
		return
	}

	contentType := http.DetectContentType(buffer)
	if contentType != "video/mp4" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid video format"})
		return
	}

	// reset the file pointer so that s3 doesn't read the video file from bytes 513, but from 0
	if _, err = video.Seek(0, io.SeekStart); err != nil {
		app.serverError(c)
		return
	}

	// store to s3 here
	key, err := app.fs.UploadVideo(c.Request.Context(), file.Filename, app.cfg.aws.s3Bucket, video)
	if err != nil {
		app.serverError(c)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"url": app.fileUrl(key, app.cfg.aws.s3Bucket),
	})
}
