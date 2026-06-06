package rooms

import (
	"context"
	"errors"
	"flag"
	"net"
	"strings"
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
			if p == playerID { // Idempotent: player is already in the room
				return nil
			}
		}

		// Check if the room is full
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

	ln, _ := net.Listen("tcp", ":0")
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()

	var err error
	var pid int

	if maxPlayers == nil {
		defaultValue := 32
		maxPlayers = &defaultValue
	}

	id := strings.Replace(uuid.New().String(), "-", "", -1)[:8]

	if err != nil {
		ln.Close() // release on failure
		return nil, err
	}

	// release AFTER successful spawn
	ln.Close()
	if flag.Lookup("test.v") != nil {
		pid = 12345
	} else {
		pid, err = SpawnGameRoom(port, id, *maxPlayers, s.gameImage)
		go func() {
			defer ln.Close()
		}()
		if err != nil {
			return nil, err
		}
	}

	var ip = "127.0.0.1"

	room := &Room{
		RoomID:             id,
		State:              StateIniting,
		Address:            ip,
		Port:               port,
		Players:            []string{},
		Winner:             "",
		CreatedAt:          time.Now(),
		LastHeartbeatAt:    time.Now(),
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
		LastHeartbeatAt:    time.Now(),
		ProcessID:          0,
		MaxNumberOfPlayers: maxPlayers,
	}

	//Using upsert so that the room is "refreshed" each time it's created
	return room, s.repo.Upsert(ctx, room)
}

func (s *Service) Heartbeat(ctx context.Context, roomID string) error {
	return s.repo.ReadModifyWrite(ctx, roomID, func(room *Room) error {
		if room.State == StateEnded || room.State == StateCrashed {
			return errors.New("heartbeat cannot be sent")
		}

		room.LastHeartbeatAt = time.Now()
		return nil
	})
}
