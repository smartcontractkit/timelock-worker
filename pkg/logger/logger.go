package logger

import (
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

var (
	once   sync.Once       //nolint: gochecknoglobals
	logger *zerolog.Logger //nolint: gochecknoglobals
)

// humanConsoleWriter configures the human-readable output.
var humanConsoleWriter = zerolog.ConsoleWriter{ //nolint: gochecknoglobals
	Out:        os.Stdout,
	TimeFormat: time.RFC3339,
}

// TODO: remove this method in the future refactor to use functional options in the initialization
// and keep `Logger()` free of parameters just returning the Logger instance or a default Logger.
//
// Instance returns the already initialized Logger or setup a defaut one and return it.
func Instance() *zerolog.Logger {
	// it should never create a logger with those parameters, just get the initialized Logger.
	return Logger("info", "human")
}

// Logger returns an instance of Logger.
func Logger(logLevel string, output string) *zerolog.Logger {
	if logger == nil {
		once.Do(
			func() {
				logger = newLogger(logLevel, output)
			})
	}

	return logger
}

// newLogger initializes the Logger with the given arguments.
func newLogger(logLevel string, output string) *zerolog.Logger {
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		// Do not crash on wrong user input, default to InfoLevel.
		level = zerolog.InfoLevel
	}

	var l zerolog.Logger
	switch output {
	case "human":
		l = zerolog.New(humanConsoleWriter).Level(level).With().Timestamp().Logger()
	case "json":
		l = zerolog.New(os.Stdout).Level(level).With().Timestamp().Logger()
	default:
		l = zerolog.New(humanConsoleWriter).Level(level).With().Timestamp().Logger()
	}

	return &l
}
