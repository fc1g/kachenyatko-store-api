package main

import (
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

	app := app.New(v, cfg, logger)
	app.Run()
}
