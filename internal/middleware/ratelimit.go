// Package middleware implements a middleware abstraction for HTTP handlers.
package middleware

import (
	"encoding/json"
	"log"
	"net"
	"net/http"

	"github.com/snacksforus/distributed-rate-limiter/api/response"
	"github.com/snacksforus/distributed-rate-limiter/internal/ratelimiter/slidingwindow"
)

// RateLimit limits the rate of requests from a client.
func RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientId, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			// Errors parsing the remote address are considered programmer errors.
			panic(err)
		}

		allow, err := slidingwindow.Allow(clientId)
		if err != nil {
			// Fail open, allow the request if there is an error with the limiter.
			log.Println(err)
			allow = true
		}

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
