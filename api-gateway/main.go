package main

import (
	"github.com/Noah-Wilderom/video-streaming/api-gateway/handlers"
	"gofr.dev/pkg/gofr"
)

func main() {

	app := gofr.New()

	app.POST("/auth/login", handlers.Login)
	app.POST("/auth/register", handlers.Register)
	app.POST("/auth/check", handlers.Check)

	app.Run()
}
