package handlers

import (
	"context"
	"errors"
	"fmt"
	"github.com/Noah-Wilderom/video-streaming/shared/token"
	"github.com/Noah-Wilderom/video-streaming/streaming-service/models"
	pb "github.com/Noah-Wilderom/video-streaming/streaming-service/proto/stream"
	pbVideo "github.com/Noah-Wilderom/video-streaming/streaming-service/proto/video"
	"github.com/Noah-Wilderom/video-streaming/streaming-service/scramble"
	"gofr.dev/pkg/gofr/container"
	"io"
	"os"
	"time"
)

type StreamHandler struct {
	*container.Container
	pb.UnimplementedStreamingServiceServer
	Video pbVideo.VideoServiceClient
}

func (h *StreamHandler) NewStream(ctx context.Context, req *pb.NewStreamRequest) (*pb.NewStreamResponse, error) {
	videoId := req.GetVideoId()
	userId := req.GetUserId()

	fmt.Printf("NEW STREAM | Video [%s] | User [%s]\n", videoId, userId)

	video, err := h.Video.GetById(ctx, &pbVideo.GetByIdRequest{Id: videoId})
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	hlsFolder := fmt.Sprintf("%s/hls", video.Path)
	scrambler := scramble.NewScrambler(&scramble.VideoProcessor{}, h.SQL, &scramble.ScramblerOptions{
		WorkerPoolSize: 100,
		MaxBatchSize:   1000,
		HLSFolder:      hlsFolder,
		Encryption: &scramble.AES256Encryption{
			Key: []byte(os.Getenv("ENCRYPTION_SECRET")),
		},
	})

	// This encryption needs a public, private key for server and client for complete network encryption
	// So this needs updates to the user or video service in order to retrieve a keypair
	//scrambler := scramble.NewScrambler(&scramble.VideoProcessor{}, h.SQL, &scramble.ScramblerOptions{
	//	WorkerPoolSize: 100,
	//	MaxBatchSize:   1000,
	//	HLSFolder:      hlsFolder,
	//	Encryption: &scramble.PKCS1Encryption{
	//		PublicKeyPath:  os.Getenv(""),
	//		PrivateKeyPath: os.Getenv(""),
	//	},
	//})

	m3u8Path := fmt.Sprintf("%s/playlist.m3u8", hlsFolder)
	contents, err := scrambler.Scramble(m3u8Path, videoId, userId)
	if err != nil {
		h.Logger.Error("NewStream error", err.Error())
		return nil, err
	}

	fmt.Println(contents)

	return &pb.NewStreamResponse{
		Type:     pb.StreamType_HLS,
		M3U8File: contents,
	}, nil
}

func (h *StreamHandler) StreamSegment(req *pb.StreamSegmentRequest, stream pb.StreamingService_StreamSegmentServer) error {
	ctx := stream.Context()
	video, err := h.Video.GetById(ctx, &pbVideo.GetByIdRequest{Id: req.GetVideoId()})
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	th := token.NewJWTTokenHandler()

	tokenValid, data := th.Validate(req.GetToken())
	if !tokenValid {
		return errors.New("invalid token")
	}

	jwtDataToken, ok := data["token"].(string)
	if !ok {
		h.Logger.Error("invalid jwt data")
		return errors.New("invalid token")
	}

	jwtDataSessionId, ok := data["session_id"].(string)
	if !ok {
		h.Logger.Error("invalid jwt data")
		return errors.New("invalid token")
	}

	videoSession, err := models.GetVideoSessionById(h.SQL, jwtDataSessionId)

	if videoSession.Token != jwtDataToken {
		fmt.Println("jwt data token is not the same as in db")
		return errors.New("invalid token")
	}

	if videoSession.UserId != req.GetUserId() {
		fmt.Println("jwt data user id is not the same as in request")
		return errors.New("invalid token user")
	}

	segmentPath := fmt.Sprintf("%s/hls/%s", video.Path, videoSession.FragmentPath)

	file, err := os.Open(segmentPath)
	if err != nil {
		return err
	}
	defer file.Close()

	buffer := make([]byte, 1024*1024) // 1MB buffer
	for {
		n, err := file.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		err = stream.Send(&pb.StreamSegmentResponse{
			Content: buffer[:n],
			Size:    int64(n),
		})
		if err != nil {
			return err
		}
	}

	deleteCtx, deleteCancel := context.WithTimeout(ctx, 3*time.Second)
	defer deleteCancel()

	if _, err := h.SQL.ExecContext(deleteCtx, "DELETE FROM video_sesssions WHERE id = ?", videoSession.Id); err != nil {
		fmt.Printf("cannot delete video_session [%s]: %s", videoSession.Id, err.Error())
	}

	return nil
}
