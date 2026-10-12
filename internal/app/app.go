package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/fc1g/kachenyatko-store-api/internal/config"
	"github.com/fc1g/kachenyatko-store-api/internal/handler"
	"github.com/fc1g/kachenyatko-store-api/internal/repository"
	"github.com/fc1g/kachenyatko-store-api/internal/router"
	"github.com/fc1g/kachenyatko-store-api/internal/service"
	"github.com/fc1g/kachenyatko-store-api/internal/validation"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

type App struct {
	Validator validation.Validator
	Logger    zerolog.Logger
	Config    *config.Config
	DB        *pgxpool.Pool
	Server    *http.Server
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

	categoryService := service.NewCategoryService(queries, v)
	toyService := service.NewToyService(queries, v)

	categoryHandler := handler.NewCategoryHandler(logger, categoryService)
	toyHandler := handler.NewToyHandler(logger, toyService)
	r := router.New(categoryHandler, toyHandler)

	server := &http.Server{Addr: ":" + cfg.Port, Handler: r}

	go server.ListenAndServe()

	logger.Info().Str("addr", ":"+cfg.Port).Msg("Server is listening")

	return &App{
		Validator: v,
		Logger:    logger,
		Config:    cfg,
		DB:        pool,
		Server:    server,
	}, nil
}

func (a *App) Close() {
	a.DB.Close()

	a.Logger.Info().Msg("Application stopped gracefully")
}
