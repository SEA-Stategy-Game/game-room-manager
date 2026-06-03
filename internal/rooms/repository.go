package rooms

import "context"

// Repository is the outbound port for listing and updating rooms.
// Later, a database adapter can implement this interface.
type Repository interface {
	List(ctx context.Context) ([]Room, error)
	GetByID(ctx context.Context, roomID string) (*Room, error)
	Update(ctx context.Context, room *Room) error
	Create(ctx context.Context, room *Room) error
	Upsert(ctx context.Context, room *Room) error
}
