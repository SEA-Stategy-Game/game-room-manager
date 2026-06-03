package rooms

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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
			RoomID: "abc",
			State:  StateIniting,
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
	if rooms[0].State != StateIniting {
		t.Fatalf("expected state %q, got %q", StateIniting, rooms[0].State)
	}
}

func TestJoinRoom_AddsPlayerToRoom(t *testing.T) {
	t.Parallel()

	repo := NewInMemoryRepository([]Room{
		{
			RoomID:  "room-123",
			State:   StateIniting,
			Players: []string{"player-1"},
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
}

func TestJoinRoom_FullRoom(t *testing.T) {
	t.Parallel()

	maxPlayers := 2
	repo := NewInMemoryRepository([]Room{
		{
			RoomID:             "room-full",
			State:              StateIniting,
			Players:            []string{"player-1", "player-2"},
			MaxNumberOfPlayers: &maxPlayers,
		},
	})
	svc := NewService(repo, "test-game-image:latest")
	h := NewHandler(svc, zap.NewNop())

	r := chi.NewRouter()
	r.Post("/rooms/{roomId}/players/{playerId}/join", h.JoinRoom)

	req := httptest.NewRequest(http.MethodPost, "/rooms/room-full/players/player-3/join", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, rec.Code)
	}

	expectedBody := "room is full\n"
	if rec.Body.String() != expectedBody {
		t.Fatalf("expected body %q, got %q", expectedBody, rec.Body.String())
	}

	updated, _ := repo.GetByID(req.Context(), "room-full")
	if len(updated.Players) != 2 {
		t.Fatalf("expected 2 players, got %d", len(updated.Players))
	}
}

func TestJoinRoom_NoLimit(t *testing.T) {
	t.Parallel()

	repo := NewInMemoryRepository([]Room{
		{
			RoomID:             "room-no-limit",
			State:              StateIniting,
			Players:            []string{"player-1", "player-2"},
			MaxNumberOfPlayers: nil, // No limit
		},
	})
	svc := NewService(repo, "test-game-image:latest")
	h := NewHandler(svc, zap.NewNop())

	r := chi.NewRouter()
	r.Post("/rooms/{roomId}/players/{playerId}/join", h.JoinRoom)

	req := httptest.NewRequest(http.MethodPost, "/rooms/room-no-limit/players/player-3/join", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestFindPlayerStatus(t *testing.T) {
	t.Parallel()

	repo := NewInMemoryRepository([]Room{
		{
			RoomID:  "room-123",
			State:   StateIniting,
			Players: []string{"player-1"},
		},
		{
			RoomID:  "room-345",
			State:   StateIniting,
			Players: []string{"player-2"},
		},
		{
			RoomID:  "room-133",
			State:   StateIniting,
			Players: []string{"player-2", "player-3"},
		},
		{
			RoomID:  "room-163",
			State:   StateEnded,
			Players: []string{"player-1", "player-3"},
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
	req = httptest.NewRequest(http.MethodGet, "/rooms?status=initing", nil)
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
		t.Fatalf("expected 3 filtered initing rooms, got %d", len(filteredRooms2))
	}
	req = httptest.NewRequest(http.MethodGet, "/rooms?player=player-3&status=ended", nil)
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
		t.Fatalf("expected 1 filtered ended rooms, got %d", len(filteredRooms3))
	}
}

func TestSetGameStatus_AllStates(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name           string
		roomID         string
		statusParam    string
		winnerParam    *string
		initialRoom    *Room
		expectedStatus int
		expectedState  State
		checkFn        func(t *testing.T, updated *Room)
	}

	strPtr := func(s string) *string { return &s }

	tests := []testCase{
		{
			name:           "Return Not Found if room doesn't exist",
			roomID:         "room-nonexistent",
			statusParam:    "running",
			winnerParam:    nil,
			initialRoom:    nil,
			expectedStatus: http.StatusNotFound,
			expectedState:  "",
		},
		{
			name:        "Transition to initing updates StartedAt time",
			roomID:      "room-initing",
			statusParam: "init",
			winnerParam: nil,
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
			winnerParam: nil,
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
			winnerParam: strPtr("player-1"),
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
			winnerParam: nil,
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
			r.Post("/rooms/{roomId}/status", h.SetStatus)

			bodyPayload := map[string]interface{}{
				"status": tc.statusParam,
				"winner": tc.winnerParam,
			}

			jsonBody, err := json.Marshal(bodyPayload)
			if err != nil {
				t.Fatalf("failed to marshal request body payload: %v", err)
			}

			url := fmt.Sprintf("/rooms/%s/status", tc.roomID)

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
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
			RoomID:  "room-123",
			State:   StateIniting,
			Players: []string{"player-1"},
		},
		{
			RoomID:  "room-345",
			State:   StateIniting,
			Players: []string{"player-2"},
		},
		{
			RoomID:  "room-133",
			State:   StateIniting,
			Players: []string{"player-2", "player-3"},
		},
		{
			RoomID:  "room-163",
			State:   StateEnded,
			Players: []string{"player-1", "player-3"},
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

func TestCreateGame(t *testing.T) {
	t.Parallel()

	intPtr := func(i int) *int { return &i }

	type testCase struct {
		name              string
		payload           interface{}
		expectedStatus    int
		expectedMaxPlayer *int
	}

	tests := []testCase{
		{
			name: "successful creation with max players",
			payload: CreateGameRequest{
				MaxNumberOfPlayers: intPtr(2),
			},
			expectedStatus:    http.StatusCreated,
			expectedMaxPlayer: intPtr(2),
		},
		{
			name:              "successful creation without max players (nil payload)",
			payload:           nil,
			expectedStatus:    http.StatusCreated,
			expectedMaxPlayer: intPtr(32),
		},
		{
			name:              "successful creation with empty request body",
			payload:           CreateGameRequest{},
			expectedStatus:    http.StatusCreated,
			expectedMaxPlayer: intPtr(32),
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := NewInMemoryRepository([]Room{})
			svc := NewService(repo, "test-game-image:latest")
			h := NewHandler(svc, zap.NewNop())

			r := chi.NewRouter()
			r.Post("/rooms/create", h.CreateGame)

			var reqBody io.Reader
			if tc.payload != nil {
				body, err := json.Marshal(tc.payload)
				if err != nil {
					t.Fatalf("failed to marshal payload: %v", err)
				}
				reqBody = bytes.NewBuffer(body)
			}

			req := httptest.NewRequest(http.MethodPost, "/rooms/create", reqBody)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			if rec.Code != tc.expectedStatus {
				t.Fatalf("expected status %d, got %d. Body: %s", tc.expectedStatus, rec.Code, rec.Body.String())
			}

			var room Room
			if err := json.Unmarshal(rec.Body.Bytes(), &room); err != nil {
				t.Fatalf("failed to unmarshal response body: %v", err)
			}

			if tc.expectedMaxPlayer == nil {
				if room.MaxNumberOfPlayers != nil {
					t.Errorf("expected max players to be nil, got %d", *room.MaxNumberOfPlayers)
				}
			} else {
				if room.MaxNumberOfPlayers == nil {
					t.Errorf("expected max players to be %d, got nil", *tc.expectedMaxPlayer)
				} else if *room.MaxNumberOfPlayers != *tc.expectedMaxPlayer {
					t.Errorf("expected max players to be %d, got %d", *tc.expectedMaxPlayer, *room.MaxNumberOfPlayers)
				}
			}

			_, err := repo.GetByID(context.Background(), room.RoomID)
			if err != nil {
				t.Fatalf("failed to get room from repo: %v", err)
			}
		})
	}
}

func TestRegisterManualGame(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name           string
		initialRooms   []Room
		payload        interface{}
		expectedStatus int
		checkFn        func(t *testing.T, repo *InMemoryRepository, body *bytes.Buffer)
	}

	intPtr := func(i int) *int { return &i }

	tests := []testCase{
		{
			name: "successful creation with max players",
			payload: RegisterManualGameRequest{
				RoomID:             "manual-room-1",
				Address:            "localhost",
				Port:               9876,
				MaxNumberOfPlayers: intPtr(8),
			},
			expectedStatus: http.StatusCreated,
			checkFn: func(t *testing.T, repo *InMemoryRepository, body *bytes.Buffer) {
				var room Room
				if err := json.Unmarshal(body.Bytes(), &room); err != nil {
					t.Fatalf("failed to unmarshal response body: %v", err)
				}
				if room.RoomID != "manual-room-1" {
					t.Errorf("expected roomID %q, got %q", "manual-room-1", room.RoomID)
				}
				if room.State != StateReady {
					t.Errorf("expected state %q, got %q", StateReady, room.State)
				}
				if room.MaxNumberOfPlayers == nil || *room.MaxNumberOfPlayers != 8 {
					t.Errorf("expected maxNumberOfPlayers to be 8, got %v", room.MaxNumberOfPlayers)
				}

				created, err := repo.GetByID(context.Background(), "manual-room-1")
				if err != nil {
					t.Fatalf("failed to get room from repo: %v", err)
				}
				if created == nil {
					t.Fatal("room was not created in repository")
				}
				if created.RoomID != "manual-room-1" {
					t.Errorf("repo: expected roomID %q, got %q", "manual-room-1", created.RoomID)
				}
				if created.MaxNumberOfPlayers == nil || *created.MaxNumberOfPlayers != 8 {
					t.Errorf("repo: expected maxNumberOfPlayers to be 8, got %v", created.MaxNumberOfPlayers)
				}
			},
		},
		{
			name: "successful creation without max players",
			payload: RegisterManualGameRequest{
				RoomID:  "manual-room-2",
				Address: "localhost",
				Port:    9877,
			},
			expectedStatus: http.StatusCreated,
			checkFn: func(t *testing.T, repo *InMemoryRepository, body *bytes.Buffer) {
				var room Room
				if err := json.Unmarshal(body.Bytes(), &room); err != nil {
					t.Fatalf("failed to unmarshal response body: %v", err)
				}

				if room.MaxNumberOfPlayers == nil || *room.MaxNumberOfPlayers != 32 {
					t.Errorf("expected maxNumberOfPlayers to be 32, got %v", room.MaxNumberOfPlayers)
				}

				created, err := repo.GetByID(context.Background(), "manual-room-2")
				if err != nil {
					t.Fatalf("failed to get room from repo: %v", err)
				}
				if created.MaxNumberOfPlayers == nil || *created.MaxNumberOfPlayers != 32 {
					t.Errorf("repo: expected maxNumberOfPlayers to be 32, got %v", created.MaxNumberOfPlayers)
				}
			},
		},
		{
			name:           "missing roomId",
			payload:        map[string]interface{}{"address": "localhost", "port": 9876},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing address",
			payload:        map[string]interface{}{"roomId": "manual-room-1", "port": 9876},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing port",
			payload:        map[string]interface{}{"roomId": "manual-room-1", "address": "localhost"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "duplicate roomId overrides existing room",
			initialRooms: []Room{
				{RoomID: "manual-room-1", Address: "old-address", Port: 1111, Players: []string{"p1"}},
			},
			payload: RegisterManualGameRequest{
				RoomID:  "manual-room-1",
				Address: "new-address",
				Port:    9876,
			},
			expectedStatus: http.StatusCreated,
			checkFn: func(t *testing.T, repo *InMemoryRepository, body *bytes.Buffer) {
				updated, err := repo.GetByID(context.Background(), "manual-room-1")
				if err != nil {
					t.Fatalf("failed to get room from repo: %v", err)
				}
				if updated.Address != "new-address" {
					t.Errorf("expected address to be 'new-address', got %q", updated.Address)
				}
				if updated.Port != 9876 {
					t.Errorf("expected port to be 9876, got %d", updated.Port)
				}
				if len(updated.Players) != 0 {
					t.Errorf("expected players to be reset to empty, got %d", len(updated.Players))
				}
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := NewInMemoryRepository(tc.initialRooms)
			svc := NewService(repo, "test-game-image:latest")
			h := NewHandler(svc, zap.NewNop())

			r := chi.NewRouter()
			r.Post("/rooms", h.RegisterManualGame)

			body, err := json.Marshal(tc.payload)
			if err != nil {
				t.Fatalf("failed to marshal payload: %v", err)
			}

			req := httptest.NewRequest(http.MethodPost, "/rooms", bytes.NewBuffer(body))
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			if rec.Code != tc.expectedStatus {
				t.Fatalf("expected status %d, got %d. Body: %s", tc.expectedStatus, rec.Code, rec.Body.String())
			}

			if tc.checkFn != nil {
				tc.checkFn(t, repo, rec.Body)
			}
		})
	}
}
