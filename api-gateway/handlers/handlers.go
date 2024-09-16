package handlers

import (
	"github.com/Noah-Wilderom/video-streaming/api-gateway/proto/auth"
	"github.com/Noah-Wilderom/video-streaming/api-gateway/proto/video"
)

type Handler struct {
	Auth  auth.AuthServiceClient
	Video video.VideoServiceClient
}

func NewHandler(authHandler auth.AuthServiceClient, videoHandler video.VideoServiceClient) *Handler {
	return &Handler{
		Auth:  authHandler,
		Video: videoHandler,
	}
}
