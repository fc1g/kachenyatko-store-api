package app

import (
	"github.com/fc1g/kachenyatko-store-api/internal/config"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
)

type App struct {
	Validator *validator.Validate
	Logger    zerolog.Logger
	Config    *config.Config
}

func New(validator *validator.Validate, cfg *config.Config, logger zerolog.Logger) *App {
	return &App{
		Validator: validator,
		Logger:    logger,
		Config:    cfg,
	}
}

func (a *App) Run() {
	a.Logger.Info().Msg("Hello, World!")
}
