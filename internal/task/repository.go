package task

import (
	"context"
	"errors"
	"fmt"
	"todo-backend/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Create(ctx context.Context, task *domain.Task) error
	GetByID(ctx context.Context, id, userID string) (*domain.Task, error)
	ListByUserID(ctx context.Context, userID string, page, limit int) ([]*domain.Task, int, error)
	Update(ctx context.Context, task *domain.Task) error
	SoftDelete(ctx context.Context, id, userID string) error
	HardDelete(ctx context.Context, id, userID string) error
	Search(ctx context.Context, userID, query string, page, limit int) ([]*domain.Task, int, error)

	UpdateStatus(ctx context.Context, id, userID string, isCompleted bool) error
	CreateFile(ctx context.Context, file *domain.TaskFile) error
	GetFileByID(ctx context.Context, fileID, taskID string) (*domain.TaskFile, error)
	ListFileByTaskID(ctx context.Context, taskID string) ([]*domain.TaskFile, error)
	DeleteFile(ctx context.Context, fileID string) error
}

type postgresRepository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &postgresRepository{
		db: db,
	}
}

func (r *postgresRepository) Create(ctx context.Context, t *domain.Task) error {
	quey := `
		INSERT INTO tasks (user_id, title,note,is_all_day, due_date)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING id, is_completed, is_deleted, created_at, updated_at
	`
	res := r.db.QueryRow(
		ctx,
		quey,
		t.UserID,
		t.Title,
		t.Note,
		t.IsAllDay,
		t.DueDate,
	).Scan(
		&t.ID,
		&t.IsCompleted,
		&t.IsDeleted,
		&t.CreatedAt,
		&t.UpdatedAt,
	)

	return res
}

func (r *postgresRepository) GetByID(ctx context.Context, id, userID string) (*domain.Task, error) {
	query := `
		SELECT
			id,
			user_id,
			title,
			note,
			is_all_day,
			due_date,
			is_completed,
			is_deleted,
			created_at,
			updated_at
		FROM tasks
		WHERE id = $1 AND user_id = $2 AND is_deleted = false
	`

	var t domain.Task
	err := r.db.QueryRow(
		ctx,
		query,
		id,
		userID,
	).Scan(
		&t.ID,
		&t.UserID,
		&t.Title,
		&t.Note,
		&t.IsAllDay,
		&t.DueDate,
		&t.IsCompleted,
		&t.IsDeleted,
		&t.CreatedAt,
		&t.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTaskNotFound
		}
		return nil, err
	}
	return &t, nil
}

