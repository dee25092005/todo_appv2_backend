package http

import (
	"errors"
	"strings"
	"todo-backend/internal/user"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

const UserIDKey = "user_id"

func JWTMiddleware(jwtSecret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.ErrUnauthorized
			}
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return echo.ErrUnauthorized
			}

			tokenStrings := parts[1]

			claims := &user.JWTCustomClaims{}
			token, err := jwt.ParseWithClaims(tokenStrings, claims, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, errors.New("unexpected signing method")
				}
				return []byte(jwtSecret), nil
			})

			if err != nil || !token.Valid {
				return echo.ErrUnauthorized
			}

			c.Set(UserIDKey, claims.UserID)

			return next(c)

		}
	}
}
