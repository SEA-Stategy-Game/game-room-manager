package rooms

import "context"

// Service is the application/service layer (use-cases) for rooms.
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListRooms(ctx context.Context) ([]Room, error) {
	return s.repo.List(ctx)
}

func (s *Service) JoinGameRoom(ctx context.Context, roomID string, playerID string) error {
	room, err := s.repo.GetByID(ctx, roomID)
	if err != nil {
		return err
	}
	if room == nil {
		return nil
	}

	room.Players = append(room.Players, playerID)
	room.Participants++

	return s.repo.Update(ctx, room)
}

