package rooms

import (
	"context"
	"errors"
	"flag"
	"time"

	"github.com/google/uuid"
)

var ErrRoomFull = errors.New("room is full")
var ErrRoomNotFound = errors.New("room not found")
var ErrRoomFinished = errors.New("cannot update a room that has already ended or crashed")

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
	return s.repo.ReadModifyWrite(ctx, roomID, func(room *Room) error {
		// Ensure the player isn't already in the room
		for _, p := range room.Players {
			if p == playerID {
				return nil // Idempotent: player is already in the room
			}
		}

		if room.MaxNumberOfPlayers != nil && len(room.Players) >= *room.MaxNumberOfPlayers {
			return ErrRoomFull
		}

		room.Players = append(room.Players, playerID)
		return nil
	})
}

func (s *Service) SetGameStatus(ctx context.Context, roomID string, status string, winner string, statusReason string) error {
	state := State(status)
	return s.repo.ReadModifyWrite(ctx, roomID, func(room *Room) error {
		if room.State == StateEnded || room.State == StateCrashed {
			return ErrRoomFinished
		}

		if state == StateIniting {
			now := time.Now()
			room.StartedAt = &now
		}

		if state == StateEnded {
			room.Winner = winner
			now := time.Now()
			room.EndedAt = &now
		}

		if state == StateCrashed {
			now := time.Now()
			room.EndedAt = &now
		}

		room.State = state
		room.StatusReason = statusReason
		return nil
	})
}

func (s *Service) RegisterGameRoom(ctx context.Context, maxPlayers *int) (*Room, error) {

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

	if maxPlayers == nil {
		defaultValue := 32
		maxPlayers = &defaultValue
	}

	room := &Room{
		RoomID:             uuid.New().String(),
		State:              StateIniting,
		Address:            "",
		Port:               1234,
		Players:            []string{},
		Winner:             "",
		CreatedAt:          time.Now(),
		ProcessID:          pid,
		MaxNumberOfPlayers: maxPlayers,
	}

	return room, s.repo.Create(ctx, room)
}

// RegisterManualGame creates a room record for a game room that was started manually.
func (s *Service) RegisterManualGame(ctx context.Context, roomID string, address string, port int, maxPlayers *int) (*Room, error) {
	if maxPlayers == nil {
		defaultValue := 32
		maxPlayers = &defaultValue
	}

	room := &Room{
		RoomID:             roomID,
		State:              StateReady,
		Address:            address,
		Port:               port,
		Players:            []string{},
		Winner:             "",
		CreatedAt:          time.Now(),
		ProcessID:          0,
		MaxNumberOfPlayers: maxPlayers,
	}

	//Using upsert so that the room is "refreshed" each time it's created
	return room, s.repo.Upsert(ctx, room)
}
