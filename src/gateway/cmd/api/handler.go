package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/ziliscite/video-to-mp3/gateway/internal/domain"
	"net/http"
)

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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reach the server: " + err.Error()})
		return
	}

	// Server returned a 4xx/5xx HTTP status (e.g., 400 Bad Request).
	// Resty doesn’t treat this as an `err`, so check the status code.
	if resp.IsError() {
		// Forward the remote server’s status code (e.g., 400) and body.
		c.JSON(resp.StatusCode(), gin.H{"error": "Remote server error: " + resp.String()})
		return
	}

	var auth domain.Auth
	if err = json.Unmarshal(resp.Body(), &auth); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse response: " + err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reach the server: " + err.Error()})
		return
	}

	if resp.IsError() {
		c.JSON(resp.StatusCode(), gin.H{"error": "Remote server error: " + resp.String()})
		return
	}

	var user domain.User
	if err = json.Unmarshal(resp.Body(), &user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse response: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}
