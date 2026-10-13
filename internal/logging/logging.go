package logging

import (
	"io"
	"os"
	"time"

	"github.com/fc1g/kachenyatko-store-api/internal/config"
	"github.com/rs/zerolog"
)

func New(cfg config.Logging) (zerolog.Logger, error) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		return zerolog.Logger{}, err
	}

	var out io.Writer = os.Stdout
	if cfg.Pretty {
		out = zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	}

	return zerolog.New(out).Level(level).With().Timestamp().Logger(), nil
}
