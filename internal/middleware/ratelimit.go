// Package middleware implements a middleware abstraction for HTTP handlers.
package middleware

import (
	"encoding/json"
	"net"
	"net/http"

	"github.com/snacksforus/distributed-rate-limiter/api/response"
	"github.com/snacksforus/distributed-rate-limiter/internal/ratelimiter/slidingwindow"
	"github.com/snacksforus/distributed-rate-limiter/internal/storage"
)

// RateLimitMiddleware is the representation for a rate limiting middleware.
type RateLimitMiddleware struct {
	rateLimiter *slidingwindow.SlidingWindow
}

// Init initializes the rate limiting middleware using storage provider s.
func Init(s *storage.Storage, rateLimit int, windowSize int) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		rateLimiter: slidingwindow.Init(s, rateLimit, windowSize),
	}
}

// RateLimit limits the rate of requests from a client.
func (rlm *RateLimitMiddleware) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientId, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			// Errors parsing the remote address are considered programmer errors.
			panic(err)
		}

		allow := rlm.rateLimiter.Allow(clientId)

		if !allow {
			w.WriteHeader(429)
			w.Header().Set("Content-Type", "application/json")
			resp := response.Error("TOO_MANY_REQUESTS", "Exceeded request rate limit")
			data, _ := json.Marshal(resp)
			_, _ = w.Write(data)
			return
		}

		next.ServeHTTP(w, r)
	})
}
