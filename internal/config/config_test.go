package config

import (
	"strconv"
	"testing"
	"time"

	"github.com/fc1g/kachenyatko-store-api/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetIntEnv(t *testing.T) {
	key := "TEST_INT"
	fallback := 8080

	tests := []struct {
		name    string
		value   string
		want    int
		wantErr error
	}{
		{
			name: "valid value", value: "3000", want: 3000,
		},
		{
			name: "empty returns fallback", value: "", want: fallback,
		},
		{
			name: "invalid returns fallback and error", value: "INT", want: fallback, wantErr: strconv.ErrSyntax,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(key, tt.value)

			got, err := getIntEnv(key, fallback)

			assert := assert.New(t)

			if tt.wantErr != nil {
				assert.ErrorIs(err, tt.wantErr)
				assert.ErrorContains(err, key)
			} else {
				assert.NoError(err)
			}
			assert.Equal(tt.want, got)
		})
	}
}

func TestGetDurationEnv(t *testing.T) {
	key := "TEST_DURATION"
	fallback := 5 * time.Second

	tests := []struct {
		name    string
		value   string
		want    time.Duration
		wantErr bool
	}{
		{
			name: "valid value", value: "3s", want: 3 * time.Second,
		},
		{
			name: "empty returns fallback", value: "", want: fallback,
		},
		{
			name: "invalid returns fallback and error", value: "DURATION", want: fallback, wantErr: true,
		},
		{
			name: "missing unit returns fallback and error", value: "10", want: fallback, wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(key, tt.value)

			got, err := getDurationEnv(key, fallback)

			assert := assert.New(t)

			if tt.wantErr {
				assert.ErrorContains(err, key)
			} else {
				assert.NoError(err)
			}
			assert.Equal(tt.want, got)
		})
	}
}

func TestLoad_ValidEnv(t *testing.T) {
	setValidEnv(t)

	cfg, err := Load(validation.New())

	require.NoError(t, err)
	assert.Equal(t, &Config{
		Environment: "test",
		HTTP: HTTP{
			Port:              3000,
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       10 * time.Second,
			WriteTimeout:      10 * time.Second,
			IdleTimeout:       60 * time.Second,
			ShutdownTimeout:   10 * time.Second,
		},
		Logging:  Logging{Level: "info", Pretty: false},
		Database: Database{URL: "postgres://user:pass@localhost:5432/db"},
	}, cfg)
}

func TestLoad_InvalidEnv(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		value       string
		wantErrText []string
	}{
		{
			name: "missing environment", key: "ENVIRONMENT", value: "", wantErrText: []string{"Environment", "required"},
		},
		{
			name: "invalid environment", key: "ENVIRONMENT", value: "staging", wantErrText: []string{"Environment", "oneof"},
		},
		{
			name: "missing log level", key: "LOG_LEVEL", value: "", wantErrText: []string{"Level", "required"},
		},
		{
			name: "invalid log level", key: "LOG_LEVEL", value: "verbose", wantErrText: []string{"Level", "oneof"},
		},
		{
			name: "missing db url", key: "DB_URL", value: "", wantErrText: []string{"URL", "required"},
		},
		{
			name: "invalid db url", key: "DB_URL", value: "not a url", wantErrText: []string{"URL", "url"},
		},
		{
			name: "port out of range", key: "PORT", value: "70000", wantErrText: []string{"Port", "port"},
		},
		{
			name: "non-numeric port", key: "PORT", value: "abc", wantErrText: []string{"PORT"},
		},
		{
			name: "read timeout without unit", key: "HTTP_READ_TIMEOUT", value: "10", wantErrText: []string{"HTTP_READ_TIMEOUT"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setValidEnv(t)
			t.Setenv(tt.key, tt.value)

			cfg, err := Load(validation.New())
			assert := assert.New(t)

			require.Error(t, err)
			assert.Nil(cfg)
			for _, text := range tt.wantErrText {
				assert.ErrorContains(err, text)
			}
		})
	}
}

func TestLoad_JoinsMultipleParseErrors(t *testing.T) {
	setValidEnv(t)
	t.Setenv("PORT", "abc")
	t.Setenv("HTTP_IDLE_TIMEOUT", "ghs")

	cfg, err := Load(validation.New())
	assert := assert.New(t)

	require.Error(t, err)
	assert.Nil(cfg)

	assert.ErrorContains(err, "PORT")
	assert.ErrorContains(err, "HTTP_IDLE_TIMEOUT")
}

func TestLoad_CustomTimeouts(t *testing.T) {
	setValidEnv(t)
	t.Setenv("HTTP_READ_HEADER_TIMEOUT", "1s")
	t.Setenv("HTTP_READ_TIMEOUT", "2s")
	t.Setenv("HTTP_WRITE_TIMEOUT", "3s")
	t.Setenv("HTTP_IDLE_TIMEOUT", "4m")
	t.Setenv("HTTP_SHUTDOWN_TIMEOUT", "500ms")

	cfg, err := Load(validation.New())
	assert := assert.New(t)

	require.NoError(t, err)
	assert.Equal(1*time.Second, cfg.HTTP.ReadHeaderTimeout)
	assert.Equal(2*time.Second, cfg.HTTP.ReadTimeout)
	assert.Equal(3*time.Second, cfg.HTTP.WriteTimeout)
	assert.Equal(4*time.Minute, cfg.HTTP.IdleTimeout)
	assert.Equal(500*time.Millisecond, cfg.HTTP.ShutdownTimeout)
}

func TestLoad_LogPretty(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{
			name: "true", value: "true", want: true,
		},
		{
			name: "numeric one", value: "1", want: true,
		},
		{
			name: "false", value: "false", want: false,
		},
		{
			name: "garbage falls back to false", value: "garbage", want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setValidEnv(t)
			t.Setenv("LOG_PRETTY", tt.value)

			cfg, err := Load(validation.New())

			require.NoError(t, err)
			assert.Equal(t, tt.want, cfg.Logging.Pretty)
		})
	}
}

func setValidEnv(t *testing.T) {
	t.Helper()
	t.Chdir(t.TempDir())

	t.Setenv("ENVIRONMENT", "test")
	t.Setenv("PORT", "3000")
	t.Setenv("LOG_LEVEL", "info")
	t.Setenv("LOG_PRETTY", "")
	t.Setenv("DB_URL", "postgres://user:pass@localhost:5432/db")
	t.Setenv("HTTP_READ_HEADER_TIMEOUT", "")
	t.Setenv("HTTP_READ_TIMEOUT", "")
	t.Setenv("HTTP_WRITE_TIMEOUT", "")
	t.Setenv("HTTP_IDLE_TIMEOUT", "")
	t.Setenv("HTTP_SHUTDOWN_TIMEOUT", "")
}
