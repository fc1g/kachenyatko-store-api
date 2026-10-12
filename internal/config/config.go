package config

import (
	"os"
	"strconv"

	"github.com/fc1g/kachenyatko-store-api/internal/validation"
	"github.com/joho/godotenv"
)

type Config struct {
	Environment string `validate:"required,oneof=development production test"`
	Port        string `validate:"required,port"`
	Logging     Logging
	Database    Database
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

	cfg := &Config{
		Environment: os.Getenv("ENVIRONMENT"),
		Port:        os.Getenv("PORT"),
		Logging: Logging{
			Level:  os.Getenv("LOG_LEVEL"),
			Pretty: logPretty,
		},
		Database: Database{
			URL: os.Getenv("DB_URL"),
		},
	}

	if err := v.Struct(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
