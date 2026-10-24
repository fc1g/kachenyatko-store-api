package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/fc1g/kachenyatko-store-api/internal/validation"
	"github.com/joho/godotenv"
)

type Environment string

const (
	EnvDevelopment Environment = "development"
	EnvProduction  Environment = "production"
)

type Config struct {
	Environment Environment `env:"ENVIRONMENT" validate:"required,oneof=development production"`
	HTTP        HTTP
	Database    Database
	Storage     Storage
	Telegram    Telegram
	Auth        Auth
	Logging     Logging
	Frontend    Frontend
}

type HTTP struct {
	Host              string        `env:"HTTP_HOST"                envDefault:"127.0.0.1" validate:"required,ip"`
	Port              int           `env:"PORT"                     envDefault:"8080" validate:"required,port"`
	ReadHeaderTimeout time.Duration `env:"HTTP_READ_HEADER_TIMEOUT" envDefault:"5s"   validate:"gt=0"`
	ReadTimeout       time.Duration `env:"HTTP_READ_TIMEOUT"        envDefault:"10s"  validate:"gt=0"`
	WriteTimeout      time.Duration `env:"HTTP_WRITE_TIMEOUT"       envDefault:"10s"  validate:"gt=0"`
	IdleTimeout       time.Duration `env:"HTTP_IDLE_TIMEOUT"        envDefault:"60s"  validate:"gt=0"`
	ShutdownTimeout   time.Duration `env:"HTTP_SHUTDOWN_TIMEOUT"    envDefault:"10s"  validate:"gt=0"`
	ClientIPHeader    string        `env:"HTTP_CLIENT_IP_HEADER"`
}

type Database struct {
	URL         string        `env:"DB_URL"          validate:"required,url"`
	PingTimeout time.Duration `env:"DB_PING_TIMEOUT" envDefault:"2s" validate:"gt=0"`
}

type Storage struct {
	Endpoint        string `env:"R2_ENDPOINT"          validate:"required,url"`
	AccessKeyID     string `env:"R2_ACCESS_KEY_ID"     validate:"required"`
	SecretAccessKey string `env:"R2_SECRET_ACCESS_KEY" validate:"required"`
	BucketName      string `env:"R2_BUCKET_NAME"       validate:"required"`
	PublicURL       string `env:"R2_PUBLIC_URL"        validate:"required,url"`
}

type Telegram struct {
	AccessToken string        `env:"TELEGRAM_ACCESS_TOKEN" validate:"required"`
	ChatID      int64         `env:"TELEGRAM_CHAT_ID"      validate:"required"`
	SendTimeout time.Duration `env:"TELEGRAM_SEND_TIMEOUT" envDefault:"10s" validate:"gt=0"`
}

type Logging struct {
	Level  string `env:"LOG_LEVEL" validate:"required,oneof=debug info warn error trace fatal panic"`
	Pretty bool   `env:"LOG_PRETTY"`
}

type Auth struct {
	SigningKey string        `env:"AUTH_SIGNING_KEY" validate:"required,base64"`
	Issuer     string        `env:"AUTH_ISSUER"      validate:"required"`
	TokenTTL   time.Duration `env:"AUTH_TOKEN_TTL"   envDefault:"24h" validate:"gt=0,lte=24h"`
}

type Frontend struct {
	URL string `env:"SITE_URL" validate:"required,http_url"`
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("load .env: %w", err)
	}

	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return nil, fmt.Errorf("parse env: %w", err)
	}

	if _, ok := os.LookupEnv("LOG_PRETTY"); !ok {
		cfg.Logging.Pretty = cfg.Environment == EnvDevelopment
	}

	if err := validation.Struct(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
