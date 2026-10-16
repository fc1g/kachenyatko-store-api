package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/fc1g/kachenyatko-store-api/internal/config"
	"github.com/fc1g/kachenyatko-store-api/internal/handler"
	"github.com/fc1g/kachenyatko-store-api/internal/repository"
	"github.com/fc1g/kachenyatko-store-api/internal/router"
	"github.com/fc1g/kachenyatko-store-api/internal/service"
	"github.com/fc1g/kachenyatko-store-api/internal/validation"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

type App struct {
	logger          zerolog.Logger
	db              *pgxpool.Pool
	server          *http.Server
	shutdownTimeout time.Duration
}

func New(ctx context.Context, cfg *config.Config, logger zerolog.Logger, v validation.Validator) (*App, error) {
	pool, err := pgxpool.New(ctx, cfg.Database.URL)
	if err != nil {
		return nil, fmt.Errorf("create db pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, time.Second*2)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	repo := repository.NewStore(pool)

	categoryService := service.NewCategoryService(repo, v)
	toyService := service.NewToyService(repo, v)

	healthHandler := handler.NewHealthHandler(pool)
	categoryHandler := handler.NewCategoryHandler(categoryService)
	toyHandler := handler.NewToyHandler(toyService)
	r := router.New(logger, router.Handlers{
		Health:   healthHandler,
		Category: categoryHandler,
		Toy:      toyHandler,
	})

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.HTTP.Port),
		Handler:           r,
		ReadHeaderTimeout: cfg.HTTP.ReadHeaderTimeout,
		ReadTimeout:       cfg.HTTP.ReadTimeout,
		WriteTimeout:      cfg.HTTP.WriteTimeout,
		IdleTimeout:       cfg.HTTP.IdleTimeout,
	}

	return &App{
		logger:          logger,
		db:              pool,
		server:          server,
		shutdownTimeout: cfg.HTTP.ShutdownTimeout,
	}, nil
}

func (a *App) Run(ctx context.Context, stop context.CancelFunc) error {
	errCh := make(chan error, 1)

	a.logger.Info().Str("addr", a.server.Addr).Msg("http server started")
	go func() {
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		stop()
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), a.shutdownTimeout)
	defer cancel()

	if err := a.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown server: %w", err)
	}

	a.logger.Info().Msg("server stopped")

	return nil
}

func (a *App) Close() {
	a.db.Close()
}
