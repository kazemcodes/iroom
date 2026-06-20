package handler

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"

	repository "github.com/iroom/iroom/internal/adapter/repository/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	schema := `
	CREATE TABLE IF NOT EXISTS rooms (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		owner_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		description TEXT DEFAULT '',
		color TEXT DEFAULT '',
		slug TEXT UNIQUE NOT NULL,
		guest_login_enabled BOOLEAN DEFAULT 1,
		max_users INTEGER DEFAULT 50,
		invite_code TEXT DEFAULT '',
		is_archived BOOLEAN DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS room_users (
		room_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		role TEXT DEFAULT 'student',
		access INTEGER DEFAULT 1,
		PRIMARY KEY (room_id, user_id),
		FOREIGN KEY (room_id) REFERENCES rooms(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS room_settings (
		room_id INTEGER PRIMARY KEY,
		max_users INTEGER DEFAULT 50,
		recording_enabled BOOLEAN DEFAULT 1,
		allow_student_video BOOLEAN DEFAULT 0,
		allow_student_audio BOOLEAN DEFAULT 1,
		allow_student_screen_share BOOLEAN DEFAULT 0,
		allow_student_whiteboard BOOLEAN DEFAULT 0,
		allow_student_chat BOOLEAN DEFAULT 1,
		session_auto_end_minutes INTEGER DEFAULT 120,
		waiting_room_enabled BOOLEAN DEFAULT 0,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (room_id) REFERENCES rooms(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT DEFAULT '',
		display_name TEXT DEFAULT '',
		role TEXT DEFAULT 'student',
		phone TEXT DEFAULT '',
		avatar_url TEXT DEFAULT '',
		is_active BOOLEAN DEFAULT 1,
		totp_secret TEXT DEFAULT '',
		totp_enabled BOOLEAN DEFAULT 0,
		totp_backup_codes TEXT DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		room_id INTEGER,
		class_id INTEGER NOT NULL,
		title TEXT DEFAULT '',
		scheduled_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		duration INTEGER DEFAULT 60,
		status TEXT DEFAULT 'scheduled',
		livekit_room TEXT DEFAULT '',
		recording_url TEXT DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err = db.Exec(schema)
	require.NoError(t, err)

	return db
}

func newTestRoomRepo(t *testing.T, db *sql.DB) *repository.RoomRepo {
	t.Helper()
	return repository.NewRoomRepo(db)
}

func newTestUserRepo(t *testing.T, db *sql.DB) *repository.UserRepo {
	t.Helper()
	return repository.NewUserRepo(db)
}

func newTestSessionRepo(t *testing.T, db *sql.DB) *repository.SessionRepo {
	t.Helper()
	return repository.NewSessionRepo(db)
}
