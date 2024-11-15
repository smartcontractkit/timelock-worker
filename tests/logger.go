package tests

import (
	"io"
	"sync"

	"github.com/rs/zerolog"
)

type TestLogger interface {
	Write(p []byte) (n int, err error)
	Logger() *zerolog.Logger
	NumMessages() int
	LastMessage() string
	Messages() []string
}

type testLogger struct {
	zerologger zerolog.Logger
	writer     io.Writer
	messages   *[]string
	mutex      *sync.Mutex
}

func NewTestLogger(writer io.Writer) TestLogger {
	logger := &testLogger{
		writer:   writer,
		messages: &[]string{},
		mutex:    new(sync.Mutex),
	}
	logger.zerologger = zerolog.New(logger)
	return logger
}

func (tl testLogger) Logger() *zerolog.Logger {
	return &tl.zerologger
}

func (tl testLogger) Write(p []byte) (n int, err error) {
	tl.mutex.Lock()
	defer tl.mutex.Unlock()

	*tl.messages = append(*tl.messages, string(p))
	return tl.writer.Write(p)
}

func (tl testLogger) NumMessages() int {
	tl.mutex.Lock()
	defer tl.mutex.Unlock()

	return len(*tl.messages)
}

func (tl testLogger) LastMessage() string {
	tl.mutex.Lock()
	defer tl.mutex.Unlock()

	return (*tl.messages)[tl.NumMessages()-1]
}

func (tl testLogger) Messages() []string {
	tl.mutex.Lock()
	defer tl.mutex.Unlock()

	return append([]string{}, *tl.messages...)
}
