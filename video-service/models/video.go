package models

import (
	"encoding/json"
	"github.com/google/uuid"
	"time"
)

type Video struct {
	Id          string     `json:"id"`
	UserId      string     `json:"user_id"`
	Status      string     `json:"status"`
	Path        string     `json:"path"`
	Size        int64      `json:"size"`
	MimeType    string     `json:"mime_type"`
	Metadata    string     `json:"metadata"`
	ProcessedAt *time.Time `json:"processed_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type Metadata struct {
	Resolution string `json:"resolution"`
	Duration   int    `json:"duration"`
	Format     string `json:"format"`
	Codec      string `json:"codec"`
	Bitrate    int    `json:"bitrate"`
}

func NewVideo(video *Video, metadata *Metadata) (*Video, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return nil, err
	}

	video.Id = id.String()
	video.Metadata = string(metadataBytes)
	return video, nil
}

func ScanToVideo(res Scanner) (*Video, *Metadata, error) {
	var (
		video        Video
		metadata     Metadata
		jsonMetadata string
	)

	err := res.Scan(
		&video.Id,
		&video.UserId,
		&video.Status,
		&video.Path,
		&video.Size,
		&video.MimeType,
		&jsonMetadata,
		&video.ProcessedAt,
		&video.CreatedAt,
		&video.UpdatedAt,
	)
	if err != nil {
		return nil, nil, err
	}

	err = json.Unmarshal([]byte(jsonMetadata), &metadata)
	if err != nil {
		return nil, nil, err
	}

	return &video, &metadata, nil
}
