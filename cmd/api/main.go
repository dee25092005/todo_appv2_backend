package main

import (
	"log"
	"todo-backend/config"
	"todo-backend/internal/database"
	"todo-backend/internal/http"
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

	//service
	userService := user.NewService(userRepo, cfg.JWTSecret)

	//handler
	userHandler := user.NewHandler(userService)

	e := echo.New()

	http.NewRouter(
		e,
		dbPool,
		userHandler,
	)

	log.Printf("Server is running on port %s", cfg.Port)

	if err := e.Start(":" + cfg.Port); err != nil {
		log.Fatal("Failed to start server: ", err)
	}

}
