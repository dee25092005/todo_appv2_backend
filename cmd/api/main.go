package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"todo-backend/config"
	"todo-backend/internal/database"
	https "todo-backend/internal/http"
	"todo-backend/internal/session"
	"todo-backend/internal/task"
	"todo-backend/internal/user"
	"todo-backend/internal/worker"
	"todo-backend/pkg/storage"

	"github.com/labstack/echo/v4"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config: ", err)
	}

	dbPool, err := database.NewPostgresPool(cfg.DatabaseUrl)
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}

	defer dbPool.Close()

	//storage
	storage, err := storage.NewR2Storage(context.Background(), *cfg)
	if err != nil {
		log.Fatal("Failed to connect to storage: ", err)
	}
	//repo
	userRepo := user.NewRepository(dbPool)
	taskRepo := task.NewRepository(dbPool)
	sessionRepo := session.NewSessionRepository(dbPool)

	//service
	userService := user.NewService(userRepo, sessionRepo, dbPool, cfg.JWTSecret)
	taskService := task.NewService(taskRepo, storage)

	//handler
	userHandler := user.NewHandler(userService)
	taskHandler := task.NewHandler(taskService)

	e := echo.New()

	https.NewRouter(
		e,
		dbPool,
		userHandler,
		taskHandler,
		cfg.JWTSecret,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cleaner := worker.NewSessionCleaner(sessionRepo, 12*time.Hour)
	go cleaner.Start(ctx)

	go func() {
		slog.Info("starting http server", slog.String("port", cfg.Port))
		if err := e.Start(":" + cfg.Port); err != nil && err != http.ErrServerClosed {
			slog.Error("server shutdown ungracefully", slog.Any("error", err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down application gracefully...")

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := e.Shutdown(shutdownCtx); err != nil {
		slog.Error("failed to shutdown HTTP server", slog.Any("error", err))
	}

	slog.Info("application stopped cleanly")

}
