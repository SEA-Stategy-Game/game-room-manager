package rooms

import (
	"encoding/json"
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

	if err := h.svc.JoinGameRoom(r.Context(), roomID, playerID); err != nil {
		h.log.Error("failed to join room", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("player joined"))
}

func (h *Handler) CreateGame(w http.ResponseWriter, r *http.Request) {
	room, err := h.svc.RegisterGameRoom(r.Context())
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

func (h *Handler) SetReady(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "roomId")

	if err := h.svc.ReadyGameRoom(r.Context(), roomID); err != nil {
		h.log.Error("failed to set to ready state", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("room set to ready"))
}

func (h *Handler) SetCrashed(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "roomId")

	if err := h.svc.CrashGameRoom(r.Context(), roomID); err != nil {
		h.log.Error("failed to set to crashed state", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("room set to crashed"))
}

func (h *Handler) SetEnded(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "roomId")
	winner := chi.URLParam(r, "winner")

	if err := h.svc.EndGameRoom(r.Context(), roomID, winner); err != nil {
		h.log.Error("failed to set to ended state", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("room set to ended"))
}
