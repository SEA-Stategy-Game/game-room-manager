package httpserver

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SEA-Stategy-Game/game-room-manager/internal/config"
	"go.uber.org/zap"
)

func TestHealthz(t *testing.T) {
	t.Parallel()

	// Note: This test may now fail if the test runner doesn't have permissions
	// to create 'manager.db' in its working directory.
	srv, err := New(&config.Config{Port: 8080}, zap.NewNop())
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	srv.server.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if rec.Body.String() != "ok" {
		t.Fatalf("expected body %q, got %q", "ok", rec.Body.String())
	}
}
