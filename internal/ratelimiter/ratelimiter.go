// Package ratelimiter implements routines for rate limiting requests using a
// fixed window algorithm.
package ratelimiter

import (
	"context"
)

type Counter interface {
	IncrWithTTL(context.Context, string, int) (int, error)
}

// Limiter is the representation for a fixed window rate limiter.
type Limiter struct {
	counter       Counter
	rateLimit     int
	windowSizeSec int
}

// New initializes the fixed window rate limiter using storage provider s.
func New(c Counter, rateLimit int, windowSizeSec int) *Limiter {
	return &Limiter{
		counter:       c,
		rateLimit:     rateLimit,
		windowSizeSec: windowSizeSec,
	}
}

// Allow reports whether a request from a client with clientID is within the rate limit.
//
// The rate limiting algorithm used is a fixed window which allows bursts of requests
// around the window boundary.
func (l *Limiter) Allow(ctx context.Context, clientID string) bool {
	count, err := l.counter.IncrWithTTL(ctx, clientID, l.windowSizeSec)
	if err != nil {
		// Fail open, allow the request if there is an error connecting to the database.
		return true
	}

	return count <= l.rateLimit
}
