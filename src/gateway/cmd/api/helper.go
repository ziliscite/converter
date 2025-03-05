package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
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
