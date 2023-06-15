package logger_test

import (
	"reflect"
	"testing"

	"github.com/rs/zerolog"
	"github.com/smartcontractkit/timelock-worker/pkg/logger"
	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	testLogger := logger.Logger("info", "human")
	assert.NotNil(t, testLogger)

	testType := reflect.TypeOf(testLogger)
	loggerType := reflect.TypeOf((*zerolog.Logger)(nil))

	if !(loggerType == testType) {
		t.Errorf("testType does not implement loggerType")
	}
}

func TestLoggerJSON(t *testing.T) {
	testLogger := logger.Logger("debug", "json")
	assert.NotNil(t, testLogger)

	testType := reflect.TypeOf(testLogger)
	loggerType := reflect.TypeOf((*zerolog.Logger)(nil))

	if !(loggerType == testType) {
		t.Errorf("testType does not implement loggerType")
	}
}

func TestLoggerWrongLogLevel(t *testing.T) {
	testLogger := logger.Logger("wrongLogLevel", "human")
	assert.NotNil(t, testLogger)

	testType := reflect.TypeOf(testLogger)
	loggerType := reflect.TypeOf((*zerolog.Logger)(nil))

	if !(loggerType == testType) {
		t.Errorf("testType does not implement loggerType")
	}
}

func TestLoggerWrongAll(t *testing.T) {
	testLogger := logger.Logger("wrongLogLevel", "wrongOutput")
	assert.NotNil(t, testLogger)

	testType := reflect.TypeOf(testLogger)
	loggerType := reflect.TypeOf((*zerolog.Logger)(nil))

	if !(loggerType == testType) {
		t.Errorf("testType does not implement loggerType")
	}
}
