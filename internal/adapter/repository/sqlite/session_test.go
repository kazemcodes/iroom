package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// seedRoomAndSession creates a room (and its room_settings) plus a session.
// Returns (roomID, sessionID).
func seedRoomAndSession(t *testing.T, repo *RoomRepo, status string, autoEndMinutes int) (roomID int64, sessionID int64) {
	t.Helper()

	// Insert a user if needed
	_, _ = repo.db.Exec(`INSERT OR IGNORE INTO users (id, email, display_name, role) VALUES (1, 'owner@test.com', 'Owner', 'admin')`)

	// Insert a room
	res, err := repo.db.Exec(`INSERT INTO rooms (owner_id, name, slug) VALUES (1, 'Test Room', 'session-test-room')`)
	require.NoError(t, err)
	roomID, err = res.LastInsertId()
	require.NoError(t, err)

	// Insert room_settings
	_, err = repo.db.Exec(`INSERT OR REPLACE INTO room_settings (room_id, session_auto_end_minutes, waiting_room_enabled)
		VALUES (?, ?, 0)`, roomID, autoEndMinutes)
	require.NoError(t, err)

	// Insert a session
	res, err = repo.db.Exec(`INSERT INTO sessions (room_id, class_id, title, status) VALUES (?, ?, 'Test Session', ?)`,
		roomID, roomID, status)
	require.NoError(t, err)
	sessionID, err = res.LastInsertId()
	require.NoError(t, err)

	return
}

// seedSessionOnly inserts a session without room_settings (edge case: no settings row yet).
func seedSessionOnly(t *testing.T, repo *RoomRepo, status string) (roomID int64, sessionID int64) {
	t.Helper()
	_, _ = repo.db.Exec(`INSERT OR IGNORE INTO users (id, email, display_name, role) VALUES (1, 'owner@test.com', 'Owner', 'admin')`)

	res, err := repo.db.Exec(`INSERT INTO rooms (owner_id, name, slug) VALUES (1, 'NoSettings Room', 'nosettings-room')`)
	require.NoError(t, err)
	roomID, err = res.LastInsertId()
	require.NoError(t, err)

	res, err = repo.db.Exec(`INSERT INTO sessions (room_id, class_id, title, status) VALUES (?, ?, 'Test Session', ?)`,
		roomID, roomID, status)
	require.NoError(t, err)
	sessionID, err = res.LastInsertId()
	require.NoError(t, err)

	return
}

// ---------------------------------------------------------------------------
// IsSessionLive
// ---------------------------------------------------------------------------

func TestSessionRepo_IsSessionLive_TrueForLive(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	roomRepo := NewRoomRepo(db)
	sessionRepo := NewSessionRepo(db)
	_, sessionID := seedRoomAndSession(t, roomRepo, "live", 120)

	assert.True(t, sessionRepo.IsSessionLive(sessionID))
}

func TestSessionRepo_IsSessionLive_FalseForScheduled(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	roomRepo := NewRoomRepo(db)
	sessionRepo := NewSessionRepo(db)
	_, sessionID := seedRoomAndSession(t, roomRepo, "scheduled", 120)

	assert.False(t, sessionRepo.IsSessionLive(sessionID))
}

func TestSessionRepo_IsSessionLive_FalseForEnded(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	roomRepo := NewRoomRepo(db)
	sessionRepo := NewSessionRepo(db)
	_, sessionID := seedRoomAndSession(t, roomRepo, "ended", 120)

	assert.False(t, sessionRepo.IsSessionLive(sessionID))
}

func TestSessionRepo_IsSessionLive_FalseForNonexistent(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	sessionRepo := NewSessionRepo(db)

	assert.False(t, sessionRepo.IsSessionLive(9999))
}

// ---------------------------------------------------------------------------
// GetAutoEndMinutesBySessionID
// ---------------------------------------------------------------------------

func TestSessionRepo_GetAutoEndMinutes_ReturnsConfiguredValue(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	roomRepo := NewRoomRepo(db)
	sessionRepo := NewSessionRepo(db)
	_, sessionID := seedRoomAndSession(t, roomRepo, "live", 45)

	assert.Equal(t, 45, sessionRepo.GetAutoEndMinutesBySessionID(sessionID))
}

func TestSessionRepo_GetAutoEndMinutes_ReturnsZeroWhenDisabled(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	roomRepo := NewRoomRepo(db)
	sessionRepo := NewSessionRepo(db)
	_, sessionID := seedRoomAndSession(t, roomRepo, "live", 0)

	assert.Equal(t, 0, sessionRepo.GetAutoEndMinutesBySessionID(sessionID))
}

func TestSessionRepo_GetAutoEndMinutes_ReturnsZeroWhenSessionNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	sessionRepo := NewSessionRepo(db)

	assert.Equal(t, 0, sessionRepo.GetAutoEndMinutesBySessionID(9999))
}

func TestSessionRepo_GetAutoEndMinutes_ReturnsZeroWhenNoSettingsRow(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	roomRepo := NewRoomRepo(db)
	sessionRepo := NewSessionRepo(db)
	_, sessionID := seedSessionOnly(t, roomRepo, "live")

	// No room_settings row → JOIN produces no rows → Scan error → returns 0
	assert.Equal(t, 0, sessionRepo.GetAutoEndMinutesBySessionID(sessionID))
}

func TestSessionRepo_GetAutoEndMinutes_ReturnsDefault120WhenSet(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	roomRepo := NewRoomRepo(db)
	sessionRepo := NewSessionRepo(db)
	_, sessionID := seedRoomAndSession(t, roomRepo, "live", 120)

	assert.Equal(t, 120, sessionRepo.GetAutoEndMinutesBySessionID(sessionID))
}

func TestSessionRepo_GetAutoEndMinutes_ReturnsZeroForNegativeValue(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	roomRepo := NewRoomRepo(db)
	sessionRepo := NewSessionRepo(db)
	_, sessionID := seedRoomAndSession(t, roomRepo, "live", 120)

	// Manually set a negative value to test the <= 0 guard
	_, err := db.Exec(`UPDATE room_settings SET session_auto_end_minutes = -1 WHERE room_id = (
		SELECT room_id FROM sessions WHERE id = ?)`, sessionID)
	require.NoError(t, err)

	assert.Equal(t, 0, sessionRepo.GetAutoEndMinutesBySessionID(sessionID))
}
