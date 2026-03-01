// Package middleware implements a middleware abstraction for HTTP handlers.
package middleware

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strconv"

	"github.com/snacksforus/distributed-rate-limiter/api/response"
)

// RateLimit is the representation for a rate limiting middleware.
type RateLimit struct {
	allower       Allower
	windowSizeSec int
}

type Allower interface {
	Allow(context.Context, string) bool
}

// New initializes the rate limiting middleware using Allower a and a window size of
// windowSizeSec seconds.
func New(a Allower, windowSizeSec int) *RateLimit {
	return &RateLimit{
		allower:       a,
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

		allow := rlm.allower.Allow(r.Context(), clientID)

		if !allow {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", strconv.Itoa(rlm.windowSizeSec))
			w.WriteHeader(http.StatusTooManyRequests)
			resp := response.NewError("TOO_MANY_REQUESTS", "Exceeded request rate limit")
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
