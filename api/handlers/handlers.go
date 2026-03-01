package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/snacksforus/distributed-rate-limiter/api/response"
	"github.com/snacksforus/distributed-rate-limiter/internal/config"
	"github.com/snacksforus/distributed-rate-limiter/internal/middleware"
	"github.com/snacksforus/distributed-rate-limiter/internal/ratelimiter"
)

// handler handles HTTP Get requests for the demo API endpoint, returns a JSON
// success message.
func handler(w http.ResponseWriter, r *http.Request) {
	resp := response.Success()
	data, err := json.Marshal(resp)
	if err != nil {
		slog.Error("failed to marshal API response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(data)
}

// NewServer returns an HTTP server configured using config and backed with the counter request counter.
// The demo API and rate limiting middleware handlers are registered with the server.
func NewServer(counter ratelimiter.Counter, config *config.Config) *http.Server {
	timeout := time.Duration(config.TimeoutMS) * time.Millisecond

	rl := ratelimiter.New(counter, config.RateLimit, config.WindowSizeSec)
	mw := middleware.New(rl, config.WindowSizeSec)
	apiHandler := mw.Handler(http.HandlerFunc(handler))

	mux := http.NewServeMux()
	mux.Handle("GET /api", http.TimeoutHandler(apiHandler, timeout, "request timed out"))

	server := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", config.Hostname, config.Port),
		ReadHeaderTimeout: time.Duration(config.ReadHeaderTimeoutMS) * time.Millisecond,
		ReadTimeout:       time.Duration(config.ReadTimeoutMS) * time.Millisecond,
		Handler:           mux,
	}

	return server
}
