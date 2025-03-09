package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/ziliscite/video-to-mp3/gateway/internal/domain"
	"net/http"
)

const maxSize = 1 << 29 // 512 MB

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

	video, contentType, err := app.extractFile(file)
	if err != nil {
		app.serverError(c)
		return
	}
	defer video.Close()

	if contentType != "video/mp4" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid video format"})
		return
	}

	// store to s3 here
	key, err := app.fs.UploadVideo(c.Request.Context(), file.Size, file.Filename, app.cfg.aws.s3Bucket, video)
	if err != nil {
		app.serverError(c)
		return
	}

	user, err := app.extractUser(c)
	if err != nil {
		// cuz previously we authorized it, then now the error is internal
		app.serverError(c)
		return
	}

	// send S3 video name, key, and user id to converter via rabbitmq.
	// after the video is converted,
	// the metadata (name, key, user id) will be stored in the database
	// with the mp3 key as well, maybe with status

	if err = app.fp.PublishVideo(c.Request.Context(), &domain.Video{
		UserId: user.ID, UserEmail: user.Email, FileName: file.Filename,
		FileSize: file.Size, FileKey: key,
	}); err != nil {

		app.background(func() {
			_ = app.fs.DeleteVideo(c.Request.Context(), app.cfg.aws.s3Bucket, key)
		})

		app.serverError(c)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("video has been uploaded, you will be notified through %s soon", user.Email),
		// the other service doesn't need file url
		"video_url": app.fileUrl(key, app.cfg.aws.s3Bucket),
	})
}
