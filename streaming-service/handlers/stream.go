package handlers

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/Noah-Wilderom/video-streaming/shared/crypt"
	"github.com/Noah-Wilderom/video-streaming/shared/token"
	"github.com/Noah-Wilderom/video-streaming/streaming-service/models"
	pb "github.com/Noah-Wilderom/video-streaming/streaming-service/proto/stream"
	"github.com/google/uuid"
	"gofr.dev/pkg/gofr/container"
	"io"
	"os"
	"strings"
	"time"
)

type StreamHandler struct {
	*container.Container
	pb.UnimplementedStreamingServiceServer
}

func (h *StreamHandler) NewStream(ctx context.Context, req *pb.NewStreamRequest) (*pb.NewStreamResponse, error) {
	m3u8Path := fmt.Sprintf("/output/%s/hls/playlist.m3u8", req.VideoId)
	contents, err := h.modifyM3U8ForAPI(m3u8Path, req.VideoId, req.UserId)
	if err != nil {
		h.Logger.Error("NewStream error", err)
		return nil, err
	}

	return &pb.NewStreamResponse{
		Type:     pb.StreamType_HLS,
		M3U8File: contents,
	}, nil
}

func (h *StreamHandler) StreamSegment(req *pb.StreamSegmentRequest, stream pb.StreamingService_StreamSegmentServer) error {
	segmentPath := fmt.Sprintf("/output/%s/hls/%s", req.VideoId)

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

	return nil
}

func (h *StreamHandler) modifyM3U8ForAPI(m3u8File string, videoId string, userId string) ([]byte, error) {
	var contents []byte

	file, err := os.Open(m3u8File)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if ok := strings.HasPrefix(line, "playlist"); !ok {
			contents = append(contents, scanner.Bytes()...)
			continue
		}
		fmt.Println(line)
		// line = playlistx.ts

		uniqueToken, err := crypt.GenerateSecretKey()
		if err != nil {
			return nil, err
		}

		encryptedToken, err := crypt.Encrypt([]byte(uniqueToken))
		if err != nil {
			return nil, err
		}
		base64Token := base64.StdEncoding.EncodeToString(encryptedToken)

		videoSessionId, err := uuid.NewV7()
		if err != nil {
			return nil, err
		}

		fragmentContents, err := os.ReadFile(fmt.Sprintf("/output/%s/%s", videoId, strings.TrimSpace(line)))
		if err != nil {
			return nil, err
		}

		hashHandler := sha256.New()
		_, err = hashHandler.Write(fragmentContents)
		if err != nil {
			return nil, err
		}

		videoSession := &models.VideoSession{
			Id:           videoSessionId.String(),
			UserId:       userId,
			VideoId:      videoId,
			FragmentHash: string(hashHandler.Sum(nil)),
			FragmentPath: strings.TrimSpace(line),
			Token:        base64Token,
		}

		th := token.NewJWTTokenHandler()
		jwtToken, err := th.New(map[string]string{
			"token":      base64Token,
			"session_id": videoSession.Id,
		})

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		_, err = h.SQL.ExecContext(
			ctx,
			"INSERT INTO video_sessions (id, user_id, video_id, fragment_hash, fragment_path, token) VALUES (?, ?, ?, ?, ?, ?)",
			videoSession.Id,
			videoSession.UserId,
			videoSession.VideoId,
			videoSession.FragmentHash,
			videoSession.FragmentPath,
			videoSession.Token,
		)
		if err != nil {
			return nil, err
		}

		newLine := fmt.Sprintf("http://localhost:8080/video/streaming/%s?token=%s\n", videoId, jwtToken)
		contents = append(contents, []byte(newLine)...)
	}

	return contents, nil
}
