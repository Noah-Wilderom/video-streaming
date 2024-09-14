package models

import (
	"time"
)

type User struct {
	Id        string    `json:"id" sql:"primary_key"`
	Name      string    `json:"name"`
	Email     string    `json:"email" sql:"unique_key"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func ScanToUser(res Scanner) (*User, error) {
	var user User

	err := res.Scan(
		&user.Id,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}
