package rooms

import "context"

// Repository is the outbound port for listing rooms.
// Later, a database adapter can implement this interface.
type Repository interface {
	List(ctx context.Context) ([]Room, error)
}

