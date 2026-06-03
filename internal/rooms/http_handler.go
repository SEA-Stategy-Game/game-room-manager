package rooms

import (
	"encoding/json"
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
func findPlayer(rooms []Room, searchPlayer string) ([]Room, error) {
	var found []Room

	for _, room := range rooms {
		for _, player := range room.Players {
			if player == searchPlayer {
				found = append(found, room)
			}
		}
	}
	return found, nil
}

func findByStatus(rooms []Room, searchStat string) ([]Room, error) {
	var found []Room

	targetState := State(searchStat)

	for _, room := range rooms {
		if room.State == targetState {
			found = append(found, room)
		}
	}
	return found, nil
}

func findDouble(rooms []Room, searchPlayer string, searchStat string) ([]Room, error) {
	var found []Room

	targetState := State(searchStat)

	for _, room := range rooms {
		if room.State == targetState {
			for _, player := range room.Players {
				if player == searchPlayer {
					found = append(found, room)
				}
			}
		}
	}
	return found, nil
}

func (h *Handler) GetRooms(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	rooms, err := h.svc.ListRooms(r.Context())

	// Check for a query, if not, just show all.
	player := query.Get("player")
	status := query.Get("status")
	if player != "" && status == "" {
		rooms, err = findPlayer(rooms, player)
	}

	if status != "" && player == "" {
		rooms, err = findByStatus(rooms, status)
	}

	if player != "" && status != "" {
		rooms, err = findDouble(rooms, player, status)
	}

	if err != nil {
		h.log.Error("failed to list rooms", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// h.log.Info("retrieved rooms", zap.Any("rooms", rooms))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(rooms); err != nil {
		h.log.Error("failed to encode rooms response", zap.Error(err))
	}
}

func (h *Handler) GetRoom(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "roomId")

	rooms, err := h.svc.FindRoom(r.Context(), roomID)

	if err != nil {
		h.log.Error("failed to find room", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// h.log.Info("retrieved rooms", zap.Any("rooms", rooms))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(rooms); err != nil {
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
	Status string  `json:"status"`
	Winner *string `json:"winner"`
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

	err := h.svc.SetGameStatus(r.Context(), roomID, req.Status, winner)
	if err != nil {
		if errors.Is(err, ErrRoomNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		h.log.Error("failed to set state", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("room status updated"))
}
