package app

import (
	"context"
	"fmt"

	"github.com/fc1g/kachenyatko-store-api/internal/config"
	"github.com/fc1g/kachenyatko-store-api/internal/repository"
	"github.com/fc1g/kachenyatko-store-api/internal/service"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

type App struct {
	Validator *validator.Validate
	Logger    zerolog.Logger
	Config    *config.Config
	DB        *pgxpool.Pool

	ToyService      *service.ToyService
	CategoryService *service.CategoriesService
}

func New(ctx context.Context, v *validator.Validate, cfg *config.Config, logger zerolog.Logger) (*App, error) {
	pool, err := pgxpool.New(ctx, cfg.Database.URL)
	if err != nil {
		return nil, fmt.Errorf("create db pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	queries := repository.New(pool)

	categoryService := service.NewCategoriesService(queries)
	toyService := service.NewToyService(queries)

	return &App{
		Validator: v,
		Logger:    logger,
		Config:    cfg,
		DB:        pool,

		CategoryService: categoryService,
		ToyService:      toyService,
	}, nil
}

func (a *App) Close() {
	a.DB.Close()

	a.Logger.Info().Msg("Application stopped gracefully")
}
