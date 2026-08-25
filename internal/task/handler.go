package task

import (
	"errors"
	"net/http"
	"todo-backend/internal/domain"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) Create(c echo.Context) error {
	userID := c.Get(domain.UserIDKey).(string)
	var req CreateTaskRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	res, err := h.service.CreateTask(c.Request().Context(), userID, req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, res)
}

func (h *Handler) List(c echo.Context) error {
	userID := c.Get(domain.UserIDKey).(string)
	var query PagingationQuery
	if err := c.Bind(&query); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	res, err := h.service.ListTasks(c.Request().Context(), userID, query)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, res)
}

func (h *Handler) GetByID(c echo.Context) error {
	userID := c.Get(domain.UserIDKey).(string)
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "id is required"})
	}
	res, err := h.service.GetTaskByID(c.Request().Context(), id, userID)
	if err != nil {
		if errors.Is(err, domain.ErrTaskNotFound) {
			return c.JSON(http.StatusNotFound, echo.Map{"error": "task not found"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "internal server error"})
	}
	return c.JSON(http.StatusOK, res)
}

func (h *Handler) Update(c echo.Context) error {
	userID := c.Get(domain.UserIDKey).(string)
	id := c.Param("id")
	var req UpdateTaskRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}
	if err := c.Validate(&req); err != nil {
		return err
	}
	res, err := h.service.UpdateTask(c.Request().Context(), id, userID, req)
	if err != nil {
		if errors.Is(err, domain.ErrTaskNotFound) {
			return c.JSON(http.StatusNotFound, echo.Map{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, res)
}

func (h *Handler) Delete(c echo.Context) error {
	userID := c.Get(domain.UserIDKey).(string)
	id := c.Param("id")
	if err := h.service.DeleteTask(c.Request().Context(), id, userID); err != nil {
		if errors.Is(err, domain.ErrTaskNotFound) {
			return c.JSON(http.StatusNotFound, echo.Map{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}
