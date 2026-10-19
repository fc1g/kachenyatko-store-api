package config

import (
	"errors"
	"fmt"
	"io/fs"
	"time"

	"github.com/fc1g/kachenyatko-store-api/internal/validation"
	"github.com/joho/godotenv"
)

const (
	defaultPort                = 3000
	defaultReadHeaderTimeout   = 5 * time.Second
	defaultReadTimeout         = 10 * time.Second
	defaultWriteTimeout        = 10 * time.Second
	defaultIdleTimeout         = 60 * time.Second
	defaultShutdownTimeout     = 10 * time.Second
	defaultDBPingTimeout       = 2 * time.Second
	unsetTelegramChatID        = 0
	defaultTelegramSendTimeout = 10 * time.Second
)

type Environment string

const (
	EnvDevelopment Environment = "development"
	EnvProduction  Environment = "production"
	EnvTest        Environment = "test"
)

type Config struct {
	// keep oneof in sync with the Environment constants
	Environment Environment `validate:"required,oneof=development production test"`
	HTTP        HTTP
	Logging     Logging
	Database    Database
	Storage     Storage
	Telegram    Telegram
	Site        Site
}

type HTTP struct {
	Port              int           `validate:"required,port"`
	ReadHeaderTimeout time.Duration `validate:"gt=0"`
	ReadTimeout       time.Duration `validate:"gt=0"`
	WriteTimeout      time.Duration `validate:"gt=0"`
	IdleTimeout       time.Duration `validate:"gt=0"`
	ShutdownTimeout   time.Duration `validate:"gt=0"`
}

type Logging struct {
	Level  string `validate:"required,oneof=debug info warn error trace fatal panic"`
	Pretty bool
}

type Database struct {
	URL         string        `validate:"required,url"`
	PingTimeout time.Duration `validate:"gt=0"`
}

type Storage struct {
	Endpoint        string `validate:"required,url"`
	AccessKeyID     string `validate:"required"`
	SecretAccessKey string `validate:"required"`
	BucketName      string `validate:"required"`
	PublicURL       string `validate:"required,url"`
}

type Telegram struct {
	AccessToken string        `validate:"required"`
	ChatID      int64         `validate:"required"`
	SendTimeout time.Duration `validate:"gt=0"`
}

type Site struct {
	URL string `validate:"required,http_url"`
}

func Load(v validation.Validator) (*Config, error) {
	if err := godotenv.Load(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("load .env: %w", err)
	}

	env := &envReader{}
	environment := Environment(env.getString("ENVIRONMENT"))

	cfg := &Config{
		Environment: environment,
		HTTP: HTTP{
			Port:              env.getInt("PORT", defaultPort),
			ReadHeaderTimeout: env.getDuration("HTTP_READ_HEADER_TIMEOUT", defaultReadHeaderTimeout),
			ReadTimeout:       env.getDuration("HTTP_READ_TIMEOUT", defaultReadTimeout),
			WriteTimeout:      env.getDuration("HTTP_WRITE_TIMEOUT", defaultWriteTimeout),
			IdleTimeout:       env.getDuration("HTTP_IDLE_TIMEOUT", defaultIdleTimeout),
			ShutdownTimeout:   env.getDuration("HTTP_SHUTDOWN_TIMEOUT", defaultShutdownTimeout),
		},
		Logging: Logging{
			Level:  env.getString("LOG_LEVEL"),
			Pretty: env.getBool("LOG_PRETTY", environment == EnvDevelopment),
		},
		Database: Database{
			URL:         env.getString("DB_URL"),
			PingTimeout: env.getDuration("DB_PING_TIMEOUT", defaultDBPingTimeout),
		},
		Storage: Storage{
			Endpoint:        env.getString("R2_ENDPOINT"),
			AccessKeyID:     env.getString("R2_ACCESS_KEY_ID"),
			SecretAccessKey: env.getString("R2_SECRET_ACCESS_KEY"),
			BucketName:      env.getString("R2_BUCKET_NAME"),
			PublicURL:       env.getString("R2_PUBLIC_URL"),
		},
		Telegram: Telegram{
			AccessToken: env.getString("TELEGRAM_ACCESS_TOKEN"),
			ChatID:      env.getInt64("TELEGRAM_CHAT_ID", unsetTelegramChatID),
			SendTimeout: env.getDuration("TELEGRAM_SEND_TIMEOUT", defaultTelegramSendTimeout),
		},
		Site: Site{
			URL: env.getString("SITE_URL"),
		},
	}

	if err := errors.Join(env.err(), v.Struct(cfg)); err != nil {
		return nil, err
	}

	return cfg, nil
}
