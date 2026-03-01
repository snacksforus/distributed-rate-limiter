//go:build integration

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/snacksforus/distributed-rate-limiter/api/handlers"
	"github.com/snacksforus/distributed-rate-limiter/api/response"
	"github.com/snacksforus/distributed-rate-limiter/internal/config"
	"github.com/snacksforus/distributed-rate-limiter/internal/storage"
)

var conf *config.Config

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

	// Verify Redis is reachable before running tests.
	store, err := storage.New(context.Background(), conf.RedisHostname, conf.RedisPort, conf.RedisPassword)
	if err != nil {
		slog.Error("failed to connect to Redis", "error", err)
		os.Exit(1)
	}
	store.Close()

	os.Exit(m.Run())
}

// startServer creates a server with the given rate limit bound to an ephemeral
// port and returns the base URL and a cleanup function.
func startServer(t *testing.T, rateLimit int) string {
	t.Helper()

	store, err := storage.New(context.Background(), conf.RedisHostname, conf.RedisPort, conf.RedisPassword)
	if err != nil {
		t.Fatalf("failed to connect to Redis: %v", err)
	}

	if err := store.FlushAll(context.Background()); err != nil {
		t.Fatalf("failed to flush Redis: %v", err)
	}

	cfg := &config.Config{
		Hostname:            "",
		Port:                0,
		RateLimit:           rateLimit,
		WindowSizeSec:       60,
		TimeoutMS:           5000,
		ReadHeaderTimeoutMS: 5000,
		ReadTimeoutMS:       5000,
	}

	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("failed to listen on ephemeral port: %v", err)
	}

	server := handlers.NewServer(store, cfg)

	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
		}
	}()

	t.Cleanup(func() {
		server.Shutdown(context.Background())
		store.Close()
	})

	return fmt.Sprintf("http://localhost:%d", listener.Addr().(*net.TCPAddr).Port)
}

func TestIntegration_AllowedRequest(t *testing.T) {
	baseURL := startServer(t, 10)

	resp, err := http.Get(baseURL + "/api")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var body response.Response
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !body.Success {
		t.Errorf("expected success to be true")
	}
	if body.Error != nil {
		t.Errorf("expected error to be nil, got %+v", body.Error)
	}
}

func TestIntegration_RateLimitExceeded(t *testing.T) {
	const rateLimit = 3
	baseURL := startServer(t, rateLimit)

	// Send rateLimit requests; all should succeed.
	for i := range rateLimit {
		resp, err := http.Get(baseURL + "/api")
		if err != nil {
			t.Fatalf("request %d: unexpected error: %v", i+1, err)
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("request %d: expected status 200, got %d", i+1, resp.StatusCode)
		}
	}

	// The next request should be rate limited.
	resp, err := http.Get(baseURL + "/api")
	if err != nil {
		t.Fatalf("rate-limited request: unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("expected status 429, got %d", resp.StatusCode)
	}

	retryAfter := resp.Header.Get("Retry-After")
	if retryAfter != "60" {
		t.Errorf("expected Retry-After header to be 60, got %q", retryAfter)
	}

	var body response.Response
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body.Success {
		t.Errorf("expected success to be false")
	}
	if body.Error == nil {
		t.Fatal("expected error to be non-nil")
	}
	if body.Error.Code != "TOO_MANY_REQUESTS" {
		t.Errorf("expected error code TOO_MANY_REQUESTS, got %q", body.Error.Code)
	}
	if body.Error.Message != "Exceeded request rate limit" {
		t.Errorf("expected error message %q, got %q", "Exceeded request rate limit", body.Error.Message)
	}
}

func TestIntegration_MethodNotAllowed(t *testing.T) {
	baseURL := startServer(t, 10)

	resp, err := http.Post(baseURL+"/api", "application/json", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405, got %d", resp.StatusCode)
	}
}
