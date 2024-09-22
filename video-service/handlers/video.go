package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Noah-Wilderom/video-streaming/video-service/models"
	pb "github.com/Noah-Wilderom/video-streaming/video-service/proto/video"
	"github.com/google/uuid"
	"gofr.dev/pkg/gofr/container"
	"google.golang.org/protobuf/types/known/timestamppb"
	"io"
	"os"
	"path/filepath"
	"time"
)

type VideoHandler struct {
	*container.Container
	pb.UnimplementedVideoServiceServer
}

func (h *VideoHandler) GetById(ctx context.Context, req *pb.GetByIdRequest) (*pb.Video, error) {
	row := h.SQL.QueryRowContext(ctx, "SELECT * FROM videos WHERE id = ?", req.GetId())
	if err := row.Err(); err != nil {
		fmt.Println("error getting videos:", err.Error())
		return nil, err
	}

	video, metadata, err := models.ScanToVideo(row)
	if err != nil {
		return nil, err
	}

	return &pb.Video{
		Id:       video.Id,
		UserId:   video.UserId,
		Status:   pb.UploadStatus_Done,
		Path:     video.Path,
		Size:     video.Size,
		Mimetype: video.MimeType,
		Metadata: &pb.Metadata{
			Resolution: metadata.Resolution,
			Duration:   uint64(metadata.Duration),
			Format:     metadata.Format,
			Codec:      metadata.Codec,
			Bitrate:    int64(metadata.Bitrate),
		},
		ProcessedAt: timestamppb.New(*video.ProcessedAt),
		CreatedAt:   timestamppb.New(video.CreatedAt),
		UpdatedAt:   timestamppb.New(video.UpdatedAt),
	}, nil
}

func (h *VideoHandler) GetAll(ctx context.Context, req *pb.GetAllRequest) (*pb.GetAllResponse, error) {
	var (
		videos []*pb.Video
	)

	rows, err := h.SQL.QueryContext(ctx, "SELECT * FROM videos WHERE user_id = ? ORDER BY created_at DESC", req.GetUserId())
	if err != nil {
		if errors.Is(sql.ErrNoRows, err) {
			return &pb.GetAllResponse{Videos: videos}, nil
		}

		fmt.Println("error on query:", err.Error())
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		video, metadata, err := models.ScanToVideo(rows)
		if err != nil {
			fmt.Println("error on scanning:", err.Error())
			return nil, err
		}

		videoStatus := pb.UploadStatus_Uploaded
		switch video.Status {
		case "processed":
			videoStatus = pb.UploadStatus_Done
		}

		videos = append(videos, &pb.Video{
			Id:       video.Id,
			UserId:   video.UserId,
			Status:   videoStatus,
			Path:     video.Path,
			Size:     video.Size,
			Mimetype: video.MimeType,
			Metadata: &pb.Metadata{
				Resolution: metadata.Resolution,
				Duration:   uint64(metadata.Duration),
				Format:     metadata.Format,
				Codec:      metadata.Codec,
				Bitrate:    int64(metadata.Bitrate),
			},
			ProcessedAt: timestamppb.New(*video.ProcessedAt),
			CreatedAt:   timestamppb.New(video.CreatedAt),
			UpdatedAt:   timestamppb.New(video.UpdatedAt),
		})
	}

	return &pb.GetAllResponse{Videos: videos}, nil
}

