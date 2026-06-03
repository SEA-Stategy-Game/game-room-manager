package rooms

import "time"

// State represents whether a room is active or inactive.
type State string

const (
	// States:
	StateIniting State = "initing"
	StateReady   State = "ready"
	StateRunning State = "running"
	StateEnded   State = "ended"
	StateCrashed State = "crashed"
)

// Room is the main entity for the game-room-manager domain.
type Room struct {
	RoomID              string    `json:"roomId"`
	State               State     `json:"state"`
	Address             string    `json:"address"`
	Port                int       `json:"port"`
	Players             []string  `json:"players"`
	MaxNumberOfPlayers *int      `json:"maxNumberOfPlayers,omitempty"`
	Winner              string    `json:"winner"`
	StatusReason        string    `json:"statusReason,omitempty"`
	StartedAt           time.Time `json:"startedAt"`
	EndedAt             time.Time `json:"endedAt"`
	ProcessID           int       `json:"processId"`
}
