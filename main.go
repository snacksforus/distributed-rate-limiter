package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/snacksforus/distributed-rate-limiter/api/response"
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
	db, err := storage.Init()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer db.Close()

	mw := middleware.Init(db)

	// API has a single endpoint that just returns success.
	http.Handle("/", mw.RateLimit(http.HandlerFunc(handler)))

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
