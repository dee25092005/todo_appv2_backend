package task

import (
	"errors"
	"fmt"
	"net/http"
	"todo-backend/internal/domain"
	"todo-backend/pkg/apperrors"

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
		return apperrors.BadRequest(err.Error())
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	res, err := h.service.CreateTask(c.Request().Context(), userID, req)
	if err != nil {
		return apperrors.Internal(err)
	}
	return c.JSON(http.StatusCreated, res)
}

func (h *Handler) List(c echo.Context) error {
	userID := c.Get(domain.UserIDKey).(string)
	var query PagingationQuery
	if err := c.Bind(&query); err != nil {
		return apperrors.BadRequest(err.Error())
	}

	res, err := h.service.ListTasks(c.Request().Context(), userID, query)
	if err != nil {
		return apperrors.Internal(err)
	}

	return c.JSON(http.StatusOK, res)
}

func (h *Handler) Search(c echo.Context) error {
	userID := c.Get(domain.UserIDKey).(string)
	var query SearchTaskQuery
	if err := c.Bind(&query); err != nil {
		return apperrors.BadRequest("invalid query")
	}
	if query.Q == "" {
		return apperrors.BadRequest("q is required")
	}

	res, err := h.service.SearchTasks(c.Request().Context(), userID, query)
	if err != nil {
		return apperrors.Internal(err)
	}
	return c.JSON(http.StatusOK, res)
}

func (h *Handler) GetByID(c echo.Context) error {
	userID := c.Get(domain.UserIDKey).(string)
	id := c.Param("id")
	if id == "" {
		return apperrors.BadRequest("id is required")
	}
	res, err := h.service.GetTaskByID(c.Request().Context(), id, userID)
	if err != nil {
		if errors.Is(err, domain.ErrTaskNotFound) {
			return apperrors.NotFound("task not found")
		}
		return apperrors.Internal(err)
	}
	return c.JSON(http.StatusOK, res)
}

func (h *Handler) Update(c echo.Context) error {
	userID := c.Get(domain.UserIDKey).(string)
	id := c.Param("id")
	var req UpdateTaskRequest
	if err := c.Bind(&req); err != nil {
		return apperrors.BadRequest("invalid request")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}
	res, err := h.service.UpdateTask(c.Request().Context(), id, userID, req)
	if err != nil {
		if errors.Is(err, domain.ErrTaskNotFound) {
			return apperrors.NotFound("task not found")
		}
		return apperrors.Internal(err)
	}
	return c.JSON(http.StatusOK, res)
}

func (h *Handler) Delete(c echo.Context) error {
	userID := c.Get(domain.UserIDKey).(string)
	id := c.Param("id")
	if err := h.service.DeleteTask(c.Request().Context(), id, userID); err != nil {
		if errors.Is(err, domain.ErrTaskNotFound) {
			return apperrors.NotFound("task not found")
		}
		return apperrors.Internal(err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) UploadFile(c echo.Context) error {
	fmt.Println("start upload file handler")

	userID := c.Get(domain.UserIDKey).(string)
	taskID := c.Param("id")

	fmt.Println("start upload file")

	file, err := c.FormFile("file")
	if err != nil {
		return apperrors.BadRequest("file is required")
	}

	fmt.Println("upload file")

	src, err := file.Open()
	if err != nil {
		return apperrors.Internal(err)
	}
	defer src.Close()

	fmt.Println("close file")

	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	fmt.Println("get content type")

	res, err := h.service.UploadTaskFile(c.Request().Context(), taskID, userID, file.Filename, src, contentType)
	if err != nil {
		if errors.Is(err, domain.ErrTaskNotFound) {
			return apperrors.NotFound("task not found")
		}
		return apperrors.Internal(err)
	}
	return c.JSON(http.StatusCreated, res)

}

func (h *Handler) DeleteFile(c echo.Context) error {
	userID := c.Get(domain.UserIDKey).(string)
	taskID := c.Param("id")
	fileID := c.Param("file_id")

	err := h.service.DeleteTaskFile(c.Request().Context(), taskID, fileID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrFileNotFound) {
			return apperrors.NotFound("file not found")
		}
		return apperrors.Internal(err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) ListFiles(c echo.Context) error {
	userID := c.Get(domain.UserIDKey).(string)
	taskID := c.Param("id")
	res, err := h.service.ListTaskFiles(c.Request().Context(), taskID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrTaskNotFound) {
			return apperrors.NotFound("task not found")
		}
		return apperrors.Internal(err)
	}
	return c.JSON(http.StatusOK, res)
}

func (h *Handler) UpdateStatus(c echo.Context) error {
	userID := c.Get(domain.UserIDKey).(string)
	taskID := c.Param("id")
	var req UpdateTaskStatusRequest
	if err := c.Bind(&req); err != nil {
		return apperrors.BadRequest("invalid request")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}
	err := h.service.UpdateTaskStatus(c.Request().Context(), taskID, userID, req.IsCompleted)
	if err != nil {
		if errors.Is(err, domain.ErrTaskNotFound) {
			return apperrors.NotFound("task not found")
		}
		return apperrors.Internal(err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) HardDeleteTask(c echo.Context) error {
	userID := c.Get(domain.UserIDKey).(string)
	taskID := c.Param("id")
	err := h.service.HardDeleteTask(c.Request().Context(), taskID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrTaskNotFound) {
			return apperrors.NotFound("task not found")
		}
		return apperrors.Internal(err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) ListTrash(c echo.Context) error {
	userID := c.Get(domain.UserIDKey).(string)
	var query PagingationQuery
	if err := c.Bind(&query); err != nil {
		return apperrors.BadRequest(err.Error())
	}
	res, err := h.service.ListTrash(c.Request().Context(), userID, query)
	if err != nil {
		return apperrors.Internal(err)
	}
	return c.JSON(http.StatusOK, res)

}

func (h *Handler) RestoreTask(c echo.Context) error {
	userID := c.Get(domain.UserIDKey).(string)
	taskID := c.Param("id")

	if err := h.service.RestoreTask(c.Request().Context(), taskID, userID); err != nil {
		if errors.Is(err, domain.ErrTaskNotFound) {
			return apperrors.NotFound("task not found")
		}
		return apperrors.Internal(err)
	}
	return c.NoContent(http.StatusNoContent)
}
