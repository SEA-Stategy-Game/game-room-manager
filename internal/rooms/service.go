package rooms

import (
	"context"
	"fmt"
)

// Service is the application/service layer (use-cases) for rooms.
type Service struct {
	repo      Repository
	gameImage string
}

func NewService(repo Repository, gameImage string) *Service {
	return &Service{
		repo:      repo,
		gameImage: gameImage,
	}
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

	// Ensure the player isn't already in the room
	for _, p := range room.Players {
		if p == playerID {
			return nil
		}
	}

	room.Players = append(room.Players, playerID)
	room.Participants++

	return s.repo.Update(ctx, room)
}

func (s *Service) RegisterGameRoom(ctx context.Context) (Room, error) {

	// This spawns a new process via os/exec and returns the PID of the child process
	pid, err := SpawnGameRoom(1234)

	if err != nil {
		return Room{}, err
	}

	room := Room{
		RoomID:       fmt.Sprint(pid),
		Participants: 0,
		Port:         1234,
		State:        StateActive,
		Players:      []string{},
	}

	return room, nil
}
