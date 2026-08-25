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
