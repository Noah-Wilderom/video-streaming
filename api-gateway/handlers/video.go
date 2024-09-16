package handlers

import (
	"context"
	"fmt"
	pb "github.com/Noah-Wilderom/video-streaming/api-gateway/proto/video"
	"github.com/labstack/echo/v4"
	"io"
	"net/http"
	"time"
)

func (h *Handler) Upload(c echo.Context) error {
	fileSize := c.Request().ContentLength

	if fileSize <= 0 {
		fmt.Println("Content-Length from request:", c.Request().ContentLength)
		fmt.Println("Content-Length from header:", c.Request().Header.Get("Content-Length"))
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": "Content-Length header missing or invalid",
		})
	}

	maxFileSize := int64(5 * 1024 * 1024 * 1024) // 5GB limit
	limitedReader := io.LimitReader(c.Request().Body, maxFileSize)

	err := h.streamToGRPC(limitedReader, fileSize)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	return c.JSON(http.StatusOK, nil)
}

func (h *Handler) streamToGRPC(file io.Reader, totalSize int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	stream, err := h.Video.Upload(ctx)
	if err != nil {
		return err
	}

	defer stream.CloseSend()

	bufferLength := 1024 * 1024
	buffer := make([]byte, bufferLength) // 1MB buffer for each chunk
	i := 0
	for {
		i++
		received := i * bufferLength
		n, err := file.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if bufferLength > n {
			received = int(totalSize)
		}

		// Send each chunk to the gRPC server
		req := &pb.Chunk{
			Content:   buffer[:n],
			TotalSize: uint64(totalSize),
			Received:  uint64(received),
		}
		err = stream.Send(req)
		if err != nil {
			return err
		}

	}

	// Close the stream and get the response
	status, err := stream.CloseAndRecv()
	if err != nil {
		return err
	}

	if status.Status != pb.ChunkStatus_Ok {
		return fmt.Errorf("upload failed - msg: %s", status.Message)
	}

	return nil
}
