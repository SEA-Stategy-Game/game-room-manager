package rooms

// State represents whether a room is active or inactive.
type State string

const (
	StateActive   State = "active"
	StateInactive State = "inactive"
)

// Room is the main entity for the game-room-manager domain.
type Room struct {
	RoomID            string `json:"roomId"`
	ConnectionDetails string `json:"connectionDetails"`
	State             State  `json:"state"`
	Participants      int    `json:"participants"`
	Address 		  string `json:"address"`
	Port 			  int 	 `json:"port"`
}

