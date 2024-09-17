package models

import "time"

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
