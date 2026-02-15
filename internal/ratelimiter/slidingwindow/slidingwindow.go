// Package slidingwindow implements routines for rate limiting HTTP requests using a
// sliding window algorithm.
package slidingwindow

import (
	"log"

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
	// Rate limiting algorithm is an approximation of a sliding window.  Set the TTL of
	// the count of requests stored in Redis to the size of the window.  Redis will
	// automatically expire the account.  This approach does allow bursts of requests
	// before and after the count is expired.  That is an acceptable tradeoff for the
	// ease of implementation.  The limit enforcement doesn't need to be exact.

	// There is a problem with this implementation for getting and setting the count.
	// It requires two round trips to the database to get and set the count for the client.
	// Even worse, there is a race condition when multiple database connections are
	// incrementing the count for a client ID.

	count, err := sw.db.GetCount(clientId)
	if err != nil {
		// Fail open, allow the request if there is an error connecting to the database.
		log.Println(err)
		return true
	}

	count++

	err = sw.db.SetCount(clientId, count)
	if err != nil {
		// Fail open, allow the request if there is an error connecting to the database.
		return true
	}

	return count < 10
}
