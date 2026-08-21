package user

import (
	"context"
	"errors"
	"fmt"
	"time"
	"todo-backend/internal/domain"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	Register(ctx context.Context, req RegisterRequest) (*UserResponse, error)
	Login(ctx context.Context, req LoginRequest) (*LoginResponse, error)
}

type userService struct {
	repo      Repository
	jwtSecret []byte
}

func NewService(repo Repository, jwtSecret string) Service {
	return &userService{
		repo:      repo,
		jwtSecret: []byte(jwtSecret),
	}
}

func (s *userService) Register(ctx context.Context, req RegisterRequest) (*UserResponse, error) {
	existingUser, err := s.repo.FindByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, domain.ErrAlreadyExists
	}
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	newUser := &domain.User{
		Email:        req.Email,
		PasswordHash: string(hashedBytes),
		DispalyName:  req.DisplayName,
	}

	if err := s.repo.Create(ctx, newUser); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &UserResponse{
		ID:          newUser.ID,
		Email:       newUser.Email,
		DisplayName: newUser.DispalyName,
		AvatarURL:   newUser.AvatarURL,
		CreatedAt:   newUser.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

func (s *userService) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	//fetch user by email
	u, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrUnauthorized
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &JWTCustomClaims{
		UserID: u.ID,
		Email:  u.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)

	if err != nil {
		return nil, fmt.Errorf("failed to create token: %w", err)
	}

	return &LoginResponse{
		Token:     tokenString,
		ExpiresAt: expirationTime,
		User: UserResponse{
			ID:          u.ID,
			Email:       u.Email,
			DisplayName: u.DispalyName,
			AvatarURL:   u.AvatarURL,
			CreatedAt:   u.CreatedAt.Format("2006-01-02 15:04:05"),
		},
	}, nil
}
