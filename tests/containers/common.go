package containers

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/testcontainers/testcontainers-go"
	tcexec "github.com/testcontainers/testcontainers-go/exec"
)

// StdoutLogConsumer is a LogConsumer that prints the log to stdout.
type StdoutLogConsumer struct{ Prefix string }

// Accept prints the log to stdout.
func (lc *StdoutLogConsumer) Accept(l testcontainers.Log) {
	fmt.Printf("%s%v", lc.Prefix, string(l.Content)) //nolint:forbidigo
}

// exec is a convenience function to execute a command in a container and
// return the output as a string.
func exec(ctx context.Context, container testcontainers.Container, command []string) (int, string, error) {
	statusCode, outputReader, err := container.Exec(ctx, command, tcexec.Multiplexed())
	if err != nil {
		return 0, "", fmt.Errorf("error executing command: %w", err)
	}

	buf := new(strings.Builder)
	_, err = io.Copy(buf, outputReader)
	if err != nil {
		return 0, "", fmt.Errorf("error reading output from io.Reader: %w", err)
	}

	return statusCode, buf.String(), nil
}

// execUntil executes the given command in the container until the `until`
// function returns `true`.
func execUntil(
	ctx context.Context, container testcontainers.Container, command []string,
	until func(int, string) error, interval time.Duration, maxElapsedTime time.Duration,
) (statusCode int, output string, err error) { //nolint:unparam
	maxTimestamp := time.Now().Add(maxElapsedTime)
	for {
		statusCode, output, err = exec(ctx, container, command)
		if err == nil {
			err = until(statusCode, output)
			if err == nil {
				return statusCode, output, nil
			}
		}
		time.Sleep(interval)
		if time.Now().After(maxTimestamp) {
			break
		}
	}
	return 0, "", err
}
