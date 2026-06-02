package rooms

import (
	"encoding/json"
	"fmt"
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
	svc := NewService(repo, "test-game-image:latest")
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
	svc := NewService(repo, "test-game-image:latest")
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

func TestFindPlayerStatus(t *testing.T) {
	t.Parallel()

	repo := NewInMemoryRepository([]Room{
		{
			RoomID:            "room-123",
			ConnectionDetails: "ws://example/rooms/room-123",
			State:             StateActive,
			Participants:      1,
			Players:           []string{"player-1"},
		},
		{
			RoomID:            "room-345",
			ConnectionDetails: "ws://example/rooms/room-345",
			State:             StateActive,
			Participants:      1,
			Players:           []string{"player-2"},
		},
		{
			RoomID:            "room-133",
			ConnectionDetails: "ws://example/rooms/room-133",
			State:             StateActive,
			Participants:      1,
			Players:           []string{"player-2", "player-3"},
		},
		{
			RoomID:            "room-163",
			ConnectionDetails: "ws://example/rooms/room-163",
			State:             StateInactive,
			Participants:      1,
			Players:           []string{"player-1", "player-3"},
		},
	})
	svc := NewService(repo, "test-game-image:latest")
	h := NewHandler(svc, zap.NewNop())

	r := chi.NewRouter()
	r.Get("/rooms", h.GetRooms)

	req := httptest.NewRequest(http.MethodGet, "/rooms?player=player-1", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var filteredRooms1 []Room
	if err := json.Unmarshal(rec.Body.Bytes(), &filteredRooms1); err != nil {
		t.Fatalf("failed to parse response JSON: %v", err)
	}
	if len(filteredRooms1) != 2 {
		t.Fatalf("expected 2 filtered rooms for player-1, got %d", len(filteredRooms1))
	}
	req = httptest.NewRequest(http.MethodGet, "/rooms?status=active", nil)
	rec = httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var filteredRooms2 []Room
	if err := json.Unmarshal(rec.Body.Bytes(), &filteredRooms2); err != nil {
		t.Fatalf("failed to parse response JSON: %v", err)
	}

	if len(filteredRooms2) != 3 {
		t.Fatalf("expected 3 filtered active rooms, got %d", len(filteredRooms2))
	}
	req = httptest.NewRequest(http.MethodGet, "/rooms?player=player-3&status=inactive", nil)
	rec = httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var filteredRooms3 []Room
	if err := json.Unmarshal(rec.Body.Bytes(), &filteredRooms3); err != nil {
		t.Fatalf("failed to parse response JSON: %v", err)
	}
	if len(filteredRooms3) != 1 {
		t.Fatalf("expected 1 filtered active rooms, got %d", len(filteredRooms3))
	}
}

// func TestReady(t *testing.T) {
// 	t.Parallel()

// 	repo := NewInMemoryRepository([]Room{
// 		{
// 			RoomID:            "room-123",
// 			ConnectionDetails: "ws://example/rooms/room-123",
// 			State:             StateActive,
// 			Participants:      1,
// 			Players:           []string{"player-1"},
// 		},
// 		{
// 			RoomID:            "room-345",
// 			ConnectionDetails: "ws://example/rooms/room-345",
// 			State:             StateActive,
// 			Participants:      1,
// 			Players:           []string{"player-2"},
// 		},
// 		{
// 			RoomID:            "room-133",
// 			ConnectionDetails: "ws://example/rooms/room-133",
// 			State:             StateActive,
// 			Participants:      1,
// 			Players:           []string{"player-2", "player-3"},
// 		},
// 		{
// 			RoomID:            "room-163",
// 			ConnectionDetails: "ws://example/rooms/room-163",
// 			State:             StateInactive,
// 			Participants:      1,
// 			Players:           []string{"player-1", "player-3"},
// 		},
// 	})
// 	svc := NewService(repo, "test-game-image:latest")
// 	h := NewHandler(svc, zap.NewNop())

// 	r := chi.NewRouter()
// 	r.Post("/rooms/{roomId}/ready", h.SetReady)

// 	req := httptest.NewRequest(http.MethodPost, "/rooms/room-163/ready", nil)
// 	rec := httptest.NewRecorder()

// 	r.ServeHTTP(rec, req)

// 	if rec.Code != http.StatusOK {
// 		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
// 	}

// 	updated, _ := repo.GetByID(req.Context(), "room-163")

// 	if updated.State != StateReady {
// 		t.Fatalf("expected state %q, got %q", StateReady, updated.State)
// 	}
// }

// func TestCrash(t *testing.T) {
// 	t.Parallel()

// 	repo := NewInMemoryRepository([]Room{
// 		{
// 			RoomID:            "room-123",
// 			ConnectionDetails: "ws://example/rooms/room-123",
// 			State:             StateActive,
// 			Participants:      1,
// 			Players:           []string{"player-1"},
// 		},
// 		{
// 			RoomID:            "room-345",
// 			ConnectionDetails: "ws://example/rooms/room-345",
// 			State:             StateActive,
// 			Participants:      1,
// 			Players:           []string{"player-2"},
// 		},
// 		{
// 			RoomID:            "room-133",
// 			ConnectionDetails: "ws://example/rooms/room-133",
// 			State:             StateActive,
// 			Participants:      1,
// 			Players:           []string{"player-2", "player-3"},
// 		},
// 		{
// 			RoomID:            "room-163",
// 			ConnectionDetails: "ws://example/rooms/room-163",
// 			State:             StateInactive,
// 			Participants:      1,
// 			Players:           []string{"player-1", "player-3"},
// 		},
// 	})
// 	svc := NewService(repo, "test-game-image:latest")
// 	h := NewHandler(svc, zap.NewNop())

// 	r := chi.NewRouter()
// 	r.Post("/rooms/{roomId}/crash", h.SetCrashed)

// 	req := httptest.NewRequest(http.MethodPost, "/rooms/room-163/crash", nil)
// 	rec := httptest.NewRecorder()

// 	r.ServeHTTP(rec, req)

// 	if rec.Code != http.StatusOK {
// 		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
// 	}

// 	updated, _ := repo.GetByID(req.Context(), "room-163")

// 	if updated.State != StateCrashed {
// 		t.Fatalf("expected state %q, got %q", StateCrashed, updated.State)
// 	}
// }

func TestSetGameStatus_AllStates(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name           string
		roomID         string
		statusParam    string
		winnerParam    string
		initialRoom    *Room
		expectedStatus int
		expectedState  State
		checkFn        func(t *testing.T, updated *Room)
	}

	tests := []testCase{
		{
			name:           "Create new room when status is ready and room doesn't exist",
			roomID:         "room-new-ready",
			statusParam:    "ready",
			initialRoom:    nil,
			expectedStatus: http.StatusOK,
			expectedState:  StateReady,
		},
		{
			name:           "Do nothing and return OK if room doesn't exist and status is not ready",
			roomID:         "room-nonexistent",
			statusParam:    "running",
			initialRoom:    nil,
			expectedStatus: http.StatusOK,
			expectedState:  "",
		},
		{
			name:        "Transition to initing updates StartedAt time",
			roomID:      "room-initing",
			statusParam: "init",
			initialRoom: &Room{
				RoomID: "room-initing",
				State:  StateReady,
			},
			expectedStatus: http.StatusOK,
			expectedState:  "init",
			checkFn: func(t *testing.T, updated *Room) {
				if updated.StartedAt.IsZero() {
					t.Error("expected StartedAt to be populated, got zero time")
				}
			},
		},
		{
			name:        "Transition to running",
			roomID:      "room-running",
			statusParam: "running",
			initialRoom: &Room{
				RoomID: "room-running",
				State:  "init",
			},
			expectedStatus: http.StatusOK,
			expectedState:  StateRunning,
		},
		{
			name:        "Transition to ended sets winner and EndedAt time",
			roomID:      "room-ended",
			statusParam: "ended",
			winnerParam: "player-1",
			initialRoom: &Room{
				RoomID: "room-ended",
				State:  StateRunning,
			},
			expectedStatus: http.StatusOK,
			expectedState:  StateEnded,
			checkFn: func(t *testing.T, updated *Room) {
				if updated.Winner != "player-1" {
					t.Errorf("expected winner 'player-1', got %q", updated.Winner)
				}
				if updated.EndedAt.IsZero() {
					t.Error("expected EndedAt to be populated, got zero time")
				}
			},
		},
		{
			name:        "Transition to crashed sets EndedAt time",
			roomID:      "room-crashed",
			statusParam: "crashed",
			initialRoom: &Room{
				RoomID: "room-crashed",
				State:  StateRunning,
			},
			expectedStatus: http.StatusOK,
			expectedState:  StateCrashed,
			checkFn: func(t *testing.T, updated *Room) {
				if updated.EndedAt.IsZero() {
					t.Error("expected EndedAt to be populated, got zero time")
				}
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var initialRooms []Room
			if tc.initialRoom != nil {
				initialRooms = append(initialRooms, *tc.initialRoom)
			}
			repo := NewInMemoryRepository(initialRooms)
			svc := NewService(repo, "test-game-image:latest")
			h := NewHandler(svc, zap.NewNop())

			r := chi.NewRouter()
			r.Post("/rooms/{roomId}/status/{status}", h.SetStatus)
			r.Post("/rooms/{roomId}/status/{status}/{winnerId}", h.SetStatus)

			url := fmt.Sprintf("/rooms/%s/status/%s", tc.roomID, tc.statusParam)
			if tc.winnerParam != "" {
				url = fmt.Sprintf("%s/%s", url, tc.winnerParam)
			}

			req := httptest.NewRequest(http.MethodPost, url, nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			if rec.Code != tc.expectedStatus {
				t.Fatalf("expected status %d, got %d", tc.expectedStatus, rec.Code)
			}

			if tc.expectedState == "" && tc.initialRoom == nil {
				return
			}

			updated, err := repo.GetByID(req.Context(), tc.roomID)
			if err != nil {
				t.Fatalf("failed to fetch updated room from repo: %v", err)
			}

			if updated == nil {
				if tc.expectedState != "" {
					t.Fatalf("expected room to exist with state %q, but got nil", tc.expectedState)
				}
				return
			}

			if updated.State != tc.expectedState {
				t.Errorf("expected state %q, got %q", tc.expectedState, updated.State)
			}

			if tc.checkFn != nil {
				tc.checkFn(t, updated)
			}
		})
	}
}

func TestFindRoom(t *testing.T) {
	t.Parallel()

	repo := NewInMemoryRepository([]Room{
		{
			RoomID:            "room-123",
			ConnectionDetails: "ws://example/rooms/room-123",
			State:             StateActive,
			Participants:      1,
			Players:           []string{"player-1"},
		},
		{
			RoomID:            "room-345",
			ConnectionDetails: "ws://example/rooms/room-345",
			State:             StateActive,
			Participants:      1,
			Players:           []string{"player-2"},
		},
		{
			RoomID:            "room-133",
			ConnectionDetails: "ws://example/rooms/room-133",
			State:             StateActive,
			Participants:      1,
			Players:           []string{"player-2", "player-3"},
		},
		{
			RoomID:            "room-163",
			ConnectionDetails: "ws://example/rooms/room-163",
			State:             StateInactive,
			Participants:      1,
			Players:           []string{"player-1", "player-3"},
		},
	})
	svc := NewService(repo, "test-game-image:latest")
	h := NewHandler(svc, zap.NewNop())

	r := chi.NewRouter()
	r.Get("/room/{roomId}", h.GetRoom)

	req := httptest.NewRequest(http.MethodGet, "/room/room-133", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var room Room
	if err := json.Unmarshal(rec.Body.Bytes(), &room); err != nil {
		t.Fatalf("failed to parse response JSON: %v", err)
	}
	if room.RoomID != "room-133" {
		t.Fatalf("expected name %s, got %s", "room-133", room.RoomID)
	}
}
