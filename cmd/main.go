package main

import (
	"context"
	"log"
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
		log.Println(err)
		os.Exit(1)
	}

	store, err := storage.New(context.Background(), config.RedisHostname, config.RedisPort, config.RedisPassword)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer store.Close()

	server := handlers.NewServer(store, config)

	// Create a context that is canceled when SIGINT or SIGTERM is received.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Start the server in a goroutine so that it doesn't block.
	go func() {
		log.Println("serving on", server.Addr)
		if err = server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	// Block until a signal is received.
	<-ctx.Done()
	// Release the signal handler so that the receipt of a second signal force kills.
	stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err = server.Shutdown(shutdownCtx); err != nil {
		log.Fatal(err)
	}
}
