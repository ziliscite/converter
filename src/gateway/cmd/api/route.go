package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *application) route() *gin.Engine {
	router := gin.Default()

	v1 := router.Group("/v1")

	v1.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	v1.POST("/register", app.register)
	v1.POST("/login", app.login)

	authenticated := v1.Group("/", app.auth())
	// get

	admin := authenticated.Group("/", app.admin())
	admin.POST("/upload", app.upload)

	//v1.POST("/upload", app.upload)

	return router
}

func (app *application) run() error {
	srv := &http.Server{
		Addr:         fmt.Sprintf("0.0.0.0:%d", app.cfg.port),
		Handler:      app.route(),
		ReadTimeout:  5 * time.Minute,
		WriteTimeout: 30 * time.Second,
	}

	shutdownError := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		// Wait for the signal
		<-quit

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			shutdownError <- err
		}

		slog.Info("Closing background task")

		app.wg.Wait()
		shutdownError <- nil
	}()

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	if err := <-shutdownError; err != nil {
		return err
	}

	slog.Info("Server exiting")
	return nil
}
