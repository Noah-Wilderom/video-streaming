package main

import (
	"github.com/Noah-Wilderom/video-streaming/streaming-service/handlers"
	"github.com/Noah-Wilderom/video-streaming/streaming-service/migrations"
	pb "github.com/Noah-Wilderom/video-streaming/streaming-service/proto/stream"
	"gofr.dev/pkg/gofr"
)

type Handler struct {
}

func main() {

	app := gofr.New()

	app.Migrate(migrations.All())

	pb.RegisterStreamingServiceServer(app, &handlers.StreamHandler{})

	app.Run()
}
