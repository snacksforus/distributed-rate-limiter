package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/snacksforus/distributed-rate-limiter/api/response"
	"github.com/snacksforus/distributed-rate-limiter/internal/config"
	"github.com/snacksforus/distributed-rate-limiter/internal/middleware"
	"github.com/snacksforus/distributed-rate-limiter/internal/storage"
)

func handler(w http.ResponseWriter, r *http.Request) {
	resp := response.Success()
	data, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(data)
}

func main() {
	config, err := config.New()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	db, err := storage.New(context.Background(), config.RedisHostname, config.RedisPort, config.RedisPassword)
	if err != nil {
		log.Println(err)
		db.Close()
		os.Exit(1)
	}
	defer db.Close()

	mw := middleware.New(db, config.RateLimit, config.WindowSizeSec)

	// API has a single endpoint that just returns success.
	http.Handle("/", mw.Handler(http.HandlerFunc(handler)))

	addr := fmt.Sprintf("%s:%d", config.Hostname, config.Port)
	log.Println("serving on", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
