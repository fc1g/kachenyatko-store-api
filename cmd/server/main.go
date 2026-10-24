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
	"github.com/rs/zerolog/log"
)

// @title Kachenyatko Store API
// @version 1.0
// @description API for a small handmade toy shop
// @BasePath /
func main() {
	if err := run(); err != nil {
		log.Error().Err(err).Msg("application failed")
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger, err := logging.Setup(cfg.Logging.Level, cfg.Logging.Pretty)
	if err != nil {
		return fmt.Errorf("initialize logger: %w", err)
	}
	log.Logger = logger

	a, err := app.New(ctx, cfg, logger)
	if err != nil {
		return fmt.Errorf("initialize application: %w", err)
	}
	defer a.Close()

	return a.Run(ctx, stop)
}
