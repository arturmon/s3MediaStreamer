package app

import (
	"context"
	"os"
	"os/signal"
	"s3MediaStreamer/app/internal/logs"
	"syscall"
)

func HandleSignals(ctx context.Context, logger *logs.Logger, cancel context.CancelFunc) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		select {
		case sig := <-sigCh:
			logger.Infof("Received signal: %s. Stopping the application...", sig)
			cancel()
		case <-ctx.Done():
			// Context cancelled, exiting goroutine
		}
	}()
}
