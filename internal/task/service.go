package task

import (
	"context"
	"fmt"
	"io"
	"todo-backend/internal/domain"
	"todo-backend/pkg/storage"

	"github.com/google/uuid"
)

type Service interface {
	CreateTask(ctx context.Context, userID string, req CreateTaskRequest) (*TaskResponse, error)
	GetTaskByID(ctx context.Context, id, userID string) (*TaskResponse, error)
	SearchTasks(ctx context.Context, userID string, q SearchTaskQuery) (*PaginatedTaskResponse, error)
	ListTasks(ctx context.Context, userID string, query PagingationQuery) (*PaginatedTaskResponse, error)
	UpdateTask(ctx context.Context, id, userID string, req UpdateTaskRequest) (*TaskResponse, error)
	DeleteTask(ctx context.Context, id, userID string) error
	HardDeleteTask(ctx context.Context, id, userID string) error

	UpdateTaskStatus(ctx context.Context, id, userID string, isCompleted bool) error

	UploadTaskFile(ctx context.Context, taskID, userID, fileName string, fileReader io.Reader, contentType string) (*domain.TaskFile, error)
	DeleteTaskFile(ctx context.Context, taskID, fileID, userID string) error
	ListTaskFiles(ctx context.Context, taskID, userID string) ([]*domain.TaskFile, error)
}

type TaskService struct {
	repo    Repository
	storage storage.StorageService
}

func NewService(repo Repository, storage storage.StorageService) Service {
	return &TaskService{
		repo:    repo,
		storage: storage,
	}
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

func (s *TaskService) SearchTasks(ctx context.Context, userID string, q SearchTaskQuery) (*PaginatedTaskResponse, error) {
	q.SetDefault()
	tasks, total, err := s.repo.Search(ctx, userID, q.Q, q.Page, q.Limit)
	if err != nil {
		return nil, err
	}

	response := make([]*TaskResponse, 0, len(tasks))
	for _, t := range tasks {
		response = append(response, toTaskResponse(t))
	}

	totalPages := (total + q.Limit - 1) / q.Limit

	return &PaginatedTaskResponse{
		Data:       response,
		TotalItems: total,
		Page:       q.Page,
		Limit:      q.Limit,
		TotalPages: totalPages,
	}, nil

}

func (s *TaskService) DeleteTask(ctx context.Context, id, userID string) error {
	return s.repo.SoftDelete(ctx, id, userID)
}

func (s *TaskService) UploadTaskFile(ctx context.Context, taskID, userID, fileName string, fileReader io.Reader, contentType string) (*domain.TaskFile, error) {
	task, err := s.repo.GetByID(ctx, taskID, userID)
	if err != nil {
		return nil, err
	}

	fileKey := fmt.Sprintf("tasks/%s/%s_%s", taskID, uuid.New().String(), fileName)

	fileURL, err := s.storage.UploadFile(ctx, fileKey, fileReader, contentType)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}
	taskUUID, _ := uuid.Parse(task.ID)
	taskFile := &domain.TaskFile{
		TaskID:  taskUUID,
		FileURL: fileURL,
		FileKey: fileKey,
	}

	if err := s.repo.CreateFile(ctx, taskFile); err != nil {
		_ = s.storage.DeleteFile(ctx, fileKey)
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	return taskFile, nil
}

func (s *TaskService) DeleteTaskFile(ctx context.Context, taskID, fileID, userID string) error {
	_, err := s.repo.GetByID(ctx, taskID, userID)
	if err != nil {
		return err
	}
	file, err := s.repo.GetFileByID(ctx, fileID, taskID)
	if err != nil {
		return err
	}
	if err := s.storage.DeleteFile(ctx, file.FileKey); err != nil {
		return fmt.Errorf("failed to delete file from storage: %w", err)
	}

	return s.repo.DeleteFile(ctx, fileID)
}

func (s *TaskService) ListTaskFiles(ctx context.Context, taskID, userID string) ([]*domain.TaskFile, error) {
	_, err := s.repo.GetByID(ctx, taskID, userID)
	if err != nil {
		return nil, err
	}
	return s.repo.ListFileByTaskID(ctx, taskID)
}

func (s *TaskService) UpdateTaskStatus(ctx context.Context, id, userID string, isCompleted bool) error {
	t, err := s.repo.GetByID(ctx, id, userID)
	if err != nil {
		return err
	}
	t.IsCompleted = isCompleted
	return s.repo.Update(ctx, t)
}

func (s *TaskService) HardDeleteTask(ctx context.Context, id, userID string) error {
	if id == "" || userID == "" {
		return fmt.Errorf("id or userID is empty")
	}
	return s.repo.HardDelete(ctx, id, userID)
}
