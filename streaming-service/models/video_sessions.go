package models

import (
	"context"
	"gofr.dev/pkg/gofr/container"
	"time"
)

type VideoSession struct {
	Id           string    `json:"id"`
	UserId       string    `json:"user_id"`
	VideoId      string    `json:"video_id"`
	FragmentHash string    `json:"fragment_hash"`
	FragmentPath string    `json:"fragment_path"`
	Token        string    `json:"token"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func ScanToVideoSession(res Scanner) (*VideoSession, error) {
	var videoSession VideoSession

	err := res.Scan(
		&videoSession.Id,
		&videoSession.UserId,
		&videoSession.VideoId,
		&videoSession.FragmentHash,
		&videoSession.FragmentPath,
		&videoSession.Token,
		&videoSession.CreatedAt,
		&videoSession.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &videoSession, nil
}

func GetVideoSessionById(sql container.DB, id string) (*VideoSession, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	row := sql.QueryRowContext(ctx, "SELECT * FROM video_sessions WHERE id = ?", id)
	if err := row.Err(); err != nil {
		return nil, err
	}

	videoSession, err := ScanToVideoSession(row)
	if err != nil {
		return nil, err
	}

	return videoSession, nil
}
