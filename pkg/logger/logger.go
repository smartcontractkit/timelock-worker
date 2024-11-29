package logger

import (
	"fmt"

	"go.uber.org/zap"
)

// NewLogger initializes the Logger with the given arguments.
func NewLogger(logLevel string, output string) (*zap.Logger, error) {
	var err error
	var loggerConfig zap.Config

	switch output {
	case "json":
		loggerConfig = zap.NewProductionConfig()
	case "human":
		loggerConfig = zap.NewDevelopmentConfig()
	default:
		return nil, fmt.Errorf("invalid logger output: %q", output)
	}

	loggerConfig.Level, err = zap.ParseAtomicLevel(logLevel)
	if err != nil {
		return nil, err
	}

	// use stdout, not stderr to maintain compatibility with zerolog
	loggerConfig.OutputPaths = []string{"stdout"}
	loggerConfig.ErrorOutputPaths = []string{"stdout"}

	return loggerConfig.Build()
}
