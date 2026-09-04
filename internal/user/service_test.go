package user

import (
	"context"
	"errors"
	"testing"
	"todo-backend/internal/domain"
	"todo-backend/internal/session"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type MockUserRepository struct {
	mock.Mock
}

type MockSessionRepository struct {
	mock.Mock
}

type MockDB struct {
	BeginFn func(ctx context.Context) (pgx.Tx, error)
}

var _ Repository = (*MockUserRepository)(nil)

func (m *MockDB) Begin(ctx context.Context) (pgx.Tx, error) {
	if m.BeginFn != nil {
		return m.BeginFn(ctx)
	}
	return nil, errors.New("Begin not implemented in this test")
}

func (m *MockUserRepository) FindByEmail(
	ctx context.Context,
	email string,
) (*domain.User, error) {
	args := m.Called(ctx, email)

	if args.Get(0) != nil {
		return args.Get(0).(*domain.User), args.Error(1)
	}

	return nil, args.Error(1)
}

func (m *MockUserRepository) Create(
	ctx context.Context,
	u *domain.User,
) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *MockUserRepository) FindByID(
	ctx context.Context,
	id string,
) (*domain.User, error) {
	args := m.Called(ctx, id)

	if args.Get(0) != nil {
		return args.Get(0).(*domain.User), args.Error(1)
	}

	return nil, args.Error(1)
}

func (m *MockUserRepository) UpdateProfile(
	ctx context.Context,
	id string,
	displayName string,
	avatarURL string,
) error {
	args := m.Called(ctx, id, displayName, avatarURL)
	return args.Error(0)
}

func (m *MockUserRepository) UpdatePassword(
	ctx context.Context,
	id string,
	passwordHash string,
) error {
	args := m.Called(ctx, id, passwordHash)
	return args.Error(0)
}

func (m *MockSessionRepository) Create(
	ctx context.Context,
	params *domain.CreateSessionParams,
) (*domain.Session, error) {
	args := m.Called(ctx, params)

	if args.Get(0) != nil {
		return args.Get(0).(*domain.Session), args.Error(1)
	}

	return nil, args.Error(1)
}

func (m *MockSessionRepository) GetByHash(
	ctx context.Context,
	hash string,
) (*domain.Session, error) {
	args := m.Called(ctx, hash)

	if args.Get(0) != nil {
		return args.Get(0).(*domain.Session), args.Error(1)
	}

	return nil, args.Error(1)
}

func (m *MockSessionRepository) Revoke(
	ctx context.Context,
	id uuid.UUID,
) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSessionRepository) RevokeAllUserID(
	ctx context.Context,
	userID uuid.UUID,
) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockSessionRepository) DeleteExpired(
	ctx context.Context,
) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}
func (m *MockSessionRepository) WithTx(tx pgx.Tx) session.SessionRepository {
	return m
}

func TestRegister_DuplicateEmail(t *testing.T) {

	mockRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)
	mockDB := new(MockDB)

	svc := NewService(mockRepo, mockSessionRepo, mockDB, "test-secret")
	req := RegisterRequest{
		Email:       "duplicate@test.com",
		Password:    "password123",
		DisplayName: "Dee",
	}

	mockRepo.On("FindByEmail", mock.Anything, "duplicate@test.com").
		Return(&domain.User{Email: "duplicate@test.com"}, nil)

	res, err := svc.Register(context.Background(), req)

	require.ErrorIs(t, err, domain.ErrAlreadyExists)
	require.Nil(t, res)
	mockRepo.AssertExpectations(t)
}

func TestWrongPassword(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)
	mockDB := new(MockDB)

	svc := NewService(mockRepo, mockSessionRepo, mockDB, "test-secret")
	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte("correct-password"), bcrypt.DefaultCost,
	)
	require.NoError(t, err)

	user := &domain.User{
		Email:        "test@test.com",
		PasswordHash: string(hashedPassword),
	}

	mockRepo.On(
		"FindByEmail",
		mock.Anything,
		"test@test.com",
	).Return(user, nil)

	req := LoginRequest{
		Email:    "test@test.com",
		Password: "wrong-password",
	}

	res, err := svc.Login(context.Background(), req, "", "")
	require.ErrorIs(t, err, domain.ErrUnauthorized)
	require.Nil(t, res)
	mockRepo.AssertExpectations(t)

}
