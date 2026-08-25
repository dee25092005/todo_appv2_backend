package domain

import (
	"errors"
	"time"
)

var (
	ErrTaskNotFound = errors.New("task not found")
	ErrForbidden    = errors.New("forbidden access")
)

type Task struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	Title       string     `json:"title"`
	Note        *string    `json:"note"`
	IsAllDay    bool       `json:"is_all_day"`
	DueDate     *time.Time `json:"due_date"`
	IsCompleted bool       `json:"is_completed"`
	IsDeleted   bool       `json:"is_deleted"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
