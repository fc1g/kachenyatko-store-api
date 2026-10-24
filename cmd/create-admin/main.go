package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"time"

	"github.com/fc1g/kachenyatko-store-api/internal/database"
	"github.com/fc1g/kachenyatko-store-api/internal/user"
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

	pool, err := database.Connect(ctx, dbURL, dbPingTimeout)
	if err != nil {
		return err
	}
	defer pool.Close()

	userService := user.NewService(database.NewStore(pool))

	user, err := userService.Create(ctx, user.CreateUserInput{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return err
	}

	log.Info().Str("email", user.Email).Msg("admin created")
	return nil
}
