package isclosed

import (
	"context"
	"sync"
)

// All returns a channel that is closed either when each of the channels is read from or the passed
// context is canceled.  All is useful for implementing graceful shutdown of a number of Sends or
// other running goroutines that indicate their state via a returned read only channel.  The graceful
// shutdown can be circumvented via the context passed to All to ensure shutdowns will not deadlock.
func All(ctx context.Context, done ...<-chan struct{}) <-chan struct{} {
	var (
		shutdown = make(chan struct{})
		wg       sync.WaitGroup
	)

	wg.Add(len(done))
	for _, ch := range done {
		go func() {
			defer wg.Done()

			select {
			case <-ctx.Done():
			case <-ch:
			}
		}()
	}

	go func() {
		defer close(shutdown)

		wg.Wait()
	}()

	return shutdown
}
