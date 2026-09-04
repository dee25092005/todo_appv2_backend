package user

import (
	"errors"
	"net/http"
	"todo-backend/internal/domain"
	"todo-backend/pkg/apperrors"

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
		return apperrors.BadRequest("invalid request payload")
	}

	if err := c.Validate(&req); err != nil {
		return apperrors.BadRequest(err.Error())
	}

	res, err := h.service.Register(c.Request().Context(), req)
	if err != nil {
		if errors.Is(err, domain.ErrAlreadyExists) {
			return apperrors.Conflict("user already exists")
		}
		return apperrors.Internal(err)
	}
	return c.JSON(http.StatusCreated, res)
}

func (h *Handler) Login(c echo.Context) error {
	var req LoginRequest

	if err := c.Bind(&req); err != nil {
		return apperrors.BadRequest("invalid request payload")
	}

	if err := c.Validate(&req); err != nil {
		return apperrors.BadRequest(err.Error())
	}

	userAgent := c.Request().UserAgent()
	clientIP := c.RealIP()

	res, err := h.service.Login(c.Request().Context(), req, userAgent, clientIP)
	if err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			return apperrors.Unauthorized("invalid email or password")
		}
		return apperrors.Internal(err)
	}

	return c.JSON(http.StatusOK, res)
}

func (h *Handler) GetByID(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return apperrors.BadRequest("id is required")
	}
	res, err := h.service.GetByID(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return apperrors.NotFound("user not found")
		}
		return apperrors.Internal(err)
	}
	return c.JSON(http.StatusOK, res)
}

func (h *Handler) GetMe(c echo.Context) error {
	userID, ok := c.Get(domain.UserIDKey).(string)
	if !ok || userID == "" {
		return apperrors.Unauthorized("unauthorized")
	}

	res, err := h.service.GetByID(c.Request().Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return apperrors.NotFound("user not found")
		}
		return apperrors.Internal(err)
	}

	return c.JSON(http.StatusOK, res)

}

func (h *Handler) RefreshToken(c echo.Context) error {
	var req RefreshTokenRequest

	if err := c.Bind(&req); err != nil {
		return apperrors.BadRequest("invalid request payload")
	}

	if err := c.Validate(&req); err != nil {
		return apperrors.BadRequest(err.Error())
	}

	userAgent := c.Request().UserAgent()
	clientIP := c.RealIP()

	res, err := h.service.RefreshToken(c.Request().Context(), req.RefreshToken, userAgent, clientIP)
	if err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			return apperrors.Unauthorized("invalid refresh token")
		}
		return apperrors.Internal(err)
	}
	return c.JSON(http.StatusOK, res)
}

func (h *Handler) Logout(c echo.Context) error {
	var req RefreshTokenRequest

	if err := c.Bind(&req); err != nil {
		return apperrors.BadRequest("invalid request payload")
	}

	if err := c.Validate(&req); err != nil {
		return apperrors.BadRequest(err.Error())
	}
	if err := h.service.Logout(c.Request().Context(), req.RefreshToken); err != nil {
		return apperrors.Internal(err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) UpdateProfile(c echo.Context) error {
	userID, ok := c.Get(domain.UserIDKey).(string)
	if !ok || userID == "" {
		return apperrors.Unauthorized("unauthorized")
	}
	var req UpdateProfileReqest
	if err := c.Bind(&req); err != nil {
		return apperrors.BadRequest(err.Error())
	}
	if err := c.Validate(&req); err != nil {
		return err
	}
	res, err := h.service.UpdateProfile(c.Request().Context(), userID, req)
	if err != nil {
		return apperrors.Internal(err)
	}
	return c.JSON(http.StatusOK, res)
}

func (h *Handler) UpdatePassword(c echo.Context) error {
	userID, ok := c.Get(domain.UserIDKey).(string)
	if !ok || userID == "" {
		return apperrors.Unauthorized("unauthorized")
	}
	var req UpdatePasswordRequest
	if err := c.Bind(&req); err != nil {
		return apperrors.BadRequest(err.Error())
	}
	if err := c.Validate(&req); err != nil {
		return err
	}
	if err := h.service.UpdatePassword(c.Request().Context(), userID, req.OldPassword, req.NewPassword); err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			return apperrors.Unauthorized("old password incorrect")
		}
		return apperrors.Internal(err)
	}
	return c.NoContent(http.StatusNoContent)
}
