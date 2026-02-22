// Package storage implements a storage client for client request counts
package storage

import (
	"context"
	"fmt"
	"strconv"

	"github.com/redis/go-redis/v9"
)

// Storage is the representation for Redis database connection.
type Storage struct {
	redisDB *redis.Client
}

// New initializes the database connection.  An error is returned if the database can't be reached.
func New(ctx context.Context, hostname string, port int, password string) (*Storage, error) {
	addr := fmt.Sprintf("%s:%d", hostname, port)
	s := Storage{
		redisDB: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
		}),
	}

	_, err := s.redisDB.Ping(ctx).Result()
	if err != nil {
		s.redisDB.Close()
		return nil, err
	}

	return &s, nil
}

// Close closes the connection to the database.
func (s *Storage) Close() error {
	return s.redisDB.Close()
}

// GetCount returns the count value for the clientID key.  A SetCount followed by a GetCount is
// not an atomic operation.
func (s *Storage) GetCount(ctx context.Context, clientID string) (int, error) {
	val, err := s.redisDB.Get(ctx, clientID).Result()

	if err == redis.Nil {
		// clientID was not found in the database, return initial count
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	var count int
	count, err = strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("error parsing key for clientId %v: %w", clientID, err)
	}

	return count, nil
}

// SetCount sets the value for the clientID key to count.  The clientID key has no expiration.
// A SetCount followed by a GetCount is not an atomic operation.
func (s *Storage) SetCount(ctx context.Context, clientID string, count int) error {
	_, err := s.redisDB.Set(ctx, clientID, count, 0).Result()
	return err
}

// incrScript is a Lua script executed in Redis that increments a key and returns the updated count.
// New keys are set to expire.
var incrScript = redis.NewScript(`
local count = redis.call('INCR', KEYS[1])
if count == 1 then
	redis.call('EXPIRE', KEYS[1], ARGV[1])
end
return count
`)

// IncrWithTTL atomically increments the value for the clientID key by one, and sets the expiration to
// ttl seconds for new keys. The updated count for the key is returned.
func (s *Storage) IncrWithTTL(ctx context.Context, clientID string, ttl int) (int, error) {
	return incrScript.Run(ctx, s.redisDB, []string{clientID}, ttl).Int()
}
