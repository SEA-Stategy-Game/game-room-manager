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

func (h *Handler) GetRooms(w http.ResponseWriter, r *http.Request) {
	rooms, err := h.svc.ListRooms(r.Context())
	if err != nil {
		h.log.Error("failed to list rooms", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(rooms); err != nil {
		h.log.Error("failed to encode rooms response", zap.Error(err))
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
