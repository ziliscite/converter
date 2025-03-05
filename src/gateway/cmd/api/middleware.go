package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/ziliscite/video-to-mp3/gateway/internal/domain"
	"net/http"
)

func (app *application) auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		resp, err := app.rc.R().
			SetHeader("Content-Type", "application/json").
			SetHeader("Authorization", c.GetHeader("Authorization")).
			Post(fmt.Sprintf("%s/v1/validate", app.cfg.addr))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
			return
		}

		if resp.IsError() {
			c.JSON(resp.StatusCode(), gin.H{"error": resp.String()})
			return
		}

		var user domain.User
		if err = json.Unmarshal(resp.Body(), &user); err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "failed to parse token"})
			return
		}

		c.Set("user", user)
		c.Next()
	}
}
