package domain

import "time"

const UserIDKey = "user_id"

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	DispalyName  string    `json:"display_name"`
	AvatarURL    string    `json:"avatar_url"`
	AvatarKey    string    `json:"avatar_key"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
