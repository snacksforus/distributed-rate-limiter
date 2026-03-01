// Package ratelimiter_test provides tests for the ratelimiter package.  The tests are intended
// to be run in a Docker container.  Test clean up is not performed since they are run in an
// isolated Docker environment.
package ratelimiter_test

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/snacksforus/distributed-rate-limiter/internal/config"
	"github.com/snacksforus/distributed-rate-limiter/internal/ratelimiter"
	"github.com/snacksforus/distributed-rate-limiter/internal/storage"
)

var (
	store *storage.Storage
	conf  *config.Config
)

type noopLogger struct{}

func (noopLogger) Printf(_ context.Context, _ string, _ ...interface{}) {}

func TestMain(m *testing.M) {
	var err error

	conf, err = config.New()
	if err != nil {
		slog.Error("failed to create configuration", "error", err)
		os.Exit(1)
	}

	redis.SetLogger(noopLogger{})

	store, err = storage.New(context.Background(), conf.RedisHostname, conf.RedisPort, conf.RedisPassword)
	if err != nil {
		slog.Error("failed to connect to Redis", "error", err)
		os.Exit(1)
	}

	exitCode := m.Run()
	store.Close()

	os.Exit(exitCode)
}

func TestAllow_BelowLimit(t *testing.T) {
	const rateLimit = 5
	clientID := t.Name()
	rl := ratelimiter.New(store, rateLimit, 60)

	for i := 0; i < rateLimit-1; i++ {
		if !rl.Allow(context.Background(), clientID) {
			t.Errorf("expected request %d to be allowed", i)
		}
	}
}

func TestAllow_AtLimit(t *testing.T) {
	const rateLimit = 5
	clientID := t.Name()
	rl := ratelimiter.New(store, rateLimit, 60)

	for i := 0; i < rateLimit; i++ {
		if !rl.Allow(context.Background(), clientID) {
			t.Errorf("expected request %d to be allowed", i)
		}
	}

	if rl.Allow(context.Background(), clientID) {
		t.Errorf("expected request %d above rate limit to be denied", rateLimit)
	}
}

func TestAllow_ExceedsLimit(t *testing.T) {
	const rateLimit = 5
	clientID := t.Name()
	rl := ratelimiter.New(store, rateLimit, 60)

	for i := 0; i < rateLimit; i++ {
		rl.Allow(context.Background(), clientID)
	}

	for i := 0; i < 3; i++ {
		if rl.Allow(context.Background(), clientID) {
			t.Errorf("expected request %d beyond rate limit to be denied", rateLimit+i+1)
		}
	}
}

func TestAllow_DifferentClients(t *testing.T) {
	const rateLimit = 5
	clientID1 := t.Name() + "_1"
	clientID2 := t.Name() + "_2"
	rl := ratelimiter.New(store, rateLimit, 60)

	for i := 0; i < rateLimit; i++ {
		rl.Allow(context.Background(), clientID1)
	}

	if !rl.Allow(context.Background(), clientID2) {
		t.Error("expected request from a different client to be allowed")
	}
}

func TestAllow_WindowExpiry(t *testing.T) {
	const rateLimit = 5
	clientID := t.Name()
	rl := ratelimiter.New(store, rateLimit, 1)

	for i := 0; i < rateLimit; i++ {
		rl.Allow(context.Background(), clientID)
	}
	if rl.Allow(context.Background(), clientID) {
		t.Fatal("expected request at rate limit to be denied before window expiry")
	}

	time.Sleep(2 * time.Second)

	if !rl.Allow(context.Background(), clientID) {
		t.Error("expected request to be allowed after window expiry")
	}
}

func TestAllow_StorageFailure(t *testing.T) {
	unavailableStore, err := storage.New(context.Background(), conf.RedisHostname, conf.RedisPort, conf.RedisPassword)
	if err != nil {
		t.Fatal(err)
	}
	unavailableStore.Close()

	rl := ratelimiter.New(unavailableStore, 10, 60)

	if !rl.Allow(context.Background(), t.Name()) {
		t.Error("expected fail open when storage is unavailable")
	}
}