func (h *VideoHandler) Upload(stream pb.VideoService_UploadServer) error {
	h.Logger.Info("Upload function called")

	var (
		sizeLeft       int64 = -1
		totalSize      int64 = -1
		uploadRequest  *pb.UploadRequest
		isFirstMessage bool = true
	)

	id, err := uuid.NewV7()
	if err != nil {
		h.Logger.Error("error generating UUID", err.Error())
		return err
	}

	path := filepath.Join("/output", id.String())
	if err := os.MkdirAll(path, 0x777); err != nil {
		h.Logger.Error("failed to create directory", "path", path, "err", err)
		return errors.New("failed to create folder")
	}

	filePath := filepath.Join(path, "original.mp4")
	fo, err := os.Create(filePath)
	if err != nil {
		h.Logger.Error("failed to create file", "path", filePath, "err", err)
		return errors.New("failed to create file")
	}

	defer func() {
		if err := fo.Close(); err != nil {
			panic(err)
		}
	}()

	for {
		msg := new(pb.UploadMessage)
		err = stream.RecvMsg(msg)
		if err == io.EOF {
			if sizeLeft != 0 {
				fmt.Println("What is going on? sizeLeft:", sizeLeft)
				return nil
			}
			err = stream.SendAndClose(&pb.UploadResponse{
				Status:  pb.ChunkStatus_Ok,
				Message: "Upload completed successfully",
			})
			if err != nil {
				return errors.New("failed to send status code")
			}
			break
		}
		if err != nil {
			h.Logger.Error("error receiving msg", err.Error())
			fmt.Println("error receiving msg", err.Error())
			return err
		}

		switch payload := msg.Payload.(type) {
		case *pb.UploadMessage_UploadRequest:
			if isFirstMessage {
				uploadRequest = payload.UploadRequest

				fmt.Printf("UserId [%s] | Mimetype [%s]\n", uploadRequest.UserId, uploadRequest.Mimetype)
				isFirstMessage = false
			} else {
				fmt.Println("Should not get a upload request a second time....")
			}
		case *pb.UploadMessage_Chunk:
			if isFirstMessage {
				return errors.New("expected message of type *pb.UploadRequest as the first message")
			}

			chunk := payload.Chunk

			if sizeLeft == -1 {
				sizeLeft = int64(chunk.TotalSize)
			}

			if totalSize == -1 {
				totalSize = int64(chunk.TotalSize)
			}

			sizeLeft -= int64(chunk.Received)

			fmt.Printf("Received: [%d] | Content [%d] | Left [%d] | Totalsize [%d]\n", chunk.Received, len(chunk.Content), sizeLeft, chunk.TotalSize)
			if _, err := fo.Write(chunk.Content); err != nil {
				h.Logger.Error("error writing to file", err.Error())
				fmt.Println("error writing to file", err.Error())
				return errors.New(err.Error())
			}
		default:
			return errors.New("received unexpected message type")
		}
	}

	metadata := uploadRequest.GetMetadata()
	video, err := models.NewVideo(
		&models.Video{
			UserId:   uploadRequest.UserId,
			Status:   "uploaded",
			Path:     path,
			Size:     totalSize,
			MimeType: uploadRequest.Mimetype,
		},
		&models.Metadata{
			Resolution: metadata.Resolution,
			Duration:   int(metadata.Duration),
			Format:     metadata.Format,
			Codec:      metadata.Codec,
			Bitrate:    int(metadata.Bitrate),
		},
	)

	if err != nil {
		h.Logger.Error(err.Error())
		return err
	}

	createCtx, createCancel := context.WithTimeout(stream.Context(), 2*time.Second)
	defer createCancel()
	_, err = h.SQL.ExecContext(
		createCtx,
		"INSERT INTO videos (id, user_id, status, path, size, mimetype, metadata) VALUES (?, ?, ?, ?, ?, ?, ?)",
		video.Id,
		video.UserId,
		video.Status,
		video.Path,
		video.Size,
		video.MimeType,
		video.Metadata,
	)

	if err != nil {
		h.Logger.Error("error creating record", err.Error())
		return err
	}

	videoData, err := json.Marshal(video)
	if err != nil {
		h.Logger.Error("error marshalling data", err.Error())
		return err
	}

	if err := h.GetPublisher().Publish(context.Background(), "video-processing", videoData); err != nil {
		h.Logger.Error("error publishing", err.Error())
		return err
	}

	return nil
}
