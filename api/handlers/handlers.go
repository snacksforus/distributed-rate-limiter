package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/snacksforus/distributed-rate-limiter/api/response"
	"github.com/snacksforus/distributed-rate-limiter/internal/config"
	"github.com/snacksforus/distributed-rate-limiter/internal/middleware"
	"github.com/snacksforus/distributed-rate-limiter/internal/storage"
)

// Register registers the demo API and middleware HTTP handlers.
func Register(store *storage.Storage, config *config.Config) {
	if store == nil || config == nil {
		// A nil store or config is considered a programmer error.
		panic("missing Register parameter")
	}

	mw := middleware.New(store, config.RateLimit, config.WindowSizeSec)

	// Demo API has a single endpoint that just returns success.
	http.Handle("/api", mw.Handler(http.HandlerFunc(handler)))
}

// handler handles HTTP Get requests for the demo API endpoint, returns a JSON
// success message.
func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	resp := response.Success()
	data, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(data)
}
