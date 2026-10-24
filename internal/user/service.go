package user

import (
	"context"
	"strings"
	"time"

	"github.com/fc1g/kachenyatko-store-api/internal/apperr"
	"github.com/fc1g/kachenyatko-store-api/internal/auth"
	"github.com/fc1g/kachenyatko-store-api/internal/database/sqlc"
	"github.com/fc1g/kachenyatko-store-api/internal/validation"
	"github.com/google/uuid"
)

type repository interface {
	CreateUser(ctx context.Context, arg sqlc.CreateUserParams) (sqlc.CreateUserRow, error)
	ListUsers(ctx context.Context, arg sqlc.ListUsersParams) ([]sqlc.ListUsersRow, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (sqlc.GetUserByIDRow, error)
	UpdateUserEmail(ctx context.Context, arg sqlc.UpdateUserEmailParams) (sqlc.UpdateUserEmailRow, error)
	DeleteUser(ctx context.Context, id uuid.UUID) (int64, error)
}

type Service struct {
	repo repository
}

func NewService(r repository) *Service {
	return &Service{
		repo: r,
	}
}

type User struct {
	ID        uuid.UUID
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// createUserInput holds the email and password for creating a user.
type CreateUserInput struct {
	// case-insensitive
	Email    string `validate:"required,email,max=254"`
	Password string `validate:"required,min=8,max=128"`
} // @name CreateUserInput

func (s *Service) Create(ctx context.Context, input CreateUserInput) (User, error) {
	email := strings.TrimSpace(input.Email)
	password := input.Password

	if err := validation.Struct(CreateUserInput{Email: email, Password: password}); err != nil {
		return User{}, err
	}

	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		return User{}, err
	}

	user, err := s.repo.CreateUser(ctx, sqlc.CreateUserParams{
		Email:          email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		return User{}, userError("create user", err)
	}

	return User{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

func (s *Service) List(ctx context.Context, limit, offset int32) ([]User, error) {
	listUsersRow, err := s.repo.ListUsers(ctx, sqlc.ListUsersParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, userError("list users", err)
	}

	users := make([]User, 0, len(listUsersRow))
	for _, user := range listUsersRow {
		users = append(users, User{
			ID:        user.ID,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		})
	}

	return users, nil
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (User, error) {
	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return User{}, userError("get user by id", err)
	}

	return User{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

// updateUserInput holds the new email for a user.
type updateUserInput struct {
	// New email; must be unique, case-insensitive
	Email string `json:"email" validate:"required,email,max=254" format:"email" maxLength:"254" example:"user@example.com"`
} // @name UpdateUserInput

func (s *Service) Update(ctx context.Context, id uuid.UUID, input updateUserInput) (User, error) {
	email := strings.TrimSpace(input.Email)
	if err := validation.Struct(updateUserInput{Email: email}); err != nil {
		return User{}, err
	}

	user, err := s.repo.UpdateUserEmail(ctx, sqlc.UpdateUserEmailParams{
		ID:    id,
		Email: email,
	})
	if err != nil {
		return User{}, userError("update user email", err)
	}

	return User{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	count, err := s.repo.DeleteUser(ctx, id)
	if err != nil {
		return userError("delete user", err)
	}

	if count == 0 {
		return apperr.NotFound("user not found", nil)
	}

	return nil
}
