package user

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
	return &Handler{service: service}
}

func (h *Handler) Register(c echo.Context) error {
	var req RegisterRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	res, err := h.service.Register(c.Request().Context(), req)
	if err != nil {
		if errors.Is(err, domain.ErrAlreadyExists) {
			return c.JSON(http.StatusConflict, echo.Map{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, res)
}

func (h *Handler) Login(c echo.Context) error {
	var req LoginRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	res, err := h.service.Login(c.Request().Context(), req)
	if err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, res)
}
