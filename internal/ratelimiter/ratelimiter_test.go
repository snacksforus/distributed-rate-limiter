package ratelimiter_test

import (
	"context"
	"errors"
	"testing"

	"github.com/snacksforus/distributed-rate-limiter/internal/ratelimiter"
)

type mockCounter struct {
	counts map[string]int
	err    error
}

func newMockCounter() *mockCounter {
	return &mockCounter{
		counts: make(map[string]int),
	}
}

func (m *mockCounter) IncrWithTTL(_ context.Context, clientID string, _ int) (int, error) {
	if m.err != nil {
		return 0, m.err
	}
	m.counts[clientID]++
	return m.counts[clientID], nil
}

func TestAllow_BelowLimit(t *testing.T) {
	const rateLimit = 5
	clientID := t.Name()
	mc := newMockCounter()
	rl := ratelimiter.New(mc, rateLimit, 60)

	for i := 0; i < rateLimit-1; i++ {
		if !rl.Allow(context.Background(), clientID) {
			t.Errorf("expected request %d to be allowed", i)
		}
	}
}

func TestAllow_AtLimit(t *testing.T) {
	const rateLimit = 5
	clientID := t.Name()
	mc := newMockCounter()
	rl := ratelimiter.New(mc, rateLimit, 60)

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
	mc := newMockCounter()
	rl := ratelimiter.New(mc, rateLimit, 60)

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
	mc := newMockCounter()
	rl := ratelimiter.New(mc, rateLimit, 60)

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
	mc := newMockCounter()
	rl := ratelimiter.New(mc, rateLimit, 1)

	for i := 0; i < rateLimit; i++ {
		rl.Allow(context.Background(), clientID)
	}
	if rl.Allow(context.Background(), clientID) {
		t.Fatal("expected request at rate limit to be denied before window expiry")
	}

	// Simulate window expiry by resetting the counter.
	mc.counts[clientID] = 0

	if !rl.Allow(context.Background(), clientID) {
		t.Error("expected request to be allowed after window expiry")
	}
}

func TestAllow_StorageFailure(t *testing.T) {
	mc := newMockCounter()
	mc.err = errors.New("connection refused")
	rl := ratelimiter.New(mc, 10, 60)

	if !rl.Allow(context.Background(), t.Name()) {
		t.Error("expected fail open when storage is unavailable")
	}
}
