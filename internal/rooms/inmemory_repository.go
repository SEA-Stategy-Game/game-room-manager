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
			Address:           "127.0.0.1",
			Port:              12345,
			Players:           []string{},
		},
		{
			RoomID:            "room-2",
			ConnectionDetails: "ws://localhost:8080/rooms/room-2",
			State:             StateInactive,
			Participants:      0,
			Players:           []string{},
		},
	})
}

func (r *InMemoryRepository) List(ctx context.Context) ([]Room, error) {
	_ = ctx
	out := make([]Room, len(r.rooms))
	copy(out, r.rooms)
	return out, nil
}

func (r *InMemoryRepository) GetByID(ctx context.Context, roomID string) (*Room, error) {
	_ = ctx
	for i := range r.rooms {
		if r.rooms[i].RoomID == roomID {
			return &r.rooms[i], nil
		}
	}
	return nil, nil
}

func (r *InMemoryRepository) Update(ctx context.Context, room *Room) error {
	_ = ctx
	for i := range r.rooms {
		if r.rooms[i].RoomID == room.RoomID {
			r.rooms[i] = *room
			return nil
		}
	}
	r.rooms = append(r.rooms, *room)
	return nil
}

func (r *InMemoryRepository) Create(ctx context.Context, room *Room) error {
	_ = ctx
	r.rooms = append(r.rooms, *room)
	return nil
}