func (r *postgresRepository) ListByUserID(ctx context.Context, userID string, page, limit int) ([]*domain.Task, int, error) {

	var total int
	countQuery := `
		SELECT COUNT(*) FROM tasks
		WHERE user_id = $1 AND is_deleted = false
	`
	if err := r.db.QueryRow(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * limit

	query := `
		SELECT id, user_id, title,note , is_all_day,due_date, is_completed, is_deleted, created_at, updated_at
		FROM tasks
		WHERE user_id = $1 AND is_deleted = false
		ORDER BY created_at DESC
		LIMIT $2
		OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	tasks := []*domain.Task{}
	for rows.Next() {
		var t domain.Task
		if err := rows.Scan(
			&t.ID,
			&t.UserID,
			&t.Title,
			&t.Note,
			&t.IsAllDay,
			&t.DueDate,
			&t.IsCompleted,
			&t.IsDeleted,
			&t.CreatedAt,
			&t.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		tasks = append(tasks, &t)
	}
	return tasks, total, nil
}

func (r *postgresRepository) Update(ctx context.Context, task *domain.Task) error {
	query := `
		UPDATE tasks
		SET title = $1, note = $2, is_all_day = $3, due_date = $4, is_completed = $5, updated_at = NOW()
		WHERE id = $6 AND user_id = $7 AND is_deleted = false
		RETURNING updated_at
	`
	err := r.db.QueryRow(ctx, query,
		task.Title,
		task.Note,
		task.IsAllDay,
		task.DueDate,
		task.IsCompleted,
		task.ID,
		task.UserID,
	).Scan(&task.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrTaskNotFound
		}
		return fmt.Errorf("failed to update task: %w", err)
	}

	return nil
}

func (r *postgresRepository) SoftDelete(ctx context.Context, id, userID string) error {
	query := `
		UPDATE tasks
		SET is_deleted = true, updated_at = NOW()
		WHERE id = $1 AND user_id = $2 AND is_deleted = false
	`

	cmdTag, err := r.db.Exec(ctx, query, id, userID)

	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return domain.ErrTaskNotFound
	}

	return nil
}

func (r *postgresRepository) Search(ctx context.Context, userID, searchKeyword string, page, limit int) ([]*domain.Task, int, error) {
	searchTerm := "%" + searchKeyword + "%"
	var total int
	countQuery := `
		SELECT COUNT(*)
		FROM tasks
		WHERE user_id = $1
			AND is_deleted = false
			AND (title || '' || COALESCE(note, '')) LIKE $2
	`
	if err := r.db.QueryRow(ctx, countQuery, userID, searchTerm).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * limit

	query := `
		SELECT id, user_id, title,note , is_all_day,due_date, is_completed, is_deleted, created_at, updated_at
		FROM tasks
		WHERE user_id = $1
		AND is_deleted = false
		AND (title || '' || COALESCE(note, '')) LIKE $2
		ORDER BY created_at DESC
		LIMIT $3
		OFFSET $4
	`

	rows, err := r.db.Query(ctx, query, userID, searchTerm, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()

	tasks := []*domain.Task{}
	for rows.Next() {
		var t domain.Task
		if err := rows.Scan(
			&t.ID,
			&t.UserID,
			&t.Title,
			&t.Note,
			&t.IsAllDay,
			&t.DueDate,
			&t.IsCompleted,
			&t.IsDeleted,
			&t.CreatedAt,
			&t.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		tasks = append(tasks, &t)
	}
	return tasks, total, nil

}

func (r *postgresRepository) CreateFile(ctx context.Context, file *domain.TaskFile) error {
	query := `
		INSERT INTO task_files (task_id, file_url,file_key)
		VALUES ($1,$2,$3)
		RETURNING id, created_at
	`
	err := r.db.QueryRow(ctx, query,
		file.TaskID,
		file.FileURL,
		file.FileKey,
	).Scan(
		&file.ID,
		&file.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	return err
}

func (r *postgresRepository) GetFileByID(ctx context.Context, FileID, taskID string) (*domain.TaskFile, error) {
	query := `
		SELECT id,task_id,file_url,file_key,created_at
		FROM task_files
		WHERE id = $1 AND task_id = $2 
	`
	var f domain.TaskFile
	err := r.db.QueryRow(
		ctx,
		query,
		FileID,
		taskID,
	).Scan(
		&f.ID,
		&f.TaskID,
		&f.FileURL,
		&f.FileKey,
		&f.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrFileNotFound
		}
		return nil, fmt.Errorf("failed to get file: %w", err)
	}
	return &f, nil
}

func (r *postgresRepository) ListFileByTaskID(ctx context.Context, taskID string) ([]*domain.TaskFile, error) {
	query := `
		SELECT id, task_id, file_url, file_key, created_at
		FROM task_files
		WHERE task_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var file []*domain.TaskFile
	for rows.Next() {
		var f domain.TaskFile
		if err := rows.Scan(
			&f.ID,
			&f.TaskID,
			&f.FileURL,
			&f.FileKey,
			&f.CreatedAt,
		); err != nil {
			return nil, err
		}
		file = append(file, &f)
	}
	return file, nil
}

func (r *postgresRepository) DeleteFile(ctx context.Context, fileID string) error {
	query := `
		DELETE FROM task_files
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query, fileID)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

func (r *postgresRepository) UpdateStatus(ctx context.Context, id, userID string, isCompleted bool) error {
	query := `
		UPDATE tasks
		SET is_completed = $1, updated_at = NOW()
		WHERE id = $2 AND user_id = $3 AND is_deleted = false
	`

	tag, err := r.db.Exec(ctx, query, isCompleted, id, userID)
	if err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrTaskNotFound
	}
	return nil
}

func (r *postgresRepository) HardDelete(ctx context.Context, id, userID string) error {
	query := `
		DELETE FROM tasks
		WHERE id = $1 AND user_id = $2
	`
	_, err := r.db.Exec(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to hard delete task: %w", err)
	}
	return nil
}
