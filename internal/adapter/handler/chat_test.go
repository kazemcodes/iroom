package handler

import (
	"database/sql"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/iroom/iroom/internal/pkg/jwt"
	"github.com/iroom/iroom/internal/services"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockSessionUC records AutoEnd calls for verification in tests.
type mockSessionUC struct {
	autoEndCalls []int64
	mu           sync.Mutex
	autoEndCh    chan int64 // signaled on each AutoEnd call; nil = no channel
}

func (m *mockSessionUC) AutoEnd(id int64) error {
	m.mu.Lock()
	m.autoEndCalls = append(m.autoEndCalls, id)
	m.mu.Unlock()
	if m.autoEndCh != nil {
		m.autoEndCh <- id
	}
	return nil
}

func (m *mockSessionUC) getAutoEndCalls() []int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	c := make([]int64, len(m.autoEndCalls))
	copy(c, m.autoEndCalls)
	return c
}

// newMockSessionUC creates a mock sessionUC without a notification channel.
func newMockSessionUC() *mockSessionUC {
	return &mockSessionUC{}
}

// newMockSessionUCWithCh creates a mock sessionUC that sends the sessionID
// on `ch` each time AutoEnd is called.
func newMockSessionUCWithCh(ch chan int64) *mockSessionUC {
	return &mockSessionUC{autoEndCh: ch}
}

// newTestChatHandler creates a ChatHandler wired to an in-memory SQLite DB
// and a standalone hub so BroadcastToRoom doesn't panic.
func newTestChatHandler(t *testing.T, db *sql.DB, sessionUC *mockSessionUC) *ChatHandler {
	t.Helper()
	return NewChatHandler(
		nil,            // messageRepo — not needed for auto-end tests
		newTestUserRepo(t, db),
		newTestSessionRepo(t, db),
		sessionUC,      // satisfies the interface: AutoEnd(int64) error
		"test-secret",
		services.NewHub(),
	)
}

// seedTestRoomAndSession inserts a room, its room_settings, and a session
// into the test DB. Returns (sessionID, ownerID).
func seedTestRoomAndSession(t *testing.T, db *sql.DB, status string, autoEndMinutes int) (sessionID int64, ownerID int64) {
	t.Helper()
	return seedTestRoomAndSessionWithSlug(t, db, status, autoEndMinutes, "test-room")
}

// seedTestRoomAndSessionWithSlug is like seedTestRoomAndSession but accepts a
// unique slug to avoid conflicts when seeding multiple rooms.
func seedTestRoomAndSessionWithSlug(t *testing.T, db *sql.DB, status string, autoEndMinutes int, slug string) (sessionID int64, ownerID int64) {
	t.Helper()

	// Insert a user (owner)
	ownerID = 1
	_, err := db.Exec(`INSERT OR IGNORE INTO users (id, email, display_name, role) VALUES (?, 'owner@test.com', 'Owner', 'admin')`, ownerID)
	require.NoError(t, err)

	// Insert a room
	res, err := db.Exec(`INSERT INTO rooms (owner_id, name, slug) VALUES (?, ?, ?)`, ownerID, slug, slug)
	require.NoError(t, err)
	roomID, err := res.LastInsertId()
	require.NoError(t, err)

	// Insert room_settings — needed by GetAutoEndMinutesBySessionID JOIN
	_, err = db.Exec(`INSERT OR REPLACE INTO room_settings (room_id, session_auto_end_minutes, waiting_room_enabled)
		VALUES (?, ?, 0)`, roomID, autoEndMinutes)
	require.NoError(t, err)

	// Insert a session
	res, err = db.Exec(`INSERT INTO sessions (room_id, class_id, title, status) VALUES (?, ?, 'Test Session', ?)`,
		roomID, roomID, status)
	require.NoError(t, err)
	sessionID, err = res.LastInsertId()
	require.NoError(t, err)

	return
}

func TestChatHandler_AutoEnd_FiresOnTimeout(t *testing.T) {
	// Speed up time for the test
	orig := autoEndTimeUnit
	autoEndTimeUnit = 50 * time.Millisecond
	t.Cleanup(func() { autoEndTimeUnit = orig })

	db := setupTestDB(t)
	defer db.Close()

	_, sessionID := seedTestRoomAndSession(t, db, "live", 1)

	ch := make(chan int64, 1)
	mockUC := newMockSessionUCWithCh(ch)
	handler := newTestChatHandler(t, db, mockUC)

	roomIDStr := strconv.FormatInt(sessionID, 10)

	// Start the auto-end timer
	handler.startAutoEnd(roomIDStr, sessionID)

	// Wait for AutoEnd to be called (within timeout + buffer)
	select {
	case calledID := <-ch:
		assert.Equal(t, sessionID, calledID, "AutoEnd should be called with the correct session ID")
	case <-time.After(2 * time.Second):
		t.Fatal("timeout: AutoEnd was not called")
	}

	// Verify only one call
	assert.Len(t, mockUC.getAutoEndCalls(), 1)
}

