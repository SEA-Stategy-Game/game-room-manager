package rooms

import (
	"context"
	"reflect"
	"testing"
	"time"

	"go.uber.org/zap"
)

func newTestRepo(t *testing.T) *SQLiteRepository {
	t.Helper()

	repo, err := NewSQLiteRepository(":memory:", zap.NewNop())
	if err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}

	return repo
}

func sampleRoom() *Room {
	maxPlayers := 4
	// Use truncated time for stable comparison after JSON marshalling/unmarshalling
	now := time.Now().UTC().Truncate(time.Second)
	started := now.Add(5 * time.Minute)
	ended := started.Add(10 * time.Minute)
	return &Room{
		RoomID:             "room-1",
		State:              StateIniting,
		Address:            "localhost",
		Port:               9000,
		Players:            []string{"alice", "bob"},
		MaxNumberOfPlayers: &maxPlayers,
		Winner:             "",
		CreatedAt:          now,
		StartedAt:          &started,
		EndedAt:            &ended,
		ProcessID:          12345,
	}
}

func TestSQLiteRepository_CreateAndGetByID(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	room := sampleRoom()

	if err := repo.Create(ctx, room); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	got, err := repo.GetByID(ctx, room.RoomID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if !reflect.DeepEqual(got, room) {
		t.Errorf("got\n%+v\nwant\n%+v", got, room)
	}
}

func TestSQLiteRepository_List(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	room1 := sampleRoom()
	room2 := sampleRoom()
	room2.RoomID = "room-2"

	if err := repo.Create(ctx, room1); err != nil {
		t.Fatal(err)
	}
	if err := repo.Create(ctx, room2); err != nil {
		t.Fatal(err)
	}

	rooms, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(rooms) != 2 {
		t.Fatalf("expected 2 rooms, got %d", len(rooms))
	}
}

func TestSQLiteRepository_Update(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	room := sampleRoom()

	if err := repo.Create(ctx, room); err != nil {
		t.Fatal(err)
	}

	room.State = StateRunning
	room.Players = append(room.Players, "charlie")
	room.Winner = "charlie"

	if err := repo.Update(ctx, room); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	got, err := repo.GetByID(ctx, room.RoomID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if !reflect.DeepEqual(got, room) {
		t.Errorf("got\n%+v\nwant\n%+v", got, room)
	}
}

func TestSQLiteRepository_GetByID_NotFound(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "missing")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
