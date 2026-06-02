package rooms

import (
	"context"
	"testing"
)

func newTestRepo(t *testing.T) *SQLiteRepository {
	t.Helper()

	repo, err := NewSQLiteRepository(":memory:")
	if err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}

	return repo
}

func sampleRoom() *Room {
	return &Room{
		RoomID:            "room-1",
		State:             StateIniting,
		Address:           "localhost",
		Port:              9000,
		Players:           []string{"alice", "bob"},
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

	if got.RoomID != room.RoomID {
		t.Fatalf("expected %s, got %s", room.RoomID, got.RoomID)
	}

	if len(got.Players) != 2 {
		t.Fatalf("expected 2 players, got %d", len(got.Players))
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

	room.Players = append(room.Players, "charlie")

	if err := repo.Update(ctx, room); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	got, err := repo.GetByID(ctx, room.RoomID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if len(got.Players) != 3 {
		t.Fatalf("expected 3 players, got %d", len(got.Players))
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
