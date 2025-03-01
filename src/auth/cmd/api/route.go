package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

func (app *application) run() error {
	router := gin.Default()

	v1 := router.Group("/v1")

	v1.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	v1.POST("/register", app.register)
	v1.POST("/login", app.login)

	return router.Run(fmt.Sprintf("0.0.0.0:%d", app.cfg.port))
}
