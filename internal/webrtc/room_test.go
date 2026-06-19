package webrtc

import (
	"encoding/json"
	"testing"

	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRoomManager() *RoomManager {
	return NewRoomManager()
}

func TestRoomManager_CreateAndGetRoom(t *testing.T) {
	rm := newTestRoomManager()

	room := rm.GetOrCreateRoom("room-1", 10, 1)
	require.NotNil(t, room)
	assert.Equal(t, "room-1", room.ID)
	assert.Equal(t, 10, room.MaxSize)
	assert.Equal(t, int64(1), room.CreatedBy)

	got := rm.GetRoom("room-1")
	require.NotNil(t, got)
	assert.Equal(t, room.ID, got.ID)
}

func TestRoomManager_GetOrCreateRoom_Idempotent(t *testing.T) {
	rm := newTestRoomManager()

	r1 := rm.GetOrCreateRoom("room-x", 5, 1)
	r2 := rm.GetOrCreateRoom("room-x", 50, 2)
	assert.Equal(t, r1, r2, "same room pointer returned")
	assert.Equal(t, int64(1), r2.CreatedBy, "original creator preserved")
}

func TestRoomManager_GetRoom_NotFound(t *testing.T) {
	rm := newTestRoomManager()
	assert.Nil(t, rm.GetRoom("nonexistent"))
}

func TestRoomManager_DeleteRoom(t *testing.T) {
	rm := newTestRoomManager()
	rm.GetOrCreateRoom("room-del", 10, 1)
	rm.DeleteRoom("room-del")
	assert.Nil(t, rm.GetRoom("room-del"))
}

func TestRoomManager_AddParticipant(t *testing.T) {
	rm := newTestRoomManager()
	rm.GetOrCreateRoom("room-ap", 10, 1)

	err := rm.AddParticipant("room-ap", &Participant{
		ID:   "user-1",
		Name: "Ali",
		Role: "teacher",
	})
	require.NoError(t, err)

	p := rm.GetParticipant("room-ap", "user-1")
	require.NotNil(t, p)
	assert.Equal(t, "Ali", p.Name)
	assert.Equal(t, "teacher", p.Role)
}

func TestRoomManager_AddParticipant_RolePreserved(t *testing.T) {
	rm := newTestRoomManager()
	rm.GetOrCreateRoom("room-role", 10, 1)

	roles := []struct {
		id   string
		name string
		role string
	}{
		{"u1", "Owner", "owner"},
		{"u2", "Admin", "admin"},
		{"u3", "Teacher", "teacher"},
		{"u4", "Student", "student"},
		{"u5", "Presenter", "presenter"},
	}

	for _, r := range roles {
		err := rm.AddParticipant("room-role", &Participant{
			ID:   r.id,
			Name: r.name,
			Role: r.role,
		})
		require.NoError(t, err)
	}

	for _, r := range roles {
		p := rm.GetParticipant("room-role", r.id)
		require.NotNil(t, p, "participant %s should exist", r.id)
		assert.Equal(t, r.role, p.Role, "role should be preserved for %s", r.name)
	}
}

func TestRoomManager_GetRoomParticipants_RoleIncluded(t *testing.T) {
	rm := newTestRoomManager()
	rm.GetOrCreateRoom("room-pi", 10, 1)

	rm.AddParticipant("room-pi", &Participant{ID: "u1", Name: "Admin User", Role: "admin", IsMuted: false, IsVideoOff: false})
	rm.AddParticipant("room-pi", &Participant{ID: "u2", Name: "Student User", Role: "student", IsMuted: true, IsVideoOff: true})

	participants := rm.GetRoomParticipants("room-pi")
	require.Len(t, participants, 2)

	roleMap := make(map[string]ParticipantInfo)
	for _, p := range participants {
		roleMap[p.ID] = p
	}

	p1 := roleMap["u1"]
	assert.Equal(t, "Admin User", p1.Name)
	assert.Equal(t, "admin", p1.Role, "role must be included in ParticipantInfo")
	assert.False(t, p1.IsMuted)
	assert.False(t, p1.IsVideoOff)

	p2 := roleMap["u2"]
	assert.Equal(t, "Student User", p2.Name)
	assert.Equal(t, "student", p2.Role)
	assert.True(t, p2.IsMuted)
	assert.True(t, p2.IsVideoOff)
}

func TestRoomManager_ParticipantInfo_JSON_RoleField(t *testing.T) {
	rm := newTestRoomManager()
	rm.GetOrCreateRoom("room-json", 10, 1)

	rm.AddParticipant("room-json", &Participant{
		ID: "u1", Name: "Test", Role: "teacher",
		IsMuted: true, IsVideoOff: false, IsScreenSharing: true,
	})

	participants := rm.GetRoomParticipants("room-json")
	require.Len(t, participants, 1)

	data, err := json.Marshal(participants[0])
	require.NoError(t, err)

	var parsed map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &parsed))

	assert.Equal(t, "teacher", parsed["role"], "role field must appear in JSON output")
	assert.Equal(t, "u1", parsed["id"])
	assert.Equal(t, "Test", parsed["name"])
	assert.Equal(t, true, parsed["is_muted"])
	assert.Equal(t, false, parsed["is_video_off"])
	assert.Equal(t, true, parsed["is_screen_sharing"])
}

