package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockAllower struct {
	allow bool
}

func (m *mockAllower) Allow(_ context.Context, _ string) bool {
	return m.allow
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
	rl := &RateLimit{
		allower:       &mockAllower{allow: true},
		windowSizeSec: 10,
	}
	next := &nextHandler{}
	handler := rl.Handler(next)

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
	rl := &RateLimit{
		allower:       &mockAllower{allow: false},
		windowSizeSec: 10,
	}
	next := &nextHandler{}
	handler := rl.Handler(next)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = t.Name() + ":12345"
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if next.called {
		t.Error("expected next handler to not be called when request is blocked")
	}
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("expected status 429 got %d", w.Code)
	}
}
