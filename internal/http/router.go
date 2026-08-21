package http

import (
	"net/http"
	"todo-backend/internal/user"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func NewRouter(e *echo.Echo, db *pgxpool.Pool, userHandler *user.Handler) {
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${time_rfc3339} | method=${method} | uri=${uri} | status=${status} | status=${status} \n",
	}))

	e.Use(middleware.Recover())
	e.Validator = NewCustomValidator()

	e.GET("/health", func(c echo.Context) error {
		if err := db.Ping(c.Request().Context()); err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"status": "unhealthy",
				"db":     "disconnected",
			})
		}
		return c.JSON(http.StatusOK, echo.Map{
			"status": "healthy",
			"db":     "connected",
		})
	})

	authGroup := e.Group("/api/v1/auth")
	authGroup.POST("/register", userHandler.Register)
	authGroup.POST("/login", userHandler.Login)
}
