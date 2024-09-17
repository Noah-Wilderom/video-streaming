package main

import (
	"fmt"
	"github.com/Noah-Wilderom/video-streaming/api-gateway/handlers"
	"github.com/Noah-Wilderom/video-streaming/api-gateway/middlewares"
	"github.com/Noah-Wilderom/video-streaming/api-gateway/proto/auth"
	"github.com/Noah-Wilderom/video-streaming/api-gateway/proto/stream"
	"github.com/Noah-Wilderom/video-streaming/api-gateway/proto/video"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"os"
)

func main() {

	e := echo.New()

	authHandler, authConn := createAuthHandler()
	defer authConn.Close()

	videoHandler, videoConn := createVideoHandler()
	defer videoConn.Close()

	streamHandler, streamConn := createStreamHandler()
	defer streamConn.Close()

	handler := handlers.NewHandler(authHandler, videoHandler, streamHandler)
	middlewareHandler := middlewares.NewHandler(authHandler)

	e.POST("/auth/login", handler.Login)
	e.POST("/auth/register", handler.Register)
	e.POST("/auth/check", handler.Check, middlewareHandler.Authenticated)

	videoGroup := e.Group("/video")
	videoGroup.Use(middlewareHandler.Authenticated)
	videoGroup.POST("/upload", handler.Upload)

	streamGroup := e.Group("/stream")
	streamGroup.Use(middlewareHandler.Authenticated)
	streamGroup.GET("/new/:videoId", handler.NewStream)

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%s", os.Getenv("APP_PORT"))))
}

func init() {
	err := godotenv.Load("/app/configs/.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func createAuthHandler() (auth.AuthServiceClient, *grpc.ClientConn) {
	conn, err := grpc.NewClient(os.Getenv("AUTH_SERVICE_HOST"), grpc.WithTransportCredentials(
		insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(5*1024*1024*1024*1024)),
	)
	if err != nil {
		panic(fmt.Errorf("did not connect: %s", err))
	}

	return auth.NewAuthServiceClient(conn), conn
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

func createStreamHandler() (stream.StreamingServiceClient, *grpc.ClientConn) {
	conn, err := grpc.NewClient(os.Getenv("STREAM_SERVICE_HOST"), grpc.WithTransportCredentials(
		insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(5*1024*1024*1024*1024)),
	)
	if err != nil {
		panic(fmt.Errorf("did not connect: %s", err))
	}

	return stream.NewStreamingServiceClient(conn), conn
}
