package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrSessionNotFound = errors.New("session not found")
)

type Session struct {
	ID               uuid.UUID `json:"id"`
	UserID           uuid.UUID `json:"user_id "`
	RefreshTokenHash string    `json:"-"`
	UserAgent        string    `json:"user_agent"`
	ClientIP         string    `json:"client_ip"`
	IsRevoked        bool      `json:"is_revoked"`
	ExpiresAt        time.Time `json:"expires_at"`
	CreatedAt        time.Time `json:"created_at"`
}

type CreateSessionParams struct {
	UserID           uuid.UUID
	RefreshTokenHash string
	UserAgent        string
	ClientIP         string
	ExpiresAt        time.Time
}
