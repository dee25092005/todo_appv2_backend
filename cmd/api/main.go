package main

import (
	"log"
	"todo-backend/config"
	"todo-backend/internal/database"
	"todo-backend/internal/http"
	"todo-backend/internal/task"
	"todo-backend/internal/user"

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

	//repo
	userRepo := user.NewRepository(dbPool)
	taskRepo := task.NewRepository(dbPool)

	//service
	userService := user.NewService(userRepo, cfg.JWTSecret)
	taskService := task.NewService(taskRepo)

	//handler
	userHandler := user.NewHandler(userService)
	taskHandler := task.NewHandler(taskService)

	e := echo.New()

	http.NewRouter(
		e,
		dbPool,
		userHandler,
		taskHandler,
		cfg.JWTSecret,
	)

	log.Printf("Server is running on port %s", cfg.Port)

	if err := e.Start(":" + cfg.Port); err != nil {
		log.Fatal("Failed to start server: ", err)
	}

}
