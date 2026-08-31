package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrTaskNotFound = errors.New("task not found")
	ErrForbidden    = errors.New("forbidden access")
	ErrFileNotFound = errors.New("file not found")
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

type TaskFile struct {
	ID        uuid.UUID `json:"id"`
	TaskID    uuid.UUID `json:"task_id"`
	FileURL   string    `json:"file_url"`
	FileKey   string    `json:"file_key"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
