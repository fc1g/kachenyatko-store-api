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
	"github.com/fc1g/kachenyatko-store-api/internal/storage"
	"github.com/fc1g/kachenyatko-store-api/internal/telegram"
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
	pool, err := newDB(ctx, cfg.Database)
	if err != nil {
		return nil, err
	}

	r2, err := newStorage(ctx, cfg.Storage)
	if err != nil {
		pool.Close()
		return nil, err
	}

	bot, err := telegram.NewBot(cfg.Telegram)
	if err != nil {
		pool.Close()
		return nil, err
	}

	handlers := newHandlers(pool, r2, bot, v, cfg)
	r := router.New(logger, handlers, cfg.Site.URL)

	server := newHTTPServer(cfg.HTTP, r)

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

func newDB(ctx context.Context, cfg config.Database) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("create db pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, cfg.PingTimeout)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	return pool, nil
}

func newStorage(ctx context.Context, cfg config.Storage) (*storage.R2, error) {
	r2, err := storage.NewR2(
		ctx,
		storage.R2Config{
			Endpoint:        cfg.Endpoint,
			AccessKeyID:     cfg.AccessKeyID,
			SecretAccessKey: cfg.SecretAccessKey,
			Bucket:          cfg.BucketName,
			PublicURL:       cfg.PublicURL,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("init r2 storage: %w", err)
	}

	return r2, nil
}

func newHandlers(pool *pgxpool.Pool, r2 *storage.R2, bot *telegram.Bot, v validation.Validator, cfg *config.Config) router.Handlers {
	repo := repository.NewStore(pool)

	categoryService := service.NewCategoryService(repo, v)
	toyService := service.NewToyService(repo, r2, v)
	toyImageService := service.NewToyImageService(repo, r2)
	orderService := service.NewOrderService(repo, bot, v, cfg.Site.URL)

	return router.Handlers{
		Health:   handler.NewHealthHandler(pool, cfg.Database.PingTimeout),
		Category: handler.NewCategoryHandler(categoryService),
		Toy:      handler.NewToyHandler(toyService, r2),
		ToyImage: handler.NewToyImageHandler(toyImageService, r2),
		Order:    handler.NewOrderHandler(orderService),
	}
}

func newHTTPServer(cfg config.HTTP, h http.Handler) *http.Server {
	return &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           h,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}
}
