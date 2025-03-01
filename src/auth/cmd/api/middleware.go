package main

import (
	"github.com/gin-gonic/gin"
	"github.com/ziliscite/video-to-mp3/auth/pkg/token"
	"net/http"
	"strings"
)

func (app *application) auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if len(authHeader) < 7 || strings.ToLower(authHeader[0:6]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		accessToken := strings.TrimSpace(authHeader[7:])
		if accessToken == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		id, email, isAdmin, err := token.Validate(accessToken, app.cfg.secrets)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		c.Set("id", id)
		c.Set("email", email)
		c.Set("isAdmin", isAdmin)

		c.Next()
	}
}
