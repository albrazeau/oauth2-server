package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"server/controller"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

func main() {

	crtl, err := controller.NewController()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := crtl.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	argLen := len(os.Args)
	switch {
	case argLen > 1:
		if os.Args[1] == "standup" {
			if err := crtl.Standup(); err != nil {
				log.Fatal(err)
			}
		} else {
			serve(crtl)
		}
	default:
		serve(crtl)
	}

}

func serve(crtl *controller.Controller) {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Logger.SetLevel(log.INFO)

	// Start server
	go func() {
		if err := e.Start(":5000"); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
	}()

	e.GET("/health", crtl.Health)
	e.POST("/token", crtl.Token)

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	// Use a buffered channel to avoid missing signals as recommended for signal.Notify
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
