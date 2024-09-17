package handlers

import (
	"context"
	"errors"
	"fmt"
	pbAuth "github.com/Noah-Wilderom/video-streaming/api-gateway/proto/auth"
	pb "github.com/Noah-Wilderom/video-streaming/api-gateway/proto/stream"
	"github.com/labstack/echo/v4"
	"io"
	"net/http"
	"time"
)

func (h *Handler) NewStream(c echo.Context) error {
	videoId := c.Param("videoId")
	fmt.Println("videoId:", videoId)

	if videoId == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "video id is empty"})
	}

	userData := c.Get("user")
	user, ok := userData.(*pbAuth.User)
	if !ok {
		return errors.New("invalid user")
	}

	streamCtx, streamCancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
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

func (h *Handler) StreamSegment(c echo.Context) error {
	token := c.QueryParam("token")
	if token == "" {
		return c.String(http.StatusBadRequest, "token is required")
	}

	videoId := c.Param("videoId")
	if videoId == "" {
		return c.String(http.StatusBadRequest, "video id is required")
	}

	userData := c.Get("user")
	user, ok := userData.(*pbAuth.User)
	if !ok {
		return errors.New("invalid user")
	}

	streamCtx, streamCancel := context.WithTimeout(c.Request().Context(), 15*time.Second)
	defer streamCancel()

	stream, err := h.Stream.StreamSegment(streamCtx, &pb.StreamSegmentRequest{
		Token:   token,
		UserId:  user.Id,
		VideoId: videoId,
	})

	if err != nil {
		fmt.Println("stream err", err.Error())
		return err
	}

	c.Response().Header().Set("Content-Type", "video/MP2T")
	c.Response().Header().Set("Transfer-Encoding", "chunked")

	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Error streaming segment")
		}

		// Write the segment data to the client
		if _, err := c.Response().Write(res.Content); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to write data")
		}
	}

	return nil
}
