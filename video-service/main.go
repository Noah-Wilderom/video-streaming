package main

import (
	"github.com/Noah-Wilderom/video-streaming/video-service/handlers"
	"github.com/Noah-Wilderom/video-streaming/video-service/migrations"
	"github.com/Noah-Wilderom/video-streaming/video-service/proto/video"
	"gofr.dev/pkg/gofr"
)

type Handler struct {
}

func main() {

	app := gofr.New()

	app.Migrate(migrations.All())

	video.RegisterVideoServiceServer(app, &handlers.VideoHandler{})

	app.Run()
}
