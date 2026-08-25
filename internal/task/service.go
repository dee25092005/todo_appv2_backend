package task

import (
	"context"
	"todo-backend/internal/domain"
)

type Service interface {
	CreateTask(ctx context.Context, userID string, req CreateTaskRequest) (*TaskResponse, error)
	GetTaskByID(ctx context.Context, id, userID string) (*TaskResponse, error)
	ListTasks(ctx context.Context, userID string, query PagingationQuery) (*PaginatedTaskResponse, error)
	UpdateTask(ctx context.Context, id, userID string, req UpdateTaskRequest) (*TaskResponse, error)
	DeleteTask(ctx context.Context, id, userID string) error
}

type TaskService struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &TaskService{repo: repo}
}
func toTaskResponse(t *domain.Task) *TaskResponse {
	if t == nil {
		return nil
	}
	return &TaskResponse{
		ID:          t.ID,
		Title:       t.Title,
		Note:        t.Note,
		IsAllDay:    t.IsAllDay,
		DueDate:     t.DueDate,
		IsCompleted: t.IsCompleted,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

func (s *TaskService) CreateTask(ctx context.Context, userID string, req CreateTaskRequest) (*TaskResponse, error) {
	t := &domain.Task{
		UserID:   userID,
		Title:    req.Title,
		Note:     req.Note,
		IsAllDay: req.IsAllDay,
		DueDate:  req.DueDate,
	}
	if err := s.repo.Create(ctx, t); err != nil {
		return nil, err
	}
	return toTaskResponse(t), nil
}

func (s *TaskService) GetTaskByID(ctx context.Context, id, userID string) (*TaskResponse, error) {
	t, err := s.repo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	return toTaskResponse(t), nil
}

func (s *TaskService) ListTasks(ctx context.Context, userID string, q PagingationQuery) (*PaginatedTaskResponse, error) {
	q.SetDefault()

	tasks, total, err := s.repo.ListByUserID(ctx, userID, q.Page, q.Limit)
	if err != nil {
		return nil, err
	}
	res := make([]*TaskResponse, 0, len(tasks))
	for _, t := range tasks {
		res = append(res, toTaskResponse(t))
	}
	totalPages := (total + q.Limit - 1) / q.Limit
	return &PaginatedTaskResponse{
		Data:       res,
		TotalItems: total,
		Page:       q.Page,
		Limit:      q.Limit,
		TotalPages: totalPages,
	}, nil
}

func (s *TaskService) UpdateTask(ctx context.Context, id, userID string, req UpdateTaskRequest) (*TaskResponse, error) {
	t, err := s.repo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if req.Title != nil {
		t.Title = *req.Title
	}
	if req.Note != nil {
		t.Note = req.Note
	}
	if req.IsAllDay != nil {
		t.IsAllDay = *req.IsAllDay
	}
	if req.DueDate != nil {
		t.DueDate = req.DueDate
	}
	if req.IsCompleted != nil {
		t.IsCompleted = *req.IsCompleted
	}
	if err := s.repo.Update(ctx, t); err != nil {
		return nil, err
	}
	return toTaskResponse(t), nil

}

func (s *TaskService) DeleteTask(ctx context.Context, id, userID string) error {
	return s.repo.SoftDelete(ctx, id, userID)
}
