package rooms

import (
	"context"
	"encoding/json"
	"os"
)

// JSONRepository stores rooms in memory but persists them to a local JSON file.
// It implements the same Repository interface as InMemoryRepository.
type JSONRepository struct {
	path  string
	rooms []Room
}

func NewJSONRepository(path string) (*JSONRepository, error) {
	r := &JSONRepository{path: path}

	// load existing rooms if the file is there
	data, err := os.ReadFile(path)
	if err == nil {
		json.Unmarshal(data, &r.rooms)
	}

	return r, nil
}

// save writes the current rooms slice back to the JSON file.
// This method is not thread-safe and must be called from a method that holds a write lock.
func (r *JSONRepository) save() error {
	data, err := json.MarshalIndent(r.rooms, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(r.path, data, 0644)
}

func (r *JSONRepository) List(ctx context.Context) ([]Room, error) {
	_ = ctx
	out := make([]Room, len(r.rooms))
	copy(out, r.rooms)
	return out, nil
}

func (r *JSONRepository) GetByID(ctx context.Context, roomID string) (*Room, error) {
	_ = ctx
	for i := range r.rooms {
		if r.rooms[i].RoomID == roomID {
			room := r.rooms[i]
			return &room, nil
		}
	}
	return nil, nil
}

func (r *JSONRepository) Update(ctx context.Context, room *Room) error {
	_ = ctx
	for i := range r.rooms {
		if r.rooms[i].RoomID == room.RoomID {
			r.rooms[i] = *room
			return r.save()
		}
	}
	return nil
}

func (r *JSONRepository) Create(ctx context.Context, room *Room) error {
	_ = ctx
	r.rooms = append(r.rooms, *room)
	return r.save()
}

func (r *JSONRepository) Upsert(ctx context.Context, room *Room) error {
	_ = ctx
	for i := range r.rooms {
		if r.rooms[i].RoomID == room.RoomID {
			r.rooms[i] = *room
			return r.save()
		}
	}
	r.rooms = append(r.rooms, *room)
	return r.save()
}

func (r *JSONRepository) ReadModifyWrite(ctx context.Context, roomID string, modifyFn func(room *Room) error) error {
	_ = ctx
	var room *Room
	for i := range r.rooms {
		if r.rooms[i].RoomID == roomID {
			room = &r.rooms[i]
		}
	}

	if room == nil {
		return ErrRoomNotFound
	}

	// Apply the modification.
	if err := modifyFn(room); err != nil {
		return err
	}

	// Persist the changes to the file.
	return r.save()
}
