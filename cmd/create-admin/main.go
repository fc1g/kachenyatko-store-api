package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"time"

	"github.com/fc1g/kachenyatko-store-api/internal/repository"
	"github.com/fc1g/kachenyatko-store-api/internal/service"
	"github.com/fc1g/kachenyatko-store-api/internal/validation"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

const (
	envAdminEmail    = "ADMIN_EMAIL"
	envAdminPassword = "ADMIN_PASSWORD"
	envDBURL         = "DB_URL"

	dbPingTimeout = 2 * time.Second
)

func main() {
	if err := run(); err != nil {
		log.Error().Err(err).Msg("create admin failed")
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()

	if err := godotenv.Load(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("load .env: %w", err)
	}

	email := strings.TrimSpace(os.Getenv(envAdminEmail))
	if email == "" {
		return fmt.Errorf("%s is required", envAdminEmail)
	}

	password := os.Getenv(envAdminPassword)
	if password == "" {
		return fmt.Errorf("%s is required", envAdminPassword)
	}

	dbURL := os.Getenv(envDBURL)
	if dbURL == "" {
		return fmt.Errorf("%s is required", envDBURL)
	}

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return fmt.Errorf("create db pool: %w", err)
	}
	defer pool.Close()

	pingCtx, cancel := context.WithTimeout(ctx, dbPingTimeout)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		return fmt.Errorf("ping db: %w", err)
	}

	authService := service.NewAuthService(repository.NewStore(pool), nil, validation.New())

	user, err := authService.CreateAdmin(ctx, service.CreateAdminInput{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return err
	}

	log.Info().Str("email", user.Email).Msg("admin created")
	return nil
}
