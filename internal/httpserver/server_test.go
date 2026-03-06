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

	srv := New(&config.Config{Port: 8080}, zap.NewNop())

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

