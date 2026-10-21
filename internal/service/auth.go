package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/fc1g/kachenyatko-store-api/internal/apperr"
	"github.com/fc1g/kachenyatko-store-api/internal/auth"
	"github.com/fc1g/kachenyatko-store-api/internal/repository"
	"github.com/fc1g/kachenyatko-store-api/internal/validation"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type authRepository interface {
	CreateUser(ctx context.Context, arg repository.CreateUserParams) (repository.CreateUserRow, error)
	GetUserByEmail(ctx context.Context, lower string) (repository.User, error)
}

type tokenIssuer interface {
	Issue(userID uuid.UUID) (token string, expiresAt time.Time, err error)
}

type AuthService struct {
	repo   authRepository
	tokens tokenIssuer
	v      validation.Validator
}

func NewAuthService(r authRepository, tokens tokenIssuer, v validation.Validator) *AuthService {
	return &AuthService{
		repo:   r,
		tokens: tokens,
		v:      v,
	}
}

type AuthResult struct {
	User        UserRow
	AccessToken string
	ExpiresAt   time.Time
}

// LoginInput holds the credentials for logging in.
type LoginInput struct {
	// Admin email, case-insensitive
	Email string `json:"email" validate:"required,max=254" format:"email" maxLength:"254" example:"user@example.com"`
	// Admin password
	Password string `json:"password" validate:"required,max=128" format:"password" maxLength:"128" example:"S3cureP@ssw0rd"`
} // @name LoginInput

var errInvalidCredentials = apperr.Error{Code: apperr.CodeUnauthorized, Message: "invalid credentials"}

func (s *AuthService) Login(ctx context.Context, input LoginInput) (AuthResult, error) {
	input.Email = strings.TrimSpace(input.Email)

	if err := validate(s.v, input); err != nil {
		return AuthResult{}, err
	}

	user, err := s.repo.GetUserByEmail(ctx, input.Email)
	if errors.Is(err, pgx.ErrNoRows) {
		auth.CompareDummyPassword(input.Password)
		return AuthResult{}, errInvalidCredentials
	}
	if err != nil {
		return AuthResult{}, fmt.Errorf("get user by email: %w", err)
	}

	match, err := auth.ComparePassword(input.Password, user.HashedPassword)
	if err != nil {
		return AuthResult{}, err
	}
	if !match {
		return AuthResult{}, errInvalidCredentials
	}

	accessToken, expiresAt, err := s.tokens.Issue(user.ID)
	if err != nil {
		return AuthResult{}, err
	}

	return AuthResult{
		User:        UserRow{ID: user.ID, Email: user.Email, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt},
		AccessToken: accessToken,
		ExpiresAt:   expiresAt,
	}, nil
}

type CreateAdminInput struct {
	Email    string `validate:"required,email,max=254"`
	Password string `validate:"required,min=8,max=128"`
}

func (s *AuthService) CreateAdmin(ctx context.Context, input CreateAdminInput) (UserRow, error) {
	input.Email = strings.TrimSpace(input.Email)

	if err := validate(s.v, input); err != nil {
		return UserRow{}, err
	}

	hashedPassword, err := auth.HashPassword(input.Password)
	if err != nil {
		return UserRow{}, err
	}

	user, err := s.repo.CreateUser(ctx, repository.CreateUserParams{
		Email:          input.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		return UserRow{}, userError(err, "create admin")
	}

	return UserRow{ID: user.ID, Email: user.Email, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt}, nil
}
