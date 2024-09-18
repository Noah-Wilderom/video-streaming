package handlers

import (
	"context"
	"errors"
	"fmt"
	pbAuth "github.com/Noah-Wilderom/video-streaming/api-gateway/proto/auth"
	pb "github.com/Noah-Wilderom/video-streaming/api-gateway/proto/video"
	"github.com/labstack/echo/v4"
	"io"
	"net/http"
	"time"
)

//func (h *Handler) GetVideo(c echo.Context) error {
//	videoId := c.Param("id")
//
//}

func (h *Handler) Upload(c echo.Context) error {
	fileSize := c.Request().ContentLength

	if fileSize <= 0 {
		fmt.Println("Content-Length from request:", c.Request().ContentLength)
		fmt.Println("Content-Length from header:", c.Request().Header.Get("Content-Length"))
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": "Content-Length header missing or invalid",
		})
	}

	start := time.Now()

	maxFileSize := int64(15 * 1024 * 1024 * 1024) // 15GB limit
	limitedReader := io.LimitReader(c.Request().Body, maxFileSize)
	userData := c.Get("user")
	user, ok := userData.(*pbAuth.User)
	if !ok {
		return errors.New("invalid user")
	}

	uploadReq := &pb.UploadRequest{
		UserId:   user.Id,
		Mimetype: c.Request().Header.Get("Content-Mimetype"),
		Metadata: &pb.Metadata{
			Resolution: "",
			Duration:   0,
			Format:     "",
			Codec:      "",
			Bitrate:    0,
		},
	}

	err := h.streamToGRPC(uploadReq, limitedReader, fileSize)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": err.Error(),
		})
	}

	fmt.Printf("uploading file [%s] took: %s\n", bytesToReadableSize(fileSize), time.Since(start))
	return c.JSON(http.StatusOK, nil)
}

func (h *Handler) GetAllVideos(c echo.Context) error {
	userData := c.Get("user")
	user, ok := userData.(*pbAuth.User)
	if !ok {
		return errors.New("invalid user")
	}

	videosCtx, videosCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer videosCancel()

	videos, err := h.Video.GetAll(videosCtx, &pb.GetAllRequest{UserId: user.Id})
	if err != nil {
		fmt.Println("Error getting videos", err.Error())
		return err
	}

	return c.JSON(http.StatusOK, videos)
}

func bytesToReadableSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d bytes", bytes)
	}
}

func (h *Handler) streamToGRPC(uploadReq *pb.UploadRequest, file io.Reader, totalSize int64) error {
	//ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	//defer cancel()

	stream, err := h.Video.Upload(context.Background())
	if err != nil {
		return fmt.Errorf("failed to open gRPC stream: %w", err)
	}
	defer stream.CloseSend()

	err = stream.Send(&pb.UploadMessage{
		Payload: &pb.UploadMessage_UploadRequest{
			UploadRequest: uploadReq,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to send upload request: %w", err)
	}

	buffer := make([]byte, 1024*1024) // 1MB buffer
	sizeLeft := totalSize

	for sizeLeft > 0 {
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			return fmt.Errorf("error reading file: %w", err)
		}

		if n > 0 {
			sizeLeft -= int64(n)
			chunk := &pb.Chunk{
				Content:   buffer[:n],
				TotalSize: uint64(totalSize),
				Received:  uint64(n),
			}

			err = stream.Send(&pb.UploadMessage{
				Payload: &pb.UploadMessage_Chunk{
					Chunk: chunk,
				},
			})
			if err != nil {
				return fmt.Errorf("failed to send chunk to stream: %w", err)
			}
		}

		if err == io.EOF {
			break
		}
	}

	fmt.Printf("closing upload with %d bytes left\n", sizeLeft)

	status, err := stream.CloseAndRecv()
	if err != nil {
		return fmt.Errorf("error closing gRPC stream: %w", err)
	}

	if status.Status != pb.ChunkStatus_Ok {
		return fmt.Errorf("upload failed - status: %v, message: %s", status.Status, status.Message)
	}

	fmt.Println("upload completed successfully!")
	return nil
}
