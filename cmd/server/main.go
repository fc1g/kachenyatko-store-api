package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/fc1g/kachenyatko-store-api/internal/app"
	"github.com/fc1g/kachenyatko-store-api/internal/config"
	"github.com/fc1g/kachenyatko-store-api/internal/logging"
	"github.com/fc1g/kachenyatko-store-api/internal/validation"
	"github.com/rs/zerolog/log"
)

func main() {
	if err := run(); err != nil {
		log.Error().Err(err).Msg("application failed")
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	v := validation.New()
	cfg, err := config.Load(v)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger, err := logging.New(cfg.Logging)
	if err != nil {
		return fmt.Errorf("initialize logger: %w", err)
	}
	log.Logger = logger

	a, err := app.New(ctx, cfg, logger, v)
	if err != nil {
		return fmt.Errorf("initialize application: %w", err)
	}
	defer a.Close()

	return a.Run(ctx, stop)
}
