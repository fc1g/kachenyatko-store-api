package service

import (
	"context"
	"strings"
	"time"

	"github.com/fc1g/kachenyatko-store-api/internal/apperr"
	"github.com/fc1g/kachenyatko-store-api/internal/repository"
	"github.com/fc1g/kachenyatko-store-api/internal/validation"
	"github.com/google/uuid"
)

type userRepository interface {
	GetUsers(ctx context.Context, arg repository.GetUsersParams) ([]repository.GetUsersRow, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (repository.GetUserByIDRow, error)
	UpdateUserEmail(ctx context.Context, arg repository.UpdateUserEmailParams) (repository.UpdateUserEmailRow, error)
	DeleteUser(ctx context.Context, id uuid.UUID) (int64, error)
}

type UserService struct {
	repo userRepository
	v    validation.Validator
}

func NewUserService(r userRepository, v validation.Validator) *UserService {
	return &UserService{
		repo: r,
		v:    v,
	}
}

type UserRow struct {
	ID        uuid.UUID
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (s *UserService) GetAll(ctx context.Context, limit, offset int32) ([]UserRow, error) {
	users, err := s.repo.GetUsers(ctx, repository.GetUsersParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, userError(err, "get users")
	}

	userRows := make([]UserRow, 0, len(users))
	for _, user := range users {
		userRows = append(userRows, UserRow{
			ID:        user.ID,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		})
	}

	return userRows, nil
}

func (s *UserService) GetOne(ctx context.Context, id uuid.UUID) (UserRow, error) {
	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return UserRow{}, userError(err, "get user by id")
	}

	return UserRow{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

type UpdateUserInput struct {
	Email string `json:"email" validate:"required,email,max=254" format:"email" maxLength:"254" example:"user@example.com"`
}

func (s *UserService) Update(ctx context.Context, id uuid.UUID, input UpdateUserInput) (UserRow, error) {
	input.Email = strings.TrimSpace(input.Email)

	if err := validate(s.v, input); err != nil {
		return UserRow{}, err
	}

	user, err := s.repo.UpdateUserEmail(ctx, repository.UpdateUserEmailParams{
		ID:    id,
		Email: input.Email,
	})
	if err != nil {
		return UserRow{}, userError(err, "update user")
	}

	return UserRow{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

func (s *UserService) Delete(ctx context.Context, id uuid.UUID) error {
	count, err := s.repo.DeleteUser(ctx, id)
	if err != nil {
		return userError(err, "delete user")
	}

	if count == 0 {
		return apperr.Error{Code: apperr.CodeNotFound, Message: "user not found"}
	}

	return nil
}
