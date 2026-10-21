package config

import (
	"testing"
	"time"

	"github.com/fc1g/kachenyatko-store-api/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
		Storage: Storage{
			Endpoint:        "https://account-id.r2.cloudflarestorage.com",
			AccessKeyID:     "test-access-key",
			SecretAccessKey: "test-secret-key",
			BucketName:      "test-bucket",
			PublicURL:       "https://pub.example.com",
		},
		Logging:  Logging{Level: "info", Pretty: false},
		Database: Database{URL: "postgres://user:pass@localhost:5432/db", PingTimeout: 2 * time.Second},
		Telegram: Telegram{AccessToken: "test-access-token", ChatID: 123456789, SendTimeout: 10 * time.Second},
		Site:     Site{URL: "https://example.com"},
		Auth:     Auth{AccessPrivateKey: "dGVzdC1hY2Nlc3MtcHJpdmF0ZS1rZXk=", AccessIssuer: "test-issuer", AccessTokenTTL: defaultAccessTokenTTL},
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
	t.Setenv("R2_ENDPOINT", "https://account-id.r2.cloudflarestorage.com")
	t.Setenv("R2_ACCESS_KEY_ID", "test-access-key")
	t.Setenv("R2_SECRET_ACCESS_KEY", "test-secret-key")
	t.Setenv("R2_BUCKET_NAME", "test-bucket")
	t.Setenv("R2_PUBLIC_URL", "https://pub.example.com")
	t.Setenv("TELEGRAM_ACCESS_TOKEN", "test-access-token")
	t.Setenv("TELEGRAM_CHAT_ID", "123456789")
	t.Setenv("TELEGRAM_SEND_TIMEOUT", "")
	t.Setenv("SITE_URL", "https://example.com")
	t.Setenv("AUTH_ACCESS_PRIVATE_KEY", "dGVzdC1hY2Nlc3MtcHJpdmF0ZS1rZXk=")
	t.Setenv("AUTH_ACCESS_ISSUER", "test-issuer")
	t.Setenv("AUTH_ACCESS_TOKEN_TTL", "")
}
