package isclosed

import (
	"context"
	"testing"
	"time"
)

var maxTestTimeout = 3 * time.Second

func TestAll_DoneAfterAllClose(t *testing.T) {
	var (
		ctx     = context.Background()
		a, b, c = make(chan struct{}), make(chan struct{}), make(chan struct{})
		done    = All(ctx, a, b, c)
	)

	close(a)
	close(b)
	close(c) // close all channels
	eventually(t, done, maxTestTimeout)
}

func TestAll_DoneAfterCtxCancel(t *testing.T) {
	var (
		ctx, cancel = context.WithCancel(context.Background())
		a, b, c     = make(chan struct{}), make(chan struct{}), make(chan struct{})
		done        = All(ctx, a, b, c)
	)

	close(a)
	close(b)
	cancel() // c is never closed, but context is canceled
	eventually(t, done, maxTestTimeout)
}

func TestAll_DoneAfterCtxCancelWithNilChannels(t *testing.T) {
	var (
		ctx, cancel = context.WithCancel(context.Background())
		done        = All(ctx, nil, nil, nil)
	)

	cancel()
	eventually(t, done, maxTestTimeout)
}

// TestAll_DoneNonBlocking verifies that if all the input channels close, the All function's returned
// channel should also close regardless of the order of the input channel arguments.
func TestAll_DoneNonBlocking(t *testing.T) {
	var (
		ctx     = context.Background()
		a, b, c = make(chan struct{}), make(chan struct{}), make(chan struct{})
		start   = make(chan struct{})

		// done is only closed once all three input channels are closed as the context is never
		// cancelled.
		done = All(ctx, c, b, a)
	)

	// By default channel a closes with no dependencies.
	go func() {
		<-start
		close(a)
	}()

	// Only close channel b after channel a is closed.
	go func() {
		<-a
		close(b)
	}()

	// Only close the c channel after both channel b and channel a are closed.
	go func() {
		<-a
		<-b
		close(c)
	}()

	// Start the closing of the channels once all waiting routines are running.
	close(start)

	// Require that the done channel is eventually closed without context cancellation even with
	// dependencies on closing between various channels as long as there is no deadlock state.
	eventually(t, done, maxTestTimeout)
}

// eventually blocks until done is closed or d time duration passes.
func eventually(t *testing.T, done <-chan struct{}, d time.Duration) {
	t.Helper()

	select {
	case <-done:
	case <-time.After(d):
		t.Fatal("timed out waiting for done to close")
	}
}
