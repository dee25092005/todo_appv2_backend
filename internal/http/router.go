package http

import (
	"net/http"

	"todo-backend/internal/task"
	"todo-backend/internal/user"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func NewRouter(
	e *echo.Echo,
	db *pgxpool.Pool,
	userHandler *user.Handler,
	taskHandler *task.Handler,
	jwtScecret string,

) {
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

	userGroup := e.Group("/api/v1/users")
	userGroup.GET("/:id", userHandler.GetByID)

	protected := e.Group("/api/v1")
	protected.Use(JWTMiddleware(jwtScecret))

	protected.GET("/users/me", userHandler.GetMe)

	//tasks
	protected.POST("/tasks", taskHandler.Create)
	protected.GET("/tasks", taskHandler.List)
	protected.GET("/tasks/:id", taskHandler.GetByID)
	protected.PUT("/tasks/:id", taskHandler.Update)
	protected.DELETE("/tasks/:id", taskHandler.Delete)

}
