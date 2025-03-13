package timelock

import (
	"context"
	"time"

	"github.com/avast/retry-go/v4"
)

var (
	retryMinDelay          = 500 * time.Millisecond
	retryAttempts          = uint(5)
	retryIncrementalDelays = [4]int{500, 2000, 8000, 32000}
	retryContextTimeout    = 30 * time.Second
	retryOpts              = func(ctx context.Context) []retry.Option {
		return []retry.Option{
			retry.Context(ctx),
			retry.DelayType(func(n uint, _ error, config *retry.Config) time.Duration {
				return time.Duration(retryIncrementalDelays[min(int(n), len(retryIncrementalDelays)-1)]) * time.Millisecond
			}),
			retry.Delay(retryMinDelay),
			retry.Attempts(uint(len(retryIncrementalDelays) + 1)),
			retry.LastErrorOnly(true),
			// retry.OnRetry(func(attempt uint, err error) {
			// 	fmt.Printf("RETRYING: %d, %s\n", attempt, err)
			// }),
		}
	}
)

type retryCallback[T any] func(ctx context.Context) (T, error)

func Retry[T any](ctx context.Context, callback retryCallback[T]) (T, error) {
	var returnValue T
	var err error

	err = retry.Do(func() error {
		ctx, cancel := context.WithTimeout(ctx, retryContextTimeout)
		defer cancel()

		returnValue, err = callback(ctx)

		return err
	}, retryOpts(ctx)...)

	return returnValue, err
}
