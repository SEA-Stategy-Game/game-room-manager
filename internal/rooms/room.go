package rooms

import "time"

// State represents whether a room is active or inactive.
type State string

const (
	// Deprecated, here for historical reasons
	StateActive   State = "active"
	StateInactive State = "inactive"
	// States:
	StateIniting State = "initing"
	StateReady   State = "ready"
	StateRunning State = "running"
	StateEnded   State = "ended"
	StateCrashed State = "crashed"
)

// Room is the main entity for the game-room-manager domain.
type Room struct {
	RoomID            string    `json:"roomId"`
	ConnectionDetails string    `json:"connectionDetails"`
	State             State     `json:"state"`
	Participants      int       `json:"participants"`
	Address           string    `json:"address"`
	Port              int       `json:"port"`
	Players           []string  `json:"players"`
	Winner            string    `json:"winner"`
	StartedAt         time.Time `json:"startedAt"`
	EndedAt           time.Time `json:"endedAt"`
}
