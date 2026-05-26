package rooms

import (
	"context"

	"github.com/google/uuid"
	"github.com/moby/moby/client"
)

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

func (s *Service) RegisterGameRoom(ctx context.Context) (Room, error) {

	id := uuid.New().String()

	room := Room{
		RoomID: id,
	}

	// Here we spin up a new docker container with the game running

	return room, nil
}

func startContainer(ctx context.Context, image string) (string, error) {

	cli, err := client.New(client.FromEnv)
	if err != nil {
		return "", err
	}

	container, err := cli.ContainerCreate(ctx, client.ContainerCreateOptions{Image: image})

	_, err = cli.ContainerStart(ctx, container.ID, client.ContainerStartOptions{})

	return container.ID, err

}
