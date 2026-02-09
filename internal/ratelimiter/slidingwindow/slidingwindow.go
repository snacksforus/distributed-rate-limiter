// Package slidingwindow provides a sliding window rate limiter.
package slidingwindow

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

func Allow(clientId string) (bool, error) {
	// Creating a Redis connection for each request is a poor design, but is okay for this
	// stage of development.  Refactor so that a single connection is used by all requests.
	rdb := redis.NewClient(&redis.Options{
		Addr:     "drl-redis:6379",
		Password: "",
		DB:       0,
		Protocol: 2,
	})
	defer rdb.Close()

	ctx := context.Background()
	val, err := rdb.Ping(ctx).Result()
	if err != nil {
		return false, err
	}

	// Rate limiting algorithm is an approximation of a sliding window.  Set the TTL of
	// the count of requests stored in Redis to the size of the window.  Redis will
	// automatically expire the account.  This approach does allow bursts of requests
	// before and after the count is expired.  That is an acceptable tradeoff for the
	// ease of implementation.  The limit enforcement doesn't need to be exact.

	// There is a problem with this implementation for getting and setting the count.
	// It requires two round trips to the database to get and set the count for the client.
	// Even worse, there is a race condition when multiple database connections are
	// incrementing the count for a client ID.

	val, err = rdb.Get(ctx, clientId).Result()
	if err != nil && err != redis.Nil {
		return false, err
	}

	var count int64
	if err != nil {
		// client ID was not found in the database, set initial count
		count = 0
	} else {
		count, err = strconv.ParseInt(val, 10, 64)
		if err != nil {
			// Fail safely if there is an error parsing the count from the database by
			// setting the count to zero.
			log.Println(err)
			count = 0
		}
	}

	count++

	val, err = rdb.Set(ctx, clientId, count, 10*time.Second).Result()
	if err != nil {
		return false, err
	}

	return count < 10, nil
}
