package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/fc1g/kachenyatko-store-api/internal/auth"
	"github.com/fc1g/kachenyatko-store-api/internal/category"
	"github.com/fc1g/kachenyatko-store-api/internal/config"
	"github.com/fc1g/kachenyatko-store-api/internal/database"
	"github.com/fc1g/kachenyatko-store-api/internal/health"
	"github.com/fc1g/kachenyatko-store-api/internal/order"
	"github.com/fc1g/kachenyatko-store-api/internal/router"
	"github.com/fc1g/kachenyatko-store-api/internal/session"
	"github.com/fc1g/kachenyatko-store-api/internal/storage"
	"github.com/fc1g/kachenyatko-store-api/internal/telegram"
	"github.com/fc1g/kachenyatko-store-api/internal/toy"
	"github.com/fc1g/kachenyatko-store-api/internal/user"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

type waiter interface {
	Wait(ctx context.Context) error
}

type App struct {
	logger          zerolog.Logger
	db              *pgxpool.Pool
	server          *http.Server
	notifications   waiter
	shutdownTimeout time.Duration
}

func New(ctx context.Context, cfg *config.Config, logger zerolog.Logger) (*App, error) {
	tokens, err := auth.NewTokenManager(cfg.Auth.SigningKey, cfg.Auth.Issuer, cfg.Auth.TokenTTL)
	if err != nil {
		return nil, err
	}

	pool, err := database.Connect(ctx, cfg.Database.URL, cfg.Database.PingTimeout)
	if err != nil {
		return nil, err
	}

	r2, err := storage.NewR2(ctx, storage.R2Config{
		Endpoint:        cfg.Storage.Endpoint,
		AccessKeyID:     cfg.Storage.AccessKeyID,
		SecretAccessKey: cfg.Storage.SecretAccessKey,
		Bucket:          cfg.Storage.BucketName,
		PublicURL:       cfg.Storage.PublicURL,
	})
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("init r2 storage: %w", err)
	}

	bot, err := telegram.NewBot(cfg.Telegram.AccessToken, cfg.Telegram.ChatID, cfg.Telegram.SendTimeout)
	if err != nil {
		pool.Close()
		return nil, err
	}

	store := database.NewStore(pool)

	sessionService := session.NewService(store, tokens)
	userService := user.NewService(store)
	categoryService := category.NewService(store)
	toyService := toy.NewService(store, r2)
	toyImageService := toy.NewImageService(store, r2)
	orderService := order.NewService(store, bot, cfg.Frontend.URL)

	secureCookies := cfg.Environment == config.EnvProduction

	handlers := router.Handlers{
		Health:   health.NewHandler(pool, cfg.Database.PingTimeout),
		Session:  session.NewHandler(sessionService, userService, secureCookies),
		User:     user.NewHandler(userService),
		Category: category.NewHandler(categoryService),
		Toy:      toy.NewHandler(toyService, r2),
		ToyImage: toy.NewImageHandler(toyImageService, r2),
		Order:    order.NewHandler(orderService),
	}

	r := router.New(logger, handlers, tokens, cfg.HTTP.ClientIPHeader)

	server := &http.Server{
		Addr:              net.JoinHostPort(cfg.HTTP.Host, strconv.Itoa(cfg.HTTP.Port)),
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
		notifications:   orderService,
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

	a.notifications.Wait(shutdownCtx)

	a.logger.Info().Msg("server stopped")

	return nil
}

func (a *App) Close() {
	a.db.Close()
}
