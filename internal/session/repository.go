package session

import (
	"context"
	"errors"
	"fmt"
	"todo-backend/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SessionRepository interface {
	Create(ctx context.Context, params *domain.CreateSessionParams) (*domain.Session, error)
	GetByHash(ctx context.Context, hash string) (*domain.Session, error)
	Revoke(ctx context.Context, id uuid.UUID) error
	RevokeAllUserID(ctx context.Context, userID uuid.UUID) error
	DeleteExpired(ctx context.Context) (int64, error)
}

type sessionRepository struct {
	db *pgxpool.Pool
}

func NewSessionRepository(db *pgxpool.Pool) SessionRepository {
	return &sessionRepository{
		db: db,
	}
}

func (r *sessionRepository) Create(ctx context.Context, params *domain.CreateSessionParams) (*domain.Session, error) {
	query := `
		INSERT INTO sessions (user_id, refresh_token_hash, user_agent,client_ip, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, refresh_token_hash, user_agent, client_ip, is_revoked, expires_at, created_at
	`

	var s domain.Session
	err := r.db.QueryRow(ctx, query,
		params.UserID,
		params.RefreshTokenHash,
		params.UserAgent,
		params.ClientIP,
		params.ExpiresAt,
	).Scan(
		&s.ID,
		&s.UserID,
		&s.RefreshTokenHash,
		&s.UserAgent,
		&s.ClientIP,
		&s.IsRevoked,
		&s.ExpiresAt,
		&s.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	return &s, nil
}

func (r *sessionRepository) GetByHash(ctx context.Context, hash string) (*domain.Session, error) {
	query := `
		SELECT id, user_id, refresh_token_hash, user_agent, client_ip, is_revoked, expires_at, created_at
		FROM sessions
		WHERE refresh_token_hash = $1
	`
	var s domain.Session
	err := r.db.QueryRow(ctx, query, hash).Scan(
		&s.ID,
		&s.UserID,
		&s.RefreshTokenHash,
		&s.UserAgent,
		&s.ClientIP,
		&s.IsRevoked,
		&s.ExpiresAt,
		&s.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	return &s, nil

}

func (r *sessionRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE sessions
		SET is_revoked = true
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}
	return nil
}

func (r *sessionRepository) RevokeAllUserID(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE sessions
		SET is_revoked = true
		WHERE user_id = $1
	`
	_, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}
	return nil
}

func (r *sessionRepository) DeleteExpired(ctx context.Context) (int64, error) {
	query := `
		DELETE FROM sessions
		WHERE expires_at < NOW()

	`
	res, err := r.db.Exec(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired sessions: %w", err)
	}
	return res.RowsAffected(), nil
}
