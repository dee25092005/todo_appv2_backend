package worker

import (
	"context"
	"log/slog"
	"time"
	"todo-backend/internal/session"
)

type SessionCleaner struct {
	sessionRepo session.SessionRepository
	interval    time.Duration
}

func NewSessionCleaner(sessionRepo session.SessionRepository, interval time.Duration) *SessionCleaner {
	return &SessionCleaner{
		sessionRepo: sessionRepo,
		interval:    interval,
	}
}

func (w *SessionCleaner) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	slog.Info("session cleaner started", slog.Duration("interval", w.interval))

	for {
		select {
		case <-ctx.Done():
			slog.Info("session cleaner stopped")
			return
		case <-ticker.C:
			w.clean(ctx)
		}
	}
}

func (w *SessionCleaner) clean(ctx context.Context) {
	cleanCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	rowDeleted, err := w.sessionRepo.DeleteExpired(cleanCtx)
	if err != nil {
		slog.Error(
			"failed to delete expired sessions",
			slog.Int64("deleted_rows", rowDeleted),
			slog.String("error", err.Error()),
		)
		return
	}
	if rowDeleted > 0 {
		slog.Info(
			"deleted expired sessions",
			slog.Int64("deleted_rows", rowDeleted),
		)
	}
}
