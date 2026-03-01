// Package middleware_test provides tests for the middleware package.  The tests are intended
// to be run in a Docker container.  Test clean up is not performed since they are run in an
// isolated Docker environment.
package middleware_test

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/snacksforus/distributed-rate-limiter/internal/config"
	"github.com/snacksforus/distributed-rate-limiter/internal/middleware"
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

// nextHandler records whether ServeHTTP was called.
type nextHandler struct {
	called bool
}

func (h *nextHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	h.called = true
	w.WriteHeader(http.StatusOK)
}

func TestRateLimit_AllowsRequest(t *testing.T) {
	m := middleware.New(store, conf.RateLimit, conf.WindowSizeSec)
	next := &nextHandler{}
	handler := m.Handler(next)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = t.Name() + ":12345"
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if !next.called {
		t.Error("expected next handler to be called when request is allowed")
	}
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200 got %d", w.Code)
	}
}

func TestRateLimit_BlocksRequest(t *testing.T) {
	m := middleware.New(store, conf.RateLimit, conf.WindowSizeSec)
	next := &nextHandler{}
	handler := m.Handler(next)

	remoteAddr := t.Name() + ":12345"

	for i := 0; i < conf.RateLimit; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = remoteAddr
		handler.ServeHTTP(httptest.NewRecorder(), req)
	}

	// Next request should be blocked.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = remoteAddr
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("expected status 429 got %d", w.Code)
	}
}
