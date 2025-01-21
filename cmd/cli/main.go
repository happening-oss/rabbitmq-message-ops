package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

// endregion

func main() {
	// Build CLI app
	app := buildCLIApp()
	// Graceful shutdown in case of SIGINT or SIGTERM
	ctx, cancel := context.WithCancel(context.Background())
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		select {
		case <-signals:
		case <-ctx.Done():
		}
		cancel()
	}()
	// Run CLI app
	if err := app.RunContext(ctx, os.Args); err != nil {
		log.Error("error occurred while running cli app", slog.Any("error", err))
		os.Exit(1)
	}
}
