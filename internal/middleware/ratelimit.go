// Package middleware implements a middleware abstraction for HTTP handlers.
package middleware

import (
	"encoding/json"
	"net"
	"net/http"
	"strconv"

	"github.com/snacksforus/distributed-rate-limiter/api/response"
	"github.com/snacksforus/distributed-rate-limiter/internal/ratelimiter"
	"github.com/snacksforus/distributed-rate-limiter/internal/storage"
)

// RateLimit is the representation for a rate limiting middleware.
type RateLimit struct {
	rateLimiter   *ratelimiter.Limiter
	windowSizeSec int
}

// New initializes the rate limiting middleware using storage s, with a limit of rateLimit,
// and a window size of windowSizeSec seconds.
func New(s *storage.Storage, rateLimit int, windowSizeSec int) *RateLimit {
	return &RateLimit{
		rateLimiter:   ratelimiter.New(s, rateLimit, windowSizeSec),
		windowSizeSec: windowSizeSec,
	}
}

// Handler limits the rate of requests from a client.
func (rlm *RateLimit) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientID, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		allow := rlm.rateLimiter.Allow(r.Context(), clientID)

		if !allow {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", strconv.Itoa(rlm.windowSizeSec))
			w.WriteHeader(http.StatusTooManyRequests)
			resp := response.Error("TOO_MANY_REQUESTS", "Exceeded request rate limit")
			var data []byte
			data, err = json.Marshal(resp)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			_, _ = w.Write(data)
			return
		}

		next.ServeHTTP(w, r)
	})
}
