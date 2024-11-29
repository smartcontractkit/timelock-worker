package tests

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func NewTestLogger() (*zap.Logger, *observer.ObservedLogs) {
	core, logs := observer.New(zap.DebugLevel)
	return zap.New(core), logs
}
