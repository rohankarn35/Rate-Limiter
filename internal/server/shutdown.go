package server

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// ShutdownOption configures graceful shutdown.
type ShutdownOption struct {
	Timeout time.Duration
}

// WaitForSignal blocks until termination signals arrive.
func WaitForSignal(ctx context.Context) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		select {
		case <-c:
			cancel()
		case <-ctx.Done():
		}
	}()

	return ctx
}
