package handlers

import "github.com/Noah-Wilderom/video-streaming/api-gateway/proto/auth"

type Handler struct {
	Auth auth.AuthServiceClient
}

func NewHandler(authHandler auth.AuthServiceClient) *Handler {
	return &Handler{
		Auth: authHandler,
	}
}
