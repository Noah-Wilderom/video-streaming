package handlers

import (
	"context"
	"errors"
	"fmt"
	pbAuth "github.com/Noah-Wilderom/video-streaming/api-gateway/proto/auth"
	pb "github.com/Noah-Wilderom/video-streaming/api-gateway/proto/stream"
	"github.com/labstack/echo/v4"
	"net/http"
	"time"
)

func (h *Handler) NewStream(c echo.Context) error {
	videoId := c.QueryParam("video_id")

	userData := c.Get("user")
	user, ok := userData.(*pbAuth.User)
	if !ok {
		return errors.New("invalid user")
	}

	streamCtx, streamCancel := context.WithTimeout(c.Request().Context(), 3*time.Second)
	defer streamCancel()
	res, err := h.Stream.NewStream(streamCtx, &pb.NewStreamRequest{
		UserId:  user.Id,
		VideoId: videoId,
	})

	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	return c.JSON(http.StatusOK, map[string]string{
		"type":    "hls",
		"content": string(res.M3U8File),
	})
}
