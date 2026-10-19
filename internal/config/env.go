package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

type envReader struct {
	errors []error
}

func (r *envReader) getString(key string) string {
	return os.Getenv(key)
}

func (r *envReader) getInt(key string, fallback int) int {
	return getEnv(r, key, fallback, strconv.Atoi)
}

func (r *envReader) getInt64(key string, fallback int64) int64 {
	return getEnv(r, key, fallback, func(s string) (int64, error) {
		return strconv.ParseInt(s, 10, 64)
	})
}

func (r *envReader) getBool(key string, fallback bool) bool {
	return getEnv(r, key, fallback, strconv.ParseBool)
}

func (r *envReader) getDuration(key string, fallback time.Duration) time.Duration {
	return getEnv(r, key, fallback, time.ParseDuration)
}

func (r *envReader) err() error {
	return errors.Join(r.errors...)
}

func getEnv[T any](r *envReader, key string, fallback T, parse func(string) (T, error)) T {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}

	v, err := parse(raw)
	if err != nil {
		r.errors = append(r.errors, fmt.Errorf("%s: %w", key, err))
		return fallback
	}

	return v
}
