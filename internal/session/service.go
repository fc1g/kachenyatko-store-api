package session

import (
	"context"
	"strings"
	"time"

	"github.com/fc1g/kachenyatko-store-api/internal/apperr"
	"github.com/fc1g/kachenyatko-store-api/internal/auth"
	"github.com/fc1g/kachenyatko-store-api/internal/database/sqlc"
	"github.com/fc1g/kachenyatko-store-api/internal/user"
	"github.com/fc1g/kachenyatko-store-api/internal/validation"
	"github.com/google/uuid"
)

type repository interface {
	GetUserByEmail(ctx context.Context, lower string) (sqlc.User, error)
}

type tokenIssuer interface {
	Issue(userID uuid.UUID) (token string, expiresAt time.Time, err error)
}

type Service struct {
	repo   repository
	tokens tokenIssuer
}

func NewService(r repository, tokens tokenIssuer) *Service {
	return &Service{
		repo:   r,
		tokens: tokens,
	}
}

type session struct {
	User      user.User
	Token     string
	ExpiresAt time.Time
}

// loginInput holds the credentials for logging in.
type loginInput struct {
	// Email, case-insensitive
	Email    string `json:"email" validate:"required,max=254" format:"email" maxLength:"254" example:"user@example.com"`
	Password string `json:"password" validate:"required,max=128" format:"password" maxLength:"128" example:"S3cureP@ssw0rd"`
} // @name LoginInput

func (s *Service) Login(ctx context.Context, input loginInput) (session, error) {
	email := strings.TrimSpace(input.Email)
	password := input.Password

	if err := validation.Struct(loginInput{Email: email, Password: password}); err != nil {
		return session{}, err
	}

	u, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return session{}, apperr.Unauthorized("invalid credentials", nil)
	}

	match, err := auth.ComparePassword(password, u.HashedPassword)
	if err != nil {
		return session{}, err
	}
	if !match {
		return session{}, apperr.Unauthorized("invalid credentials", nil)
	}

	accessToken, expiresAt, err := s.tokens.Issue(u.ID)
	if err != nil {
		return session{}, err
	}

	return session{
		User: user.User{
			ID:        u.ID,
			Email:     u.Email,
			CreatedAt: u.CreatedAt,
			UpdatedAt: u.UpdatedAt,
		},
		Token:     accessToken,
		ExpiresAt: expiresAt,
	}, nil
}