func TestChatHandler_AutoEnd_CancelsOnRejoin(t *testing.T) {
	orig := autoEndTimeUnit
	autoEndTimeUnit = 50 * time.Millisecond
	t.Cleanup(func() { autoEndTimeUnit = orig })

	db := setupTestDB(t)
	defer db.Close()

	_, sessionID := seedTestRoomAndSession(t, db, "live", 1)

	mockUC := newMockSessionUC()
	handler := newTestChatHandler(t, db, mockUC)

	roomIDStr := strconv.FormatInt(sessionID, 10)

	// Start the auto-end timer
	handler.startAutoEnd(roomIDStr, sessionID)

	// Immediately cancel (simulating operator rejoining)
	handler.cancelAutoEnd(roomIDStr)

	// Wait well past the timer duration
	time.Sleep(300 * time.Millisecond)

	// AutoEnd should NOT have been called
	assert.Empty(t, mockUC.getAutoEndCalls(), "AutoEnd should not be called after cancel")
}

func TestChatHandler_AutoEnd_CancelAutoEndIdempotent(t *testing.T) {
	orig := autoEndTimeUnit
	autoEndTimeUnit = 50 * time.Millisecond
	t.Cleanup(func() { autoEndTimeUnit = orig })

	db := setupTestDB(t)
	defer db.Close()

	_, sessionID := seedTestRoomAndSession(t, db, "live", 1)

	mockUC := newMockSessionUC()
	handler := newTestChatHandler(t, db, mockUC)

	roomIDStr := strconv.FormatInt(sessionID, 10)

	// Start the auto-end timer
	handler.startAutoEnd(roomIDStr, sessionID)

	// First cancel — should work normally
	handler.cancelAutoEnd(roomIDStr)

	// Second cancel — must NOT panic (idempotent)
	handler.cancelAutoEnd(roomIDStr)

	// Also verify cancelAutoEnd on a room that never had a timer doesn't panic
	handler.cancelAutoEnd("nonexistent-room")

	// Wait past the timer duration
	time.Sleep(300 * time.Millisecond)

	// AutoEnd should NOT have been called
	assert.Empty(t, mockUC.getAutoEndCalls(), "AutoEnd should not be called after cancel")
}

func TestChatHandler_AutoEnd_SkipNonLiveSession(t *testing.T) {
	orig := autoEndTimeUnit
	autoEndTimeUnit = 50 * time.Millisecond
	t.Cleanup(func() { autoEndTimeUnit = orig })

	db := setupTestDB(t)
	defer db.Close()

	_, sessionID := seedTestRoomAndSession(t, db, "ended", 1)

	mockUC := newMockSessionUC()
	handler := newTestChatHandler(t, db, mockUC)

	roomIDStr := strconv.FormatInt(sessionID, 10)

	// Start the auto-end timer — should be a no-op for non-live sessions
	handler.startAutoEnd(roomIDStr, sessionID)

	// Wait past the timer duration
	time.Sleep(300 * time.Millisecond)

	// AutoEnd should NOT have been called
	assert.Empty(t, mockUC.getAutoEndCalls(), "AutoEnd should not be called for non-live sessions")
}

func TestChatHandler_AutoEnd_SkipZeroMinutes(t *testing.T) {
	orig := autoEndTimeUnit
	autoEndTimeUnit = 50 * time.Millisecond
	t.Cleanup(func() { autoEndTimeUnit = orig })

	db := setupTestDB(t)
	defer db.Close()

	_, sessionID := seedTestRoomAndSession(t, db, "live", 0)

	mockUC := newMockSessionUC()
	handler := newTestChatHandler(t, db, mockUC)

	roomIDStr := strconv.FormatInt(sessionID, 10)

	// Start the auto-end timer — should be a no-op when minutes <= 0
	handler.startAutoEnd(roomIDStr, sessionID)

	// Wait past the timer duration
	time.Sleep(300 * time.Millisecond)

	// AutoEnd should NOT have been called
	assert.Empty(t, mockUC.getAutoEndCalls(), "AutoEnd should not be called when minutes is 0")
}

