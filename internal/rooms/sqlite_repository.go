package rooms

import (
	"context"
	"database/sql"
	"encoding/json"

	_ "modernc.org/sqlite"
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(path string) (*SQLiteRepository, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	r := &SQLiteRepository{db: db}

	if err := r.init(); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *SQLiteRepository) init() error {
	_, err := r.db.Exec(`
	CREATE TABLE IF NOT EXISTS rooms (
		room_id TEXT PRIMARY KEY,
		data    TEXT NOT NULL
	);
	`)
	return err
}

func (r *SQLiteRepository) Create(ctx context.Context, room *Room) error {
	data, err := json.Marshal(room)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(
		ctx,
		`INSERT INTO rooms (room_id, data) VALUES (?, ?)`,
		room.RoomID,
		string(data),
	)

	return err
}

func (r *SQLiteRepository) GetByID(ctx context.Context, roomID string) (*Room, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT data FROM rooms WHERE room_id = ?`,
		roomID,
	)

	var raw string
	if err := row.Scan(&raw); err != nil {
		return nil, err
	}

	var room Room
	if err := json.Unmarshal([]byte(raw), &room); err != nil {
		return nil, err
	}

	return &room, nil
}

func (r *SQLiteRepository) List(ctx context.Context) ([]Room, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT data FROM rooms`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Room

	for rows.Next() {
		var raw string

		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}

		var room Room
		if err := json.Unmarshal([]byte(raw), &room); err != nil {
			return nil, err
		}

		result = append(result, room)
	}

	// TODO: Remove this, it's a hardcoded room for testing now 
	if len(result) == 0 {
		result = append(result, Room{
			
			RoomID:            "testgame",
			ConnectionDetails: "ws://localhost:8080/rooms/room-1",
			State:             StateActive,
			Participants:      3,
			Address:           "127.0.0.1",
			Port:              12345,
			Players:           []string{},
		})
	}

	return result, rows.Err()
}

func (r *SQLiteRepository) Update(ctx context.Context, room *Room) error {
	data, err := json.Marshal(room)
	if err != nil {
		return err
	}

	res, err := r.db.ExecContext(
		ctx,
		`UPDATE rooms SET data = ? WHERE room_id = ?`,
		string(data),
		room.RoomID,
	)

	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
