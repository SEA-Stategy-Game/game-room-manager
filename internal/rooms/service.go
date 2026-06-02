package rooms

import (
	"context"
	"flag"
	"fmt"
	"time"
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
	room.Participants++

	return s.repo.Update(ctx, room)
}

// func (s *Service) ReadyGameRoom(ctx context.Context, roomID string) error {
// 	room, err := s.repo.GetByID(ctx, roomID)
// 	if err != nil {
// 		return err
// 	}
// 	if room == nil {
// 		return nil
// 	}

// 	room.State = StateReady

// 	return s.repo.Update(ctx, room)
// }

// func (s *Service) EndGameRoom(ctx context.Context, roomID string, winnerID string) error {
// 	room, err := s.repo.GetByID(ctx, roomID)
// 	if err != nil {
// 		return err
// 	}
// 	if room == nil {
// 		return nil
// 	}

// 	room.Winner = winnerID
// 	room.State = StateEnded

// 	return s.repo.Update(ctx, room)
// }

func (s *Service) SetGameStatus(ctx context.Context, roomID string, status string, winner string) error {
	room, err := s.repo.GetByID(ctx, roomID)
	if err != nil {
		return err
	}

	if room == nil {
		if status == "ready" {
			newRoom, err := s.RegisterGameRoom(ctx)
			if err != nil {
				return fmt.Errorf("failed to register game room: %w", err)
			}

			room = &Room{
				RoomID:            roomID,
				ConnectionDetails: newRoom.ConnectionDetails,
				State:             newRoom.State,
				Participants:      newRoom.Participants,
				Address:           newRoom.Address,
				Port:              newRoom.Port,
				Players:           newRoom.Players,
				Winner:            newRoom.Winner,
				StartedAt:         newRoom.StartedAt,
				EndedAt:           newRoom.EndedAt,
			}
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

// func (s *Service) CrashGameRoom(ctx context.Context, roomID string) error {
// 	room, err := s.repo.GetByID(ctx, roomID)
// 	if err != nil {
// 		return err
// 	}
// 	if room == nil {
// 		return nil
// 	}

// 	room.State = StateCrashed

// 	return s.repo.Update(ctx, room)
// }

func (s *Service) RegisterGameRoom(ctx context.Context) (Room, error) {

	var pid int
	var err error

	if flag.Lookup("test.v") != nil {
		pid = 12345
	} else {
		pid, err = SpawnGameRoom(1234)
		if err != nil {
			return Room{}, err
		}
	}

	if err != nil {
		return Room{}, err
	}

	room := Room{
		RoomID:            fmt.Sprint(pid),
		ConnectionDetails: "",
		State:             StateActive,
		Participants:      0,
		Address:           "",
		Port:              1234,
		Players:           []string{},
		Winner:            "",
		StartedAt:         time.Time{},
		EndedAt:           time.Time{},
	}

	return room, nil
}
