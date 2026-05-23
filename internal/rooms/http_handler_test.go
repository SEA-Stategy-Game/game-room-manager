package rooms

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func TestGetRooms_ReturnsJSONList(t *testing.T) {
	t.Parallel()

	repo := NewInMemoryRepository([]Room{
		{
			RoomID:            "abc",
			ConnectionDetails: "ws://example/rooms/abc",
			State:             StateActive,
			Participants:      2,
		},
	})
	svc := NewService(repo)
	h := NewHandler(svc, zap.NewNop())

	r := chi.NewRouter()
	r.Get("/rooms", h.GetRooms)

	req := httptest.NewRequest(http.MethodGet, "/rooms", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected Content-Type %q, got %q", "application/json", got)
	}

	var rooms []Room
	if err := json.Unmarshal(rec.Body.Bytes(), &rooms); err != nil {
		t.Fatalf("failed to parse response JSON: %v", err)
	}
	if len(rooms) != 1 {
		t.Fatalf("expected 1 room, got %d", len(rooms))
	}
	if rooms[0].RoomID != "abc" {
		t.Fatalf("expected roomId %q, got %q", "abc", rooms[0].RoomID)
	}
	if rooms[0].ConnectionDetails != "ws://example/rooms/abc" {
		t.Fatalf("expected connectionDetails %q, got %q", "ws://example/rooms/abc", rooms[0].ConnectionDetails)
	}
	if rooms[0].State != StateActive {
		t.Fatalf("expected state %q, got %q", StateActive, rooms[0].State)
	}
	if rooms[0].Participants != 2 {
		t.Fatalf("expected participants %d, got %d", 2, rooms[0].Participants)
	}
}

func TestJoinRoom_AddsPlayerToRoom(t *testing.T) {
	t.Parallel()

	repo := NewInMemoryRepository([]Room{
		{
			RoomID:            "room-123",
			ConnectionDetails: "ws://example/rooms/room-123",
			State:             StateActive,
			Participants:      1,
			Players:           []string{"player-1"},
		},
	})
	svc := NewService(repo)
	h := NewHandler(svc, zap.NewNop())

	r := chi.NewRouter()
	r.Post("/rooms/{roomId}/players/{playerId}/join", h.JoinRoom)

	req := httptest.NewRequest(http.MethodPost, "/rooms/room-123/players/player-2/join", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	updated, _ := repo.GetByID(req.Context(), "room-123")
	if len(updated.Players) != 2 {
		t.Fatalf("expected 2 players, got %d", len(updated.Players))
	}
	if updated.Players[1] != "player-2" {
		t.Fatalf("expected player-2, got %q", updated.Players[1])
	}
	if updated.Participants != 2 {
		t.Fatalf("expected participants 2, got %d", updated.Participants)
	}
}

