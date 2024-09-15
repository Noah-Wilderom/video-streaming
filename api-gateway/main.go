package main

import (
	"fmt"
	"github.com/Noah-Wilderom/video-streaming/api-gateway/handlers"
	"github.com/Noah-Wilderom/video-streaming/api-gateway/middlewares"
	"github.com/Noah-Wilderom/video-streaming/api-gateway/proto/auth"
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

	handler := handlers.NewHandler(authHandler)
	middlewareHandler := middlewares.NewHandler(authHandler)

	e.POST("/auth/login", handler.Login)
	e.POST("/auth/register", handler.Register)
	e.POST("/auth/check", handler.Check, middlewareHandler.Authenticated)

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%s", os.Getenv("APP_PORT"))))
}

func init() {
	err := godotenv.Load("/app/configs/.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func createAuthHandler() (auth.AuthServiceClient, *grpc.ClientConn) {
	conn, err := grpc.NewClient(os.Getenv("AUTH_SERVICE_HOST"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(fmt.Errorf("did not connect: %s", err))
	}

	return auth.NewAuthServiceClient(conn), conn
}
