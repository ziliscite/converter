package main

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/ziliscite/video-to-mp3/auth/internal/service"
	"github.com/ziliscite/video-to-mp3/auth/pkg/token"
	"net/http"
)

func (app *application) register(c *gin.Context) {
	var request struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.Bind(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := app.us.SignUp(c, app.v, request.Username, request.Email, request.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidUser):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": app.v.Errors()})
		case errors.Is(err, service.ErrDuplicateMail):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
	})
}

func (app *application) login(c *gin.Context) {
	var request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.Bind(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := app.us.SignIn(c, request.Email, request.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	accessToken, exp, err := token.Create(user.ID, user.IsAdmin, user.Email, app.cfg.secrets)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": accessToken,
		"exp":          exp,
		"is_admin":     user.IsAdmin,
	})
}
