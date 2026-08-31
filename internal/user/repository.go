package user

import (
	"context"
	"errors"
	"fmt"
	"todo-backend/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Create(ctx context.Context, u *domain.User) error
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByID(ctx context.Context, id string) (*domain.User, error)
	UpdateProfile(ctx context.Context, id, displayName, avatarURL string) error
	UpdatePassword(ctx context.Context, id string, passwordHash string) error
}

type postgresRepository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &postgresRepository{
		db: db,
	}
}

func (r *postgresRepository) Create(ctx context.Context, u *domain.User) error {
	query := `
		INSERT INTO users (email,password_hash,display_name,avatar_url,avatar_key)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(
		ctx,
		query,
		u.Email,
		u.PasswordHash,
		u.DispalyName,
		u.AvatarURL,
		u.AvatarKey,
	).Scan(
		&u.ID, &u.CreatedAt, &u.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *postgresRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT 
			id,email,
			password_hash,
			display_name,
			COALESCE(avatar_url,''),
			COALESCE(avatar_key,''),
			created_at,
			updated_at
		FROM users
		WHERE email = $1
	`

	var u domain.User

	err := r.db.QueryRow(
		ctx,
		query,
		email,
	).Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.DispalyName,
		&u.AvatarURL,
		&u.AvatarKey,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	return &u, nil
}

func (r *postgresRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT 
			id,
			email,
			password_hash,
			display_name,
			COALESCE(avatar_url,''),
			COALESCE(avatar_key,''),
			created_at,
			updated_at
		FROM users
		WHERE id = $1
	`

	var u domain.User
	err := r.db.QueryRow(
		ctx, query, id,
	).Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.DispalyName,
		&u.AvatarURL,
		&u.AvatarKey,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return &u, nil
}

func (r *postgresRepository) UpdateProfile(ctx context.Context, id, displayName, avatarURL string) error {
	query := `
		UPDATE users
		SET display_name = $1, avatar_url = $2, updated_at = NOW()
		WHERE id = $3
	`
	_, err := r.db.Exec(
		ctx,
		query,
		displayName,
		avatarURL,
		id,
	)

	return err
}

func (r *postgresRepository) UpdatePassword(ctx context.Context, id string, passwordHash string) error {
	query := `
		UPDATE users
		SET password_hash = $1, updated_at = NOW()
		WHERE id = $2
	`
	_, err := r.db.Exec(
		ctx,
		query,
		passwordHash,
		id,
	)

	return err
}
