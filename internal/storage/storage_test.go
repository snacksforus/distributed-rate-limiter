// Package storage_test provides tests for the storage package.  The tests are intended to
// be run in a Docker container.  Test clean up is not performed since they are run in an
// isolated Docker environment.
package storage_test

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/snacksforus/distributed-rate-limiter/internal/config"
	"github.com/snacksforus/distributed-rate-limiter/internal/storage"
)

var store *storage.Storage

type noopLogger struct{}

func (noopLogger) Printf(_ context.Context, _ string, _ ...interface{}) {}

func TestMain(m *testing.M) {
	var err error
	var conf *config.Config

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

func TestGetCount_ClientIDNotFound(t *testing.T) {
	clientID := t.Name()
	count, err := store.GetCount(context.Background(), clientID)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Errorf("expected count = 0 when client Id not found, got %v", count)
	}
}

func TestSetAndGet(t *testing.T) {
	clientID := t.Name()
	expectedCount := 7

	err := store.SetCount(context.Background(), clientID, expectedCount)
	if err != nil {
		t.Fatal(err)
	}
	count, err := store.GetCount(context.Background(), clientID)
	if err != nil {
		t.Fatal(err)
	}
	if expectedCount != count {
		t.Errorf("expected count = %v, got %v", expectedCount, count)
	}
}

func TestSetCount_Update(t *testing.T) {
	clientID := t.Name()
	count1 := 100
	count2 := 110

	err := store.SetCount(context.Background(), clientID, count1)
	if err != nil {
		t.Fatal(err)
	}
	err = store.SetCount(context.Background(), clientID, count2)
	if err != nil {
		t.Fatal(err)
	}

	count, err := store.GetCount(context.Background(), clientID)
	if err != nil {
		t.Fatal(err)
	}

	if count != count2 {
		t.Errorf("expected count = %v, got %v", count2, count)
	}
}

func TestIncrWithExpr_NewClientID(t *testing.T) {
	clientID := t.Name()

	count, err := store.IncrWithTTL(context.Background(), clientID, 3)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Errorf("expected count = 1, got %v", count)
	}
}

func TestIncrWithExpr_Increment(t *testing.T) {
	clientID := t.Name()

	for i := 1; i <= 10; i++ {
		count, err := store.IncrWithTTL(context.Background(), clientID, 3)
		if err != nil {
			t.Fatal(err)
		}
		if count != i {
			t.Errorf("expected count = %v, got %v", i, count)
		}
	}
}

func TestIncrWithExpr_Expiration(t *testing.T) {
	clientID := t.Name()

	_, err := store.IncrWithTTL(context.Background(), clientID, 1)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(2 * time.Second)

	count, err := store.GetCount(context.Background(), clientID)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Error("expected count to be reset after expiration")
	}
}

func TestNew_ConnectionFailure(t *testing.T) {
	_, err := storage.New(context.Background(), "localhost", 6555, "")
	if err == nil {
		t.Error("expected error connecting to invalid database")
	}
}
