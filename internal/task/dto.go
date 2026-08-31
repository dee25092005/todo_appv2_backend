package task

import "time"

type PagingationQuery struct {
	Page  int `query:"page" validate:"omitempty,min=1"`
	Limit int `query:"limit" validate:"omitempty,min=1,max=100"`
}

func (p *PagingationQuery) SetDefault() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.Limit < 1 {
		p.Limit = 10
	}
}

type UpdateTaskStatusRequest struct {
	IsCompleted bool `json:"is_completed"`
}

type CreateTaskRequest struct {
	Title    string     `json:"title" validate:"required,min=1,max=255"`
	Note     *string    `json:"note"`
	IsAllDay bool       `json:"is_all_day"`
	DueDate  *time.Time `json:"due_date"`
}

type UpdateTaskRequest struct {
	Title       *string    `json:"title" validate:"min=1,max=255"`
	Note        *string    `json:"note"`
	IsAllDay    *bool      `json:"is_all_day"`
	DueDate     *time.Time `json:"due_date"`
	IsCompleted *bool      `json:"is_completed"`
}

type TaskResponse struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Note        *string    `json:"note"`
	IsAllDay    bool       `json:"is_all_day"`
	DueDate     *time.Time `json:"due_date"`
	IsCompleted bool       `json:"is_completed"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type PaginatedTaskResponse struct {
	Data       []*TaskResponse `json:"data"`
	TotalItems int             `json:"total_items"`
	Page       int             `json:"page"`
	Limit      int             `json:"limit"`
	TotalPages int             `json:"total_pages"`
}

type SearchTaskQuery struct {
	Q     string `query:"q" validate:"required,min=1"`
	Page  int    `query:"page" validate:"omitempty,min=1"`
	Limit int    `query:"limit" validate:"omitempty,min=1,max=100"`
}

func (p *SearchTaskQuery) SetDefault() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.Limit < 1 {
		p.Limit = 10
	}
}
