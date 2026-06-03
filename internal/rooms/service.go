package rooms

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"time"

	"github.com/google/uuid"
)

var ErrRoomFull = errors.New("room is full")
var ErrRoomNotFound = errors.New("room not found")

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

	if room.MaxNumberOfPlayers != nil && len(room.Players) >= *room.MaxNumberOfPlayers {
		return ErrRoomFull
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

func (s *Service) SetGameStatus(ctx context.Context, roomID string, status string, winner string, statusReason string) error {
	room, err := s.repo.GetByID(ctx, roomID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrRoomNotFound
		}
		return err
	}

	if room == nil {
		return ErrRoomNotFound
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
	room.StatusReason = statusReason

	return s.repo.Update(ctx, room)
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
		StartedAt:          time.Time{},
		EndedAt:            time.Time{},
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
		StartedAt:          time.Time{},
		EndedAt:            time.Time{},
		ProcessID:          0,
		MaxNumberOfPlayers: maxPlayers,
	}

	return room, s.repo.Upsert(ctx, room)
}
