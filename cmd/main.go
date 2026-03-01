package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/snacksforus/distributed-rate-limiter/api/handlers"
	"github.com/snacksforus/distributed-rate-limiter/internal/config"
	"github.com/snacksforus/distributed-rate-limiter/internal/storage"
)

func main() {
	config, err := config.New()
	if err != nil {
		slog.Error("failed to create configuration", "error", err)
		os.Exit(1)
	}

	store, err := storage.New(context.Background(), config.RedisHostname, config.RedisPort, config.RedisPassword)
	if err != nil {
		slog.Error("failed to connect to Redis server", "error", err)
		os.Exit(1)
	}
	defer store.Close()

	server := handlers.NewServer(store, config)

	// Create a context that is canceled when SIGINT or SIGTERM is received.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Start the server in a goroutine so that it doesn't block.
	go func() {
		slog.Info("serving on", "addr", server.Addr)
		if err = server.ListenAndServe(); err != http.ErrServerClosed {
			slog.Error("HTTP server failure", "error", err)
			os.Exit(1)
		}
	}()

	// Block until a signal is received.
	<-ctx.Done()
	// Release the signal handler so that the receipt of a second signal force kills.
	stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err = server.Shutdown(shutdownCtx); err != nil {
		slog.Error("failed to shutdown HTTP server", "error", err)
		os.Exit(1)
	}
}
