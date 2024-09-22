package scramble

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/Noah-Wilderom/video-streaming/shared/crypt"
	"github.com/Noah-Wilderom/video-streaming/shared/token"
	"github.com/Noah-Wilderom/video-streaming/streaming-service/models"
	"github.com/google/uuid"
	"os"
	"strings"
)

type Processor interface {
	ProcessSegment(videoId string, userId string, segmentPath string, hlsFolder string) (*models.VideoSession, string, error)
}

type VideoProcessor struct{}

// ProcessSegment implements the Processor interface
func (vp *VideoProcessor) ProcessSegment(videoId string, userId string, segmentPath string, hlsFolder string) (*models.VideoSession, string, error) {
	uniqueToken, err := crypt.GenerateSecretKey()
	if err != nil {
		return nil, "", fmt.Errorf("GenerateSecretKey error: %v", err)
	}

	encryptedToken, err := crypt.Encrypt([]byte(uniqueToken))
	if err != nil {
		return nil, "", fmt.Errorf("Encrypt error: %v", err)
	}
	base64Token := base64.StdEncoding.EncodeToString(encryptedToken)

	videoSessionId, err := uuid.NewV7()
	if err != nil {
		return nil, "", fmt.Errorf("UUID error: %v", err)
	}

	fragmentContents, err := os.ReadFile(fmt.Sprintf("%s/%s", hlsFolder, strings.TrimSpace(segmentPath)))
	if err != nil {
		return nil, "", fmt.Errorf("ReadFile error: %v", err)
	}

	hashHandler := sha256.New()
	_, err = hashHandler.Write(fragmentContents)
	if err != nil {
		return nil, "", fmt.Errorf("Hash error: %v", err)
	}

	videoSession := &models.VideoSession{
		Id:           videoSessionId.String(),
		UserId:       userId,
		VideoId:      videoId,
		FragmentHash: fmt.Sprintf("%x", hashHandler.Sum(nil)),
		FragmentPath: strings.TrimSpace(segmentPath),
		Token:        base64Token,
	}

	th := token.NewJWTTokenHandler()
	jwtToken, err := th.New(map[string]string{
		"token":      base64Token,
		"session_id": videoSession.Id,
	})
	if err != nil {
		return nil, "", fmt.Errorf("JWT error: %v", err)
	}

	newLine := fmt.Sprintf("http://localhost:8080/stream/segment/%s?token=%s\n", videoId, jwtToken)
	fmt.Println("Segment processed:", newLine)
	return videoSession, newLine, nil
}
