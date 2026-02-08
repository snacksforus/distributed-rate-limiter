package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/snacksforus/distributed-rate-limiter/api/response"
	"github.com/snacksforus/distributed-rate-limiter/internal/middleware"
)

func handler(w http.ResponseWriter, r *http.Request) {
	resp := response.Success()
	data, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(data)
}

func main() {
	// API has a single endpoint that just returns success.
	http.Handle("/", middleware.RateLimit(http.HandlerFunc(handler)))

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
