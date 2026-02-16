// Package slidingwindow implements routines for rate limiting HTTP requests using a
// sliding window algorithm.
package slidingwindow

import (
	"log"
	"time"

	"github.com/snacksforus/distributed-rate-limiter/internal/storage"
)

// SlidingWindow is the representation for a sliding window rate limiter.
type SlidingWindow struct {
	db *storage.Storage
}

// Init initializes the sliding window rate limiter using storage provider s.
func Init(s *storage.Storage) *SlidingWindow {
	return &SlidingWindow{
		db: s,
	}
}

// Allow reports whether a request from a client with clientId is within the rate limit.
//
// The rate limiting algorithm used is an approximation of a sliding window.  The algorithm
// allows bursts of requests around the window boundary.
func (sw SlidingWindow) Allow(clientId string) bool {
	// The rate limiting algorithm is an approximation of a sliding window.  The TTL of the count of requests for
	// the clientId is the size of the window.  Redis will automatically expire the count.  This approach does
	// allow bursts of requests at the window boundary, but this is an acceptable tradeoff for the ease of
	// implementation.

	count, err := sw.db.IncrWithExpr(clientId, 10*time.Second)
	if err != nil {
		log.Println(err)
		// Fail open, allow the request if there is an error connecting to the database.
		return true
	}

	return count < 10
}
