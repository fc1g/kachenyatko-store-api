package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/fc1g/kachenyatko-store-api/internal/app"
	"github.com/fc1g/kachenyatko-store-api/internal/config"
	"github.com/fc1g/kachenyatko-store-api/internal/logging"
	"github.com/fc1g/kachenyatko-store-api/internal/validation"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	v := validation.New()
	cfg, err := config.Load(v)
	if err != nil {
		log.Fatal().Err(err).Msg("Couldn't load config")
	}

	logger, err := logging.New(cfg.Logging)
	if err != nil {
		log.Fatal().Err(err).Msg("Couldn't initialize logger")
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	a, err := app.New(ctx, v, cfg, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Couldn't initialize application")
	}

	<-ctx.Done()
	stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.Server.Shutdown(shutdownCtx); err != nil {
		logger.Error().Err(err).Msg("Server forced to shutdown")
	}

	a.Close()
}
