package main

import (
	"github.com/Noah-Wilderom/video-streaming/user-service/handlers"
	"github.com/Noah-Wilderom/video-streaming/user-service/migrations"
	"github.com/Noah-Wilderom/video-streaming/user-service/proto/auth"
	"gofr.dev/pkg/gofr"
)

type Handler struct {
}

func main() {

	app := gofr.New()

	app.Migrate(migrations.All())

	auth.RegisterAuthServiceServer(app, &handlers.AuthHandler{})

	app.Run()
}