func TestChatHandler_AutoEnd_TriggerOnLastOperatorDisconnect(t *testing.T) {
	orig := autoEndTimeUnit
	autoEndTimeUnit = 50 * time.Millisecond
	t.Cleanup(func() { autoEndTimeUnit = orig })

	db := setupTestDB(t)
	defer db.Close()

	_, sessionID := seedTestRoomAndSession(t, db, "live", 1)
	roomIDStr := strconv.FormatInt(sessionID, 10)

	ch := make(chan int64, 1)
	mockUC := newMockSessionUCWithCh(ch)
	handler := newTestChatHandler(t, db, mockUC)

	// Hub must be running for channel-based Register/Unregister to work
	go handler.hub.Run()

	// Set up Echo test server
	e := echo.New()
	e.GET("/ws/:id", handler.HandleWS)
	server := httptest.NewServer(e)
	defer server.Close()

	// Generate JWT tokens for two operators (different userIDs)
	token1, err := jwt.Generate("test-secret", jwt.Claims{UserID: 100, Email: "op1@test.com", Role: "admin"}, 60)
	require.NoError(t, err)
	token2, err := jwt.Generate("test-secret", jwt.Claims{UserID: 200, Email: "op2@test.com", Role: "operator"}, 60)
	require.NoError(t, err)

	wsBaseURL := "ws" + strings.TrimPrefix(server.URL, "http")
	wsURL := wsBaseURL + "/ws/" + strconv.FormatInt(sessionID, 10)

	// Connect operator 1 (admin role → IsOperator = true)
	conn1, _, err := websocket.DefaultDialer.Dial(wsURL+"?token="+token1, nil)
	require.NoError(t, err)

	// Poll until operator 1 is registered (Register is channel-based, async)
	pollOperatorCount(t, handler, roomIDStr, 1, 2*time.Second)

	// Connect operator 2 (operator role → IsOperator = true)
	conn2, _, err := websocket.DefaultDialer.Dial(wsURL+"?token="+token2, nil)
	require.NoError(t, err)

	pollOperatorCount(t, handler, roomIDStr, 2, 2*time.Second)

	// Disconnect operator 1 — should NOT trigger auto-end (operator 2 still connected)
	conn1.Close()
	pollOperatorCount(t, handler, roomIDStr, 1, 2*time.Second)
	assert.Empty(t, mockUC.getAutoEndCalls(),
		"AutoEnd should NOT be called when another operator remains")

	// Disconnect operator 2 (last operator) — SHOULD trigger auto-end
	conn2.Close()
	pollOperatorCount(t, handler, roomIDStr, 0, 2*time.Second)

	// Wait for AutoEnd timer to fire
	select {
	case calledID := <-ch:
		assert.Equal(t, sessionID, calledID, "AutoEnd should be called with the correct session ID")
	case <-time.After(3 * time.Second):
		t.Fatal("timeout: AutoEnd was not called after last operator disconnected")
	}

	assert.Len(t, mockUC.getAutoEndCalls(), 1, "exactly one AutoEnd should be called")
}

// pollOperatorCount polls GetOperatorConnectionCount until it reaches want or the
// deadline expires. Fails the test via t.Fatal if the count never reaches want.
func pollOperatorCount(t *testing.T, handler *ChatHandler, roomID string, want int, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for {
		if handler.hub.GetOperatorConnectionCount(roomID) == want {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("operator count: got %d, want %d after %v",
				handler.hub.GetOperatorConnectionCount(roomID), want, timeout)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestChatHandler_AutoEnd_TimerReplacement(t *testing.T) {
	orig := autoEndTimeUnit
	autoEndTimeUnit = 50 * time.Millisecond
	t.Cleanup(func() { autoEndTimeUnit = orig })

	db := setupTestDB(t)
	defer db.Close()

	_, sessionID := seedTestRoomAndSession(t, db, "live", 1)
	// Also need a second room+session so each startAutoEnd targets a different session
	_, sessionID2 := seedTestRoomAndSessionWithSlug(t, db, "live", 1, "test-room-2")

	ch := make(chan int64, 2)
	mockUC := newMockSessionUCWithCh(ch)
	handler := newTestChatHandler(t, db, mockUC)

	// Use the SAME roomID so the second startAutoEnd replaces the first timer
	roomIDStr := "shared-room"

	// First startAutoEnd — will be cancelled by the second
	handler.startAutoEnd(roomIDStr, sessionID)

	// Second startAutoEnd — cancels first, starts its own timer
	handler.startAutoEnd(roomIDStr, sessionID2)

	// Wait for AutoEnd to fire
	select {
	case calledID := <-ch:
		assert.Equal(t, sessionID2, calledID, "only the second (replacement) timer should fire")
	case <-time.After(2 * time.Second):
		t.Fatal("timeout: replacement timer did not fire")
	}

	// Verify exactly ONE call (the replaced timer should NOT have fired)
	assert.Len(t, mockUC.getAutoEndCalls(), 1, "only one AutoEnd should be called (the replacement)")
}


