package config

import (
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

type Config struct {
	Environment string `validate:"required,oneof=development production test"`
	Logging     Logging
	Database    Database
}

type Logging struct {
	Level  string `validate:"oneof=debug info warn error trace fatal panic"`
	Pretty bool
}

type Database struct {
	URL string `validate:"required,url"`
}

func Load(v *validator.Validate) (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Environment: getEnv("ENVIRONMENT"),
		Logging: Logging{
			Level:  getEnv("LOG_LEVEL"),
			Pretty: getEnv("LOG_PRETTY") == "true",
		},
		Database: Database{
			URL: getEnv("DB_URL"),
		},
	}

	if err := v.Struct(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func getEnv(key string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return ""
}
