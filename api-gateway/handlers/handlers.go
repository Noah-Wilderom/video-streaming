package handlers

import (
	"github.com/Noah-Wilderom/video-streaming/api-gateway/proto/auth"
	"github.com/Noah-Wilderom/video-streaming/api-gateway/proto/stream"
	"github.com/Noah-Wilderom/video-streaming/api-gateway/proto/video"
)

type Handler struct {
	Auth   auth.AuthServiceClient
	Video  video.VideoServiceClient
	Stream stream.StreamingServiceClient
}

func NewHandler(authHandler auth.AuthServiceClient, videoHandler video.VideoServiceClient, streamHandler stream.StreamingServiceClient) *Handler {
	return &Handler{
		Auth:   authHandler,
		Video:  videoHandler,
		Stream: streamHandler,
	}
}
