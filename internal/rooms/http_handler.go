package rooms

import (
	"database/sql"
	"encoding/json"
	"database/sql"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// Handler is the inbound adapter for HTTP.
type Handler struct {
	svc *Service
	log *zap.Logger
}

func NewHandler(svc *Service, log *zap.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

// Helper functions for querying.
func findPlayer(rooms []Room, searchPlayer string) []Room {
	var found []Room

	for _, room := range rooms {
		for _, player := range room.Players {
			if player == searchPlayer {
				found = append(found, room)
				break
			}
		}
	}
	return found
}

func findByStatus(rooms []Room, searchStat string) []Room {
	var found []Room

	targetState := State(searchStat)

	for _, room := range rooms {
		if room.State == targetState {
			found = append(found, room)
		}
	}
	return found
}

func findDouble(rooms []Room, searchPlayer string, searchStat string) []Room {
	var found []Room

	targetState := State(searchStat)

	for _, room := range rooms {
		if room.State == targetState {
			for _, player := range room.Players {
				if player == searchPlayer {
					found = append(found, room)
					break
				}
			}
		}
	}
	return found
}

func (h *Handler) GetRooms(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	rooms, err := h.svc.ListRooms(r.Context())

	if err != nil {
		h.log.Error("failed to list rooms", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Check for a query, if not, just show all.
	player := query.Get("player")
	status := query.Get("status")
	if player != "" && status == "" {
		rooms = findPlayer(rooms, player)
	}

	if status != "" && player == "" {
		rooms = findByStatus(rooms, status)
	}

	if player != "" && status != "" {
		rooms = findDouble(rooms, player, status)
	}

	// h.log.Info("retrieved rooms", zap.Any("rooms", rooms))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(rooms); err != nil {
		h.log.Error("failed to encode response", zap.Error(err))
	}
}

func (h *Handler) GetRoom(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "roomId")

	room, err := h.svc.FindRoom(r.Context(), roomID)

	if err != nil {
		// The service layer passes up repository errors.
		// We check if it's a known "not found" error.
		if errors.Is(err, ErrRoomNotFound) || errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "room not found", http.StatusNotFound)
			return
		}
		h.log.Error("failed to find room", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(room); err != nil {
		h.log.Error("failed to encode response", zap.Error(err))
	}
}

func (h *Handler) JoinRoom(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "roomId")
	playerID := chi.URLParam(r, "playerId")

	err := h.svc.JoinGameRoom(r.Context(), roomID, playerID)
	if err != nil {
		if errors.Is(err, ErrRoomFull) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		h.log.Error("failed to join room", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("player joined"))
}

type CreateGameRequest struct {
	MaxNumberOfPlayers *int `json:"maxNumberOfPlayers,omitempty"`
}

func (h *Handler) CreateGame(w http.ResponseWriter, r *http.Request) {
	var req CreateGameRequest
	if r.Body != http.NoBody {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.log.Error("failed to decode request body", zap.Error(err))
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()
	}

	room, err := h.svc.RegisterGameRoom(r.Context(), req.MaxNumberOfPlayers)
	if err != nil {
		h.log.Error("failed to create game", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(room); err != nil {
		h.log.Error("failed to encode room response", zap.Error(err))
	}
}

type RegisterManualGameRequest struct {
	RoomID             string `json:"roomId"`
	Address            string `json:"address"`
	Port               int    `json:"port"`
	MaxNumberOfPlayers *int   `json:"maxNumberOfPlayers,omitempty"`
}

// RegisterManualGame is used for the local test gaming room that is manually created.
func (h *Handler) RegisterManualGame(w http.ResponseWriter, r *http.Request) {
	var req RegisterManualGameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error("failed to decode request body", zap.Error(err))
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.RoomID == "" || req.Address == "" || req.Port == 0 {
		http.Error(w, "roomId, address and port are required", http.StatusBadRequest)
		return
	}

	room, err := h.svc.RegisterManualGame(r.Context(), req.RoomID, req.Address, req.Port, req.MaxNumberOfPlayers)
	if err != nil {
		h.log.Error("failed to create manual game", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(room); err != nil {
		h.log.Error("failed to encode room response", zap.Error(err))
	}
}

type SetStatusRequest struct {
	Status       string  `json:"status"`
	Winner       *string `json:"winner"`
	StatusReason *string `json:"statusReason,omitempty"`
}

func (h *Handler) SetStatus(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "roomId")

	var req SetStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error("failed to decode request body", zap.Error(err))
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.Status == "" {
		http.Error(w, "status is required", http.StatusBadRequest)
		return
	}

	var winner string
	if req.Winner != nil {
		winner = *req.Winner
	}

	var statusReason string
	if req.StatusReason != nil {
		statusReason = *req.StatusReason
	}

	err := h.svc.SetGameStatus(r.Context(), roomID, req.Status, winner, statusReason)
	if err != nil {
		if errors.Is(err, ErrRoomNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if errors.Is(err, ErrRoomFinished) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		h.log.Error("failed to set state", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("room status updated"))
}

func (h *Handler) Heartbeat(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "roomId")
	if roomID == "" {
		http.Error(w, "missing room id", http.StatusBadRequest)
		return
	}

	if err := h.svc.Heartbeat(r.Context(), roomID); err != nil {
		if errors.Is(err, ErrRoomNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if errors.Is(err, errors.New("heartbeat cannot be sent")) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"alive"}`))
}
