package handlers

import (
	"errors"
	"fmt"
	pb "github.com/Noah-Wilderom/video-streaming/video-service/proto/video"
	"github.com/google/uuid"
	"gofr.dev/pkg/gofr/container"
	"io"
	"os"
)

type VideoHandler struct {
	*container.Container
	pb.UnimplementedVideoServiceServer
}

func (h *VideoHandler) Upload(stream pb.VideoService_UploadServer) error {
	id, err := uuid.NewV7()
	if err != nil {
		return err
	}

	fo, err := os.Create(fmt.Sprintf("/output/%s", id.String()))
	if err != nil {
		return errors.New("failed to create file")
	}

	defer func() {
		if err := fo.Close(); err != nil {
			panic(err)
		}
	}()

	var res *pb.Chunk
	sizeLeft := -1
	for {
		res, err = stream.Recv()
		if err == io.EOF {
			err = stream.SendAndClose(&pb.UploadResponse{
				Status: pb.ChunkStatus_Ok,
			})
			if err != nil {
				return errors.New("failed to send status code")
			}
			return nil
		}

		if sizeLeft == -1 {
			sizeLeft = int(res.TotalSize)
		}

		sizeLeft -= int(res.TotalSize)

		fmt.Printf("Received: [%d] | Content [%d] | Left [%d]\n", res.Received, len(res.Content), sizeLeft)
		if _, err := fo.Write(res.Content); err != nil {
			return errors.New(err.Error())
		}
	}
}
