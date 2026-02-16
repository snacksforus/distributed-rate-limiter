// Package storage implements a storage client for client request counts
package storage

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// Storage is the representation for Redis database connection.
type Storage struct {
	redisDB  *redis.Client
	redisCtx context.Context
}

// Init initializes the database connection.  An error is returned if the database can't be reached.
func Init() (*Storage, error) {
	s := Storage{
		redisDB: redis.NewClient(&redis.Options{
			Addr:     "drl-redis:6379",
			Password: "",
			DB:       0,
			Protocol: 2,
		}),
		redisCtx: context.Background(),
	}

	_, err := s.redisDB.Ping(s.redisCtx).Result()
	if err != nil {
		return nil, err
	}

	return &s, nil
}

// Close closes the connection to the database.
func (s *Storage) Close() error {
	return s.redisDB.Close()
}

// GetCount returns the count value for the clientId key.
func (s *Storage) GetCount(clientId string) (int64, error) {
	val, err := s.redisDB.Get(s.redisCtx, clientId).Result()
	if err != nil && err != redis.Nil {
		return 0, err
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

	return count, nil
}

// SetCount sets the value for the clientId key to count.
func (s *Storage) SetCount(clientId string, count int64) error {
	_, err := s.redisDB.Set(s.redisCtx, clientId, count, 10*time.Second).Result()
	return err
}

// incrScript is a Lua script executed in Redis the increments a key, and returns the updated count.
// New keys are set to expire.
var incrScript = redis.NewScript(`
local count = redis.call('INCR', KEYS[1])
if count == 1 then
	redis.call('EXPIRE', KEYS[1], ARGV[1])
end
return count
`)

// IncrWithExpr increments the value for the clientId key by one, and sets the expiration to ttl for new keys.
// The updated count for the key is returned.
func (s *Storage) IncrWithExpr(clientId string, ttl time.Duration) (int64, error) {
	count, err := incrScript.Run(s.redisCtx, s.redisDB, []string{clientId}, ttl.Seconds()).Int64()
	if err != nil {
		return 0, err
	}

	return count, nil
}
