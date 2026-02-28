package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

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

	handlers.Register(store, config)

	addr := fmt.Sprintf("%s:%d", config.Hostname, config.Port)
	log.Println("serving on", addr)
	if err = http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
