package user

import (
	"context"
	"errors"
	"fmt"
	"time"
	"todo-backend/internal/domain"
	"todo-backend/internal/session"
	"todo-backend/pkg/utils"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	Register(ctx context.Context, req RegisterRequest) (*UserResponse, error)
	Login(ctx context.Context, req LoginRequest, userAgent string, clientIP string) (*LoginResponse, error)
	GetByID(ctx context.Context, id string) (*UserResponse, error)
	RefreshToken(ctx context.Context, refreshToken string, userAgent string, clientIP string) (*LoginResponse, error)
	Logout(ctx context.Context, refreshToken string) error
	UpdatePassword(ctx context.Context, userId, oldpw, newpw string) error
	UpdateProfile(ctx context.Context, userID string, req UpdateProfileReqest) (*UserResponse, error)
}

type DB interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

type userService struct {
	repo      Repository
	session   session.SessionRepository
	db        DB
	jwtSecret []byte
}

func NewService(repo Repository, session session.SessionRepository, db DB, jwtSecret string) Service {
	return &userService{
		repo:      repo,
		session:   session,
		db:        db,
		jwtSecret: []byte(jwtSecret),
	}
}

func (s *userService) generateTokenPairAndSession(ctx context.Context, userID uuid.UUID, u *domain.User, userAgent, clientIP string) (*LoginResponse, error) {
	accessExpiration := time.Now().Add(15 * time.Minute)
	claims := &JWTCustomClaims{
		UserID: u.ID,
		Email:  u.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExpiration),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to create access token: %w", err)
	}

	rawRefreshToken, err := utils.GenerateRandomToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	refreshExpiration := time.Now().Add(7 * 24 * time.Hour)
	tokenHash := utils.HashToken(rawRefreshToken)

	_, err = s.session.Create(ctx, &domain.CreateSessionParams{
		UserID:           userID,
		RefreshTokenHash: tokenHash,
		UserAgent:        userAgent,
		ClientIP:         clientIP,
		ExpiresAt:        refreshExpiration,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to persist session: %w", err)
	}

	return &LoginResponse{
		AcessToken:   accessToken,
		RefreshToken: rawRefreshToken,
		ExpiresAt:    accessExpiration,
		User: UserResponse{
			ID:          u.ID,
			Email:       u.Email,
			DisplayName: u.DispalyName,
			AvatarURL:   u.AvatarURL,
			CreatedAt:   u.CreatedAt.Format("2006-01-02 15:04:05"),
		},
	}, nil
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

func (s *userService) Login(ctx context.Context, req LoginRequest, userAgent string, clientIP string) (*LoginResponse, error) {

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

	userUUID, err := uuid.Parse(u.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user id: %w", err)
	}

	if err := s.session.RevokeAllUserID(ctx, userUUID); err != nil {
		return nil, fmt.Errorf("failed to revoke sessions: %w", err)
	}

	return s.generateTokenPairAndSession(ctx, userUUID, u, userAgent, clientIP)
}

func (s *userService) GetByID(ctx context.Context, id string) (*UserResponse, error) {
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	return &UserResponse{
		ID:          u.ID,
		Email:       u.Email,
		DisplayName: u.DispalyName,
		AvatarURL:   u.AvatarURL,
		CreatedAt:   u.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

func (s *userService) RefreshToken(ctx context.Context, refreshToken string, userAgent string, clientIP string) (*LoginResponse, error) {
	tokenHash := utils.HashToken(refreshToken)
	sess, err := s.session.GetByHash(ctx, tokenHash)

	if err != nil || sess == nil {

		return nil, domain.ErrUnauthorized
	}
	if sess.IsRevoked {
		_ = s.session.RevokeAllUserID(ctx, sess.UserID)
		return nil, domain.ErrUnauthorized
	}

	if time.Now().After(sess.ExpiresAt) {
		return nil, domain.ErrUnauthorized
	}
	if err := s.session.Revoke(ctx, sess.ID); err != nil {
		return nil, fmt.Errorf("failed to revoke session: %w", err)
	}

	u, err := s.repo.FindByID(ctx, sess.UserID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	return s.generateTokenPairAndSession(ctx, sess.UserID, u, userAgent, clientIP)
}

func (s *userService) Logout(ctx context.Context, refreshToken string) error {
	tokenHash := utils.HashToken(refreshToken)
	sess, err := s.session.GetByHash(ctx, tokenHash)
	if err != nil || sess == nil {
		return nil
	}
	return s.session.Revoke(ctx, sess.ID)
}

func (s *userService) UpdatePassword(ctx context.Context, userId, oldpw, newpw string) error {
	u, err := s.repo.FindByID(ctx, userId)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}
	err = bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(oldpw))
	if err != nil {
		return domain.ErrUnauthorized
	}
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(newpw), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	u.PasswordHash = string(hashedBytes)
	if err := s.repo.UpdatePassword(ctx, userId, u.PasswordHash); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	return s.session.RevokeAllUserID(ctx, uuid.MustParse(userId))
}

func (s *userService) UpdateProfile(ctx context.Context, userID string, req UpdateProfileReqest) (*UserResponse, error) {
	if err := s.repo.UpdateProfile(ctx, userID, req.DisplayName, req.AvatarURL); err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}
	u, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	return &UserResponse{
		ID:          u.ID,
		Email:       u.Email,
		DisplayName: u.DispalyName,
		AvatarURL:   u.AvatarURL,
		CreatedAt:   u.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}
