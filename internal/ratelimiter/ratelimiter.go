// Package ratelimiter implements routines for rate limiting requests using a
// fixed window algorithm.
package ratelimiter

import (
	"context"

	"github.com/snacksforus/distributed-rate-limiter/internal/storage"
)

// Limiter is the representation for a fixed window rate limiter.
type Limiter struct {
	store         *storage.Storage
	rateLimit     int
	windowSizeSec int
}

// New initializes the fixed window rate limiter using storage provider s.
func New(s *storage.Storage, rateLimit int, windowSizeSec int) *Limiter {
	return &Limiter{
		store:         s,
		rateLimit:     rateLimit,
		windowSizeSec: windowSizeSec,
	}
}

// Allow reports whether a request from a client with clientID is within the rate limit.
//
// The rate limiting algorithm used is a fixed window which allows bursts of requests
// around the window boundary.
func (l *Limiter) Allow(ctx context.Context, clientID string) bool {
	count, err := l.store.IncrWithTTL(ctx, clientID, l.windowSizeSec)
	if err != nil {
		// Fail open, allow the request if there is an error connecting to the database.
		return true
	}

	return count <= l.rateLimit
}
