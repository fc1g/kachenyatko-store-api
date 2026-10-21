package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/fc1g/kachenyatko-store-api/internal/auth"
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
	tokens, err := auth.NewTokenManager(cfg.Auth.AccessPrivateKey, cfg.Auth.AccessIssuer, cfg.Auth.AccessTokenTTL)
	if err != nil {
		return nil, err
	}

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

	d := deps{pool: pool, r2: r2, bot: bot, tokens: tokens, v: v, cfg: cfg}
	handlers := newHandlers(d)
	r := router.New(logger, handlers, tokens, cfg.Site.URL)

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

type deps struct {
	pool   *pgxpool.Pool
	r2     *storage.R2
	bot    *telegram.Bot
	tokens *auth.TokenManager
	v      validation.Validator
	cfg    *config.Config
}

func newHandlers(d deps) router.Handlers {
	repo := repository.NewStore(d.pool)

	categoryService := service.NewCategoryService(repo, d.v)
	toyService := service.NewToyService(repo, d.r2, d.v)
	toyImageService := service.NewToyImageService(repo, d.r2)
	orderService := service.NewOrderService(repo, d.bot, d.v, d.cfg.Site.URL)
	userService := service.NewUserService(repo, d.v)
	authService := service.NewAuthService(repo, d.tokens, d.v)

	secureCookies := d.cfg.Environment == config.EnvProduction

	return router.Handlers{
		Health:   handler.NewHealthHandler(d.pool, d.cfg.Database.PingTimeout),
		Category: handler.NewCategoryHandler(categoryService),
		Toy:      handler.NewToyHandler(toyService, d.r2),
		ToyImage: handler.NewToyImageHandler(toyImageService, d.r2),
		Order:    handler.NewOrderHandler(orderService),
		User:     handler.NewUserHandler(userService),
		Auth:     handler.NewAuthHandler(authService, userService, secureCookies),
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
