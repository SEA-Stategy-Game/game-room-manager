package rooms

import (
	"context"
	"database/sql"
	"encoding/json"
	"sync"

	"go.uber.org/zap"
	_ "modernc.org/sqlite"
)

type SQLiteRepository struct {
	db      *sql.DB
	log     *zap.Logger
	writeMu sync.Mutex
}

func NewSQLiteRepository(path string, logger *zap.Logger) (*SQLiteRepository, error) {
	// Add query parameters for WAL mode and a busy timeout.
	// WAL (Write-Ahead Logging) allows for higher concurrency.
	// busy_timeout tells SQLite to wait for a bit if the DB is locked,
	// rather than failing immediately with SQLITE_BUSY.
	dsn := path + "?_journal_mode=WAL&_busy_timeout=5000"

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	r := &SQLiteRepository{db: db, log: logger}

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
	} else if err == nil && raw == "" {
		// This can happen if the row exists but the data is empty.
		// Treat as not found.
		return nil, sql.ErrNoRows
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

func (r *SQLiteRepository) Upsert(ctx context.Context, room *Room) error {
	data, err := json.Marshal(room)
	if err != nil {
		return err
	}

	// REPLACE INTO is a SQLite extension that will delete the existing row
	// and insert a new one if a primary key constraint fails.
	_, err = r.db.ExecContext(
		ctx,
		`REPLACE INTO rooms (room_id, data) VALUES (?, ?)`,
		room.RoomID,
		string(data),
	)

	return err
}

func (r *SQLiteRepository) ReadModifyWrite(ctx context.Context, roomID string, modifyFn func(room *Room) error) error {
	r.writeMu.Lock()
	defer r.writeMu.Unlock()

	// Use Serializable isolation to start an IMMEDIATE transaction.
	// This acquires a write lock at the beginning, preventing deadlocks.
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		r.log.Error("ReadModifyWrite: failed to begin transaction", zap.Error(err), zap.String("roomID", roomID))
		return err
	}
	defer tx.Rollback() // Rollback is a no-op if the transaction is committed.

	// 1. Get the room
	row := tx.QueryRowContext(ctx, `SELECT data FROM rooms WHERE room_id = ?`, roomID)

	var raw string
	if err := row.Scan(&raw); err != nil {
		if err == sql.ErrNoRows {
			return ErrRoomNotFound
		}
		return err
	}

	var room Room
	if err := json.Unmarshal([]byte(raw), &room); err != nil {
		return err
	}

	// 2. Apply the modification function provided by the service layer.
	if err := modifyFn(&room); err != nil {
		// This could be a business logic error (like ErrRoomFull), so we just return it.
		return err
	}

	// 3. Write the modified room back to the database.
	data, err := json.Marshal(room)
	if err != nil {
		return err
	}

	if _, err = tx.ExecContext(ctx, `UPDATE rooms SET data = ? WHERE room_id = ?`, string(data), room.RoomID); err != nil {
		return err
	}

	// 4. Commit the transaction.
	if err := tx.Commit(); err != nil {
		r.log.Error("ReadModifyWrite: failed to commit transaction", zap.Error(err), zap.String("roomID", roomID))
		return err
	}
	return nil
}
