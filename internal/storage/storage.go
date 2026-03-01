// Package storage implements a storage client for client request counts
package storage

import (
	"context"
	"fmt"

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

// FlushAll removes all keys from the database.
func (s *Storage) FlushAll(ctx context.Context) error {
	return s.redisDB.FlushAll(ctx).Err()
}

// Close closes the connection to the database.
func (s *Storage) Close() error {
	return s.redisDB.Close()
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