func TestRoomManager_GetRoomStats_RoleIncluded(t *testing.T) {
	rm := newTestRoomManager()
	rm.GetOrCreateRoom("room-stats", 10, 1)

	rm.AddParticipant("room-stats", &Participant{ID: "u1", Name: "A", Role: "admin"})
	rm.AddParticipant("room-stats", &Participant{ID: "u2", Name: "B", Role: "student"})

	stats := rm.GetRoomStats("room-stats")
	require.NotNil(t, stats)
	assert.Equal(t, "room-stats", stats.RoomID)
	assert.Equal(t, 2, stats.ParticipantCount)
	require.Len(t, stats.Participants, 2)

	for _, p := range stats.Participants {
		assert.NotEmpty(t, p.Role, "role must not be empty in stats")
	}
}

func TestRoomManager_AddParticipant_RoomNotFound(t *testing.T) {
	rm := newTestRoomManager()
	err := rm.AddParticipant("no-room", &Participant{ID: "u1", Name: "X"})
	assert.ErrorIs(t, err, ErrRoomNotFound)
}

func TestRoomManager_AddParticipant_RoomFull(t *testing.T) {
	rm := newTestRoomManager()
	rm.GetOrCreateRoom("room-full", 2, 1)

	require.NoError(t, rm.AddParticipant("room-full", &Participant{ID: "u1", Name: "A"}))
	require.NoError(t, rm.AddParticipant("room-full", &Participant{ID: "u2", Name: "B"}))
	err := rm.AddParticipant("room-full", &Participant{ID: "u3", Name: "C"})
	assert.ErrorIs(t, err, ErrRoomFull)
}

func TestRoomManager_RemoveParticipant(t *testing.T) {
	rm := newTestRoomManager()
	rm.GetOrCreateRoom("room-rm", 10, 1)

	rm.AddParticipant("room-rm", &Participant{ID: "u1", Name: "A"})
	rm.RemoveParticipant("room-rm", "u1")

	assert.Nil(t, rm.GetParticipant("room-rm", "u1"))
}

func TestRoomManager_RemoveParticipant_LastCleansUpRoom(t *testing.T) {
	rm := newTestRoomManager()
	rm.GetOrCreateRoom("room-last", 10, 1)

	rm.AddParticipant("room-last", &Participant{ID: "u1", Name: "A"})
	rm.RemoveParticipant("room-last", "u1")

	// Give the goroutine a moment to clean up
	assert.Eventually(t, func() bool {
		return rm.GetRoom("room-last") == nil
	}, time.Second, 10*time.Millisecond, "room should be cleaned up after last participant leaves")
}

func TestRoomManager_GetParticipants_EmptyRoom(t *testing.T) {
	rm := newTestRoomManager()
	rm.GetOrCreateRoom("room-empty", 10, 1)

	participants := rm.GetRoomParticipants("room-empty")
	assert.Nil(t, participants, "empty room should return nil")
}

func TestRoomManager_GetRoomParticipants_NonexistentRoom(t *testing.T) {
	rm := newTestRoomManager()
	participants := rm.GetRoomParticipants("no-room")
	assert.Nil(t, participants)
}

func TestRoomManager_GetRoomStats_NonexistentRoom(t *testing.T) {
	rm := newTestRoomManager()
	stats := rm.GetRoomStats("no-room")
	assert.Nil(t, stats)
}

func TestRoomManager_ParticipantFields_AllPreserved(t *testing.T) {
	rm := newTestRoomManager()
	rm.GetOrCreateRoom("room-all", 10, 1)

	p := &Participant{
		ID:              "u-full",
		Name:            "Full User",
		Role:            "presenter",
		IsMuted:         true,
		IsVideoOff:      true,
		IsScreenSharing: true,
	}
	require.NoError(t, rm.AddParticipant("room-all", p))

	participants := rm.GetRoomParticipants("room-all")
	require.Len(t, participants, 1)

	pi := participants[0]
	assert.Equal(t, "u-full", pi.ID)
	assert.Equal(t, "Full User", pi.Name)
	assert.Equal(t, "presenter", pi.Role)
	assert.True(t, pi.IsMuted)
	assert.True(t, pi.IsVideoOff)
	assert.True(t, pi.IsScreenSharing)
}
