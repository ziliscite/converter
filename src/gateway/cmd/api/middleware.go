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
			SetHeaders(map[string]string{
				"Content-Type":  "application/json",
				"Authorization": c.GetHeader("Authorization"),
			}).
			Post(fmt.Sprintf("%s/v1/validate", app.cfg.addr))
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
			c.JSON(http.StatusUnauthorized, gin.H{"error": "failed to parse token"})
			return
		}

		c.Set("user", user)
		c.Next()
	}
}

func (app *application) admin() gin.HandlerFunc {
	return func(c *gin.Context) {
		userCtx, ok := c.Get("user")
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "invalid user"})
			return
		}

		user := userCtx.(domain.User)

		if !user.IsAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "permission denied"})
			return
		}

		c.Next()
	}
}
