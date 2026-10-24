package logging

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

func New(level string, pretty bool) (zerolog.Logger, error) {
	parsedLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		return zerolog.Logger{}, err
	}

	var out io.Writer = os.Stdout
	if pretty {
		out = zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	}

	return zerolog.New(out).Level(parsedLevel).With().Timestamp().Logger(), nil
}

func Setup(level string, pretty bool) (zerolog.Logger, error) {
	logger, err := New(level, pretty)
	if err != nil {
		return zerolog.Logger{}, err
	}

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	zerolog.DefaultContextLogger = &logger

	return logger, nil
}
