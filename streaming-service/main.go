package main

import (
	"fmt"
	"github.com/Noah-Wilderom/video-streaming/streaming-service/handlers"
	"github.com/Noah-Wilderom/video-streaming/streaming-service/migrations"
	pb "github.com/Noah-Wilderom/video-streaming/streaming-service/proto/stream"
	"github.com/Noah-Wilderom/video-streaming/streaming-service/proto/video"
	"gofr.dev/pkg/gofr"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"os"
)

type Handler struct {
}

func main() {

	app := gofr.New()

	app.Migrate(migrations.All())

	videoHandler, videoConn := createVideoHandler()
	defer videoConn.Close()

	pb.RegisterStreamingServiceServer(app, &handlers.StreamHandler{Video: videoHandler})

	app.Run()
}

func createVideoHandler() (video.VideoServiceClient, *grpc.ClientConn) {
	conn, err := grpc.NewClient(os.Getenv("VIDEO_SERVICE_HOST"), grpc.WithTransportCredentials(
		insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(5*1024*1024*1024*1024)),
	)
	if err != nil {
		panic(fmt.Errorf("did not connect: %s", err))
	}

	return video.NewVideoServiceClient(conn), conn
}
