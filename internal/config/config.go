package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/fc1g/kachenyatko-store-api/internal/validation"
	"github.com/joho/godotenv"
)

type Config struct {
	Environment string `validate:"required,oneof=development production test"`
	HTTP        HTTP
	Logging     Logging
	Database    Database
}

type HTTP struct {
	Port              string        `validate:"required,port"`
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
	URL string `validate:"required,url"`
}

func Load(v validation.Validator) (*Config, error) {
	_ = godotenv.Load()

	logPretty, _ := strconv.ParseBool(os.Getenv("LOG_PRETTY"))

	var httpErrors []error

	getDuration := func(key string, fallback time.Duration) time.Duration {
		d, err := getDurationEnv(key, fallback)
		if err != nil {
			httpErrors = append(httpErrors, err)
		}
		return d
	}

	cfg := &Config{
		Environment: os.Getenv("ENVIRONMENT"),
		HTTP: HTTP{
			Port:              os.Getenv("PORT"),
			ReadHeaderTimeout: getDuration("HTTP_READ_HEADER_TIMEOUT", 5*time.Second),
			ReadTimeout:       getDuration("HTTP_READ_TIMEOUT", 10*time.Second),
			WriteTimeout:      getDuration("HTTP_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:       getDuration("HTTP_IDLE_TIMEOUT", 60*time.Second),
			ShutdownTimeout:   getDuration("HTTP_SHUTDOWN_TIMEOUT", 10*time.Second),
		},
		Logging: Logging{
			Level:  os.Getenv("LOG_LEVEL"),
			Pretty: logPretty,
		},
		Database: Database{
			URL: os.Getenv("DB_URL"),
		},
	}

	if err := errors.Join(httpErrors...); err != nil {
		return nil, err
	}

	if err := v.Struct(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func getDurationEnv(key string, fallback time.Duration) (time.Duration, error) {
	v := os.Getenv(key)
	if v == "" {
		return fallback, nil
	}

	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", key, err)
	}

	return d, nil
}
