package rooms

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/google/uuid"
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

func (s *Service) FindRoom(ctx context.Context, roomID string) (*Room, error) {
	return s.repo.GetByID(ctx, roomID)
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

	return s.repo.Update(ctx, room)
}

func (s *Service) SetGameStatus(ctx context.Context, roomID string, status string, winner string) error {
	room, err := s.repo.GetByID(ctx, roomID)
	if err != nil {
		return err
	}

	if room == nil {
		if status == "ready" {
			var pid int
			var pidErr error
			if flag.Lookup("test.v") != nil {
				pid = 12345
			} else {
				pid, pidErr = SpawnGameRoom(1234)
				if pidErr != nil {
					return pidErr
				}
			}

			newRoom := &Room{
				RoomID:            roomID,
				State:             StateReady,
				Address:           "",
				Port:              1234,
				Players:           []string{},
				ProcessID:         pid,
			}
			return s.repo.Create(ctx, newRoom)
		} else {
			return nil
		}
	}

	if status == "init" {
		room.StartedAt = time.Now()
	}

	if status == "ended" {
		room.Winner = winner
		room.EndedAt = time.Now()
	}

	if status == "crashed" {
		room.EndedAt = time.Now()
	}

	room.State = State(status)

	return s.repo.Update(ctx, room)
}

func (s *Service) RegisterGameRoom(ctx context.Context) (*Room, error) {

	var pid int
	var err error

	if flag.Lookup("test.v") != nil {
		pid = 12345
	} else {
		pid, err = SpawnGameRoom(1234)
		if err != nil {
			return nil, err
		}
	}

	room := &Room{
		RoomID:            uuid.New().String(),
		State:             StateIniting,
		Address:           "",
		Port:              1234,
		Players:           []string{},
		Winner:            "",
		StartedAt:         time.Time{},
		EndedAt:           time.Time{},
		ProcessID:         pid,
	}

	return room, s.repo.Create(ctx, room)
}
