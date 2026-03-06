package rooms

import (
	"encoding/json"
	"net/http"

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

