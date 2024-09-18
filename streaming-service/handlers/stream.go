package handlers

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/Noah-Wilderom/video-streaming/shared/crypt"
	"github.com/Noah-Wilderom/video-streaming/shared/token"
	"github.com/Noah-Wilderom/video-streaming/streaming-service/models"
	pb "github.com/Noah-Wilderom/video-streaming/streaming-service/proto/stream"
	pbVideo "github.com/Noah-Wilderom/video-streaming/streaming-service/proto/video"
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

	m3u8Path := fmt.Sprintf("%s/hls/playlist.m3u8", video.Path)
	contents, err := h.modifyM3U8ForAPI(fmt.Sprintf("%s/hls", video.Path), m3u8Path, videoId, userId)
	if err != nil {
		h.Logger.Error("NewStream error", err.Error())
		return nil, err
	}

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

	return nil
}

func (h *StreamHandler) modifyM3U8ForAPI(hlsFolder string, m3u8File string, videoId string, userId string) ([]byte, error) {
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
			contents = append(contents, '\n')
			continue
		}

		uniqueToken, err := crypt.GenerateSecretKey()
		if err != nil {
			fmt.Println("GenerateSecretKey error", err)
			return nil, err
		}

		encryptedToken, err := crypt.Encrypt([]byte(uniqueToken))
		if err != nil {
			fmt.Println("Encrypting error", err)
			return nil, err
		}
		base64Token := base64.StdEncoding.EncodeToString(encryptedToken)

		videoSessionId, err := uuid.NewV7()
		if err != nil {
			fmt.Println("NewV7 error", err)
			return nil, err
		}

		fragmentContents, err := os.ReadFile(fmt.Sprintf("%s/%s", hlsFolder, strings.TrimSpace(line)))
		if err != nil {
			fmt.Println("ReadFile error", err)
			return nil, err
		}

		hashHandler := sha256.New()
		_, err = hashHandler.Write(fragmentContents)
		if err != nil {
			fmt.Println("WriteHash error", err)
			return nil, err
		}

		videoSession := &models.VideoSession{
			Id:           videoSessionId.String(),
			UserId:       userId,
			VideoId:      videoId,
			FragmentHash: fmt.Sprintf("%x", hashHandler.Sum(nil)),
			FragmentPath: strings.TrimSpace(line),
			Token:        base64Token,
		}

		th := token.NewJWTTokenHandler()
		jwtToken, err := th.New(map[string]string{
			"token":      base64Token,
			"session_id": videoSession.Id,
		})
		if err != nil {
			fmt.Println("NewJWTToken error", err)
			return nil, err
		}

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
			fmt.Println("insert into db error", err)
			return nil, err
		}

		newLine := fmt.Sprintf("http://localhost:8080/stream/segment/%s?token=%s\n", videoId, jwtToken)
		contents = append(contents, []byte(newLine)...)
	}

	return contents, nil
}
