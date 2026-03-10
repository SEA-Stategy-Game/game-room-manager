package rooms

import "context"

// InMemoryRepository is a simple adapter that serves hard-coded rooms.
// It can be replaced by a database-backed repository later.
type InMemoryRepository struct {
	rooms []Room
}

func NewInMemoryRepository(rooms []Room) *InMemoryRepository {
	copied := make([]Room, len(rooms))
	copy(copied, rooms)
	return &InMemoryRepository{rooms: copied}
}

func NewDefaultInMemoryRepository() *InMemoryRepository {
	return NewInMemoryRepository([]Room{
		{
			RoomID:            "room-1",
			ConnectionDetails: "ws://localhost:8080/rooms/room-1",
			State:             StateActive,
			Participants:      3,
		},
		{
			RoomID:            "room-2",
			ConnectionDetails: "ws://localhost:8080/rooms/room-2",
			State:             StateInactive,
			Participants:      0,
		},
	})
}

func (r *InMemoryRepository) List(ctx context.Context) ([]Room, error) {
	_ = ctx
	out := make([]Room, len(r.rooms))
	copy(out, r.rooms)
	return out, nil
}

