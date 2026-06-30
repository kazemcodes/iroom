package usecase

import (
	"testing"

	"github.com/iroom/iroom/internal/domain/entity"
	"github.com/iroom/iroom/internal/pkg/slug"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func strPtr(s string) *string { return &s }
func boolPtr(b bool) *bool    { return &b }
func intPtr(i int) *int       { return &i }

func TestGenerateRoomSlug(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple english name",
			input:    "Math Class",
			expected: "math-class",
		},
		{
			name:     "persian digits converted, persian letters kept",
			input:    "ریاضی ۱۰۱",
			expected: "ریاضی-101",
		},
		{
			name:     "mixed persian digits converted",
			input:    "کلاس ۲۰۲۴",
			expected: "کلاس-2024",
		},
		{
			name:     "special characters removed",
			input:    "Hello @World! #1",
			expected: "hello-world-1",
		},
		{
			name:     "multiple spaces collapse",
			input:    "my   room   name",
			expected: "my-room-name",
		},
		{
			name:     "leading/trailing dashes trimmed",
			input:    " my room ",
			expected: "my-room",
		},
		{
			name:     "zero-width non-joiner removed",
			input:    "اتاق‌test",
			expected: "اتاقtest",
		},
		{
			name:     "all special chars yields fallback",
			input:    "@#$%^&*()",
			expected: "",
		},
		{
			name:     "single letter",
			input:    "a",
			expected: "a",
		},
		{
			name:     "empty string gets fallback",
			input:    "",
			expected: "",
		},
		{
			name:     "persian digits full range",
			input:    "۰۱۲۳۴۵۶۷۸۹",
			expected: "0123456789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := slug.Generate(tt.input)
			if tt.input == "" || tt.input == "@#$%^&*()" {
				assert.Regexp(t, `^room-\d+$`, result, "fallback slug should match pattern")
				return
			}
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateRoomSlugDoubleDashCollapse(t *testing.T) {
	s := slug.Generate("a  b")
	assert.NotContains(t, s, "--", "double dashes should be collapsed")
}

func TestRoomUseCase_Create_DefaultGuestLogin(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	roomRepo := newTestRoomRepo(t, db)
	userRepo := newTestUserRepo(t, db)
	sessionRepo := newTestSessionRepo(t, db)

	uc := NewRoomUseCase(roomRepo, userRepo, sessionRepo)

	room, err := uc.Create(1, "Test Room", "Description", "#FF0000", 30, "")
	require.NoError(t, err)
	assert.Equal(t, "Test Room", room.Name)
	assert.Equal(t, "Description", room.Description)
	assert.Equal(t, "#FF0000", room.Color)
	assert.Equal(t, "test-room", room.Slug)
	assert.True(t, room.GuestLoginEnabled)
	assert.NotZero(t, room.ID)
}

func TestRoomUseCase_Create_EmptyName(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	roomRepo := newTestRoomRepo(t, db)
	userRepo := newTestUserRepo(t, db)
	sessionRepo := newTestSessionRepo(t, db)

	uc := NewRoomUseCase(roomRepo, userRepo, sessionRepo)

	_, err := uc.Create(1, "", "Description", "#FF0000", 0, "")
	require.Error(t, err, "empty name must be rejected")
}

func TestRoomUseCase_GetByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	roomRepo := newTestRoomRepo(t, db)
	userRepo := newTestUserRepo(t, db)
	sessionRepo := newTestSessionRepo(t, db)

	uc := NewRoomUseCase(roomRepo, userRepo, sessionRepo)

	created, err := uc.Create(1, "My Room", "Desc", "#00FF00", 50, "")
	require.NoError(t, err)

	fetched, err := uc.GetByID(created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, fetched.ID)
	assert.Equal(t, "My Room", fetched.Name)
}

func TestRoomUseCase_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	roomRepo := newTestRoomRepo(t, db)
	userRepo := newTestUserRepo(t, db)
	sessionRepo := newTestSessionRepo(t, db)

	uc := NewRoomUseCase(roomRepo, userRepo, sessionRepo)

	_, err := uc.GetByID(9999)
	assert.Error(t, err)
}

func TestRoomUseCase_GetBySlug(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	roomRepo := newTestRoomRepo(t, db)
	userRepo := newTestUserRepo(t, db)
	sessionRepo := newTestSessionRepo(t, db)

	uc := NewRoomUseCase(roomRepo, userRepo, sessionRepo)

	created, err := uc.Create(1, "Slug Room", "Desc", "#0000FF", 50, "")
	require.NoError(t, err)

	fetched, err := uc.GetBySlug(created.Slug)
	require.NoError(t, err)
	assert.Equal(t, created.ID, fetched.ID)
}

func TestRoomUseCase_List(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	roomRepo := newTestRoomRepo(t, db)
	userRepo := newTestUserRepo(t, db)
	sessionRepo := newTestSessionRepo(t, db)

	uc := NewRoomUseCase(roomRepo, userRepo, sessionRepo)

	for i := 0; i < 5; i++ {
		_, err := uc.Create(1, "Room "+string(rune('A'+i)), "Desc", "#FF0000", 50, "")
		require.NoError(t, err)
	}

	rooms, total, err := uc.List(1, 10, "")
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, rooms, 5)
}

func TestRoomUseCase_List_WithSearch(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	roomRepo := newTestRoomRepo(t, db)
	userRepo := newTestUserRepo(t, db)
	sessionRepo := newTestSessionRepo(t, db)

	uc := NewRoomUseCase(roomRepo, userRepo, sessionRepo)

	_, _ = uc.Create(1, "Math Class", "Algebra", "#FF0000", 50, "")
	_, _ = uc.Create(1, "Science Lab", "Physics", "#00FF00", 50, "")

	rooms, total, err := uc.List(1, 10, "Math")
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, rooms, 1)
	assert.Equal(t, "Math Class", rooms[0].Name)
}

func TestRoomUseCase_Update(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	roomRepo := newTestRoomRepo(t, db)
	userRepo := newTestUserRepo(t, db)
	sessionRepo := newTestSessionRepo(t, db)

	uc := NewRoomUseCase(roomRepo, userRepo, sessionRepo)

	created, err := uc.Create(1, "Original", "Old Desc", "#FF0000", 50, "")
	require.NoError(t, err)

	updated, err := uc.Update(created.ID, 1, "admin", RoomUpdate{Name: strPtr("Updated"), Description: strPtr("New Desc"), Color: strPtr("#00FF00"), GuestLoginEnabled: boolPtr(false), MaxUsers: intPtr(100)})
	require.NoError(t, err)
	assert.Equal(t, "Updated", updated.Name)
	assert.Equal(t, "New Desc", updated.Description)
	assert.Equal(t, "#00FF00", updated.Color)
	assert.False(t, updated.GuestLoginEnabled)
	assert.Equal(t, 100, updated.MaxUsers)
}

func TestRoomUseCase_Update_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	roomRepo := newTestRoomRepo(t, db)
	userRepo := newTestUserRepo(t, db)
	sessionRepo := newTestSessionRepo(t, db)

	uc := NewRoomUseCase(roomRepo, userRepo, sessionRepo)

	_, err := uc.Update(9999, 1, "admin", RoomUpdate{Name: strPtr("X"), Description: strPtr("X"), Color: strPtr("X"), GuestLoginEnabled: boolPtr(true)})
	assert.Error(t, err)
}

func TestRoomUseCase_Update_Forbidden(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	roomRepo := newTestRoomRepo(t, db)
	userRepo := newTestUserRepo(t, db)
	sessionRepo := newTestSessionRepo(t, db)

	uc := NewRoomUseCase(roomRepo, userRepo, sessionRepo)

	created, err := uc.Create(1, "Original", "Desc", "#FF0000", 50, "")
	require.NoError(t, err)

	_, err = uc.Update(created.ID, 999, "student", RoomUpdate{Name: strPtr("Hacked"), GuestLoginEnabled: boolPtr(true)})
	assert.Error(t, err)
}

func TestRoomUseCase_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	roomRepo := newTestRoomRepo(t, db)
	userRepo := newTestUserRepo(t, db)
	sessionRepo := newTestSessionRepo(t, db)

	uc := NewRoomUseCase(roomRepo, userRepo, sessionRepo)

	created, err := uc.Create(1, "To Delete", "Desc", "#FF0000", 50, "")
	require.NoError(t, err)

	err = uc.Delete(created.ID, 1, "admin")
	require.NoError(t, err)

	_, err = uc.GetByID(created.ID)
	assert.Error(t, err)
}

func TestRoomUseCase_Delete_Forbidden(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	roomRepo := newTestRoomRepo(t, db)
	userRepo := newTestUserRepo(t, db)
	sessionRepo := newTestSessionRepo(t, db)

	uc := NewRoomUseCase(roomRepo, userRepo, sessionRepo)

	created, err := uc.Create(1, "Protected", "Desc", "#FF0000", 50, "")
	require.NoError(t, err)

	err = uc.Delete(created.ID, 999, "student")
	assert.Error(t, err)
}

func TestRoomUseCase_AddUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	roomRepo := newTestRoomRepo(t, db)
	userRepo := newTestUserRepo(t, db)
	sessionRepo := newTestSessionRepo(t, db)

	uc := NewRoomUseCase(roomRepo, userRepo, sessionRepo)

	room, _ := uc.Create(1, "Room", "Desc", "#FF0000", 50, "")
	err := uc.AddUser(room.ID, 2, 1, "student", 1, "admin")
	require.NoError(t, err)

	count, err := uc.GetUserCount(room.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestRoomUseCase_AddUser_DefaultRole(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	roomRepo := newTestRoomRepo(t, db)
	userRepo := newTestUserRepo(t, db)
	sessionRepo := newTestSessionRepo(t, db)

	uc := NewRoomUseCase(roomRepo, userRepo, sessionRepo)

	room, _ := uc.Create(1, "Room", "Desc", "#FF0000", 50, "")
	err := uc.AddUser(room.ID, 3, 1, "", 1, "admin")
	require.NoError(t, err)
}

func TestRoomUseCase_RemoveUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	roomRepo := newTestRoomRepo(t, db)
	userRepo := newTestUserRepo(t, db)
	sessionRepo := newTestSessionRepo(t, db)

	uc := NewRoomUseCase(roomRepo, userRepo, sessionRepo)

	room, _ := uc.Create(1, "Room", "Desc", "#FF0000", 50, "")
	_ = uc.AddUser(room.ID, 4, 1, "student", 1, "admin")

	err := uc.RemoveUser(room.ID, 4, 1, "admin")
	require.NoError(t, err)

	count, err := uc.GetUserCount(room.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestRoomUseCase_RemoveUser_Forbidden(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	roomRepo := newTestRoomRepo(t, db)
	userRepo := newTestUserRepo(t, db)
	sessionRepo := newTestSessionRepo(t, db)

	uc := NewRoomUseCase(roomRepo, userRepo, sessionRepo)

	room, _ := uc.Create(1, "Room", "Desc", "#FF0000", 50, "")
	_ = uc.AddUser(room.ID, 4, 1, "student", 1, "admin")

	err := uc.RemoveUser(room.ID, 4, 999, "student")
	assert.Error(t, err)
}

func TestRoomUseCase_GetUsers_Empty(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	roomRepo := newTestRoomRepo(t, db)
	userRepo := newTestUserRepo(t, db)
	sessionRepo := newTestSessionRepo(t, db)

	uc := NewRoomUseCase(roomRepo, userRepo, sessionRepo)

	room, _ := uc.Create(1, "Room", "Desc", "#FF0000", 50, "")

	users, err := uc.GetUsers(room.ID)
	require.NoError(t, err)
	assert.Empty(t, users)
	assert.NotNil(t, users, "should return empty slice, not nil")
}

func TestRoomUseCase_Settings(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	roomRepo := newTestRoomRepo(t, db)
	userRepo := newTestUserRepo(t, db)
	sessionRepo := newTestSessionRepo(t, db)

	uc := NewRoomUseCase(roomRepo, userRepo, sessionRepo)

	room, _ := uc.Create(1, "Room", "Desc", "#FF0000", 50, "")

	settings, err := uc.GetSettings(room.ID)
	require.NoError(t, err)
	assert.Equal(t, 50, settings.MaxUsers)
	assert.True(t, settings.RecordingEnabled)
	assert.True(t, settings.AllowStudentAudio)
	assert.True(t, settings.AllowStudentChat)
	assert.Equal(t, 120, settings.SessionAutoEndMinutes)

	settings.MaxUsers = 100
	settings.AllowStudentVideo = true
	err = uc.UpdateSettings(room.ID, 1, "admin", settings)
	require.NoError(t, err)

	fetched, err := uc.GetSettings(room.ID)
	require.NoError(t, err)
	assert.Equal(t, 100, fetched.MaxUsers)
	assert.True(t, fetched.AllowStudentVideo)
}

func TestRoomUseCase_Create_OwnerID_Preserved(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	roomRepo := newTestRoomRepo(t, db)
	userRepo := newTestUserRepo(t, db)
	sessionRepo := newTestSessionRepo(t, db)
	uc := NewRoomUseCase(roomRepo, userRepo, sessionRepo)

	room, err := uc.Create(42, "Owner Room", "Desc", "#000", 50, "")
	require.NoError(t, err)
	assert.Equal(t, int64(42), room.OwnerID)

	fetched, err := uc.GetByID(room.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(42), fetched.OwnerID, "owner_id must survive create → get")
}

func TestRoomUseCase_Update_PreservesUntouchedFields(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	roomRepo := newTestRoomRepo(t, db)
	userRepo := newTestUserRepo(t, db)
	sessionRepo := newTestSessionRepo(t, db)
	uc := NewRoomUseCase(roomRepo, userRepo, sessionRepo)

	room, _ := uc.Create(5, "Original", "Original Desc", "#FF0000", 50, "")
	assert.True(t, room.GuestLoginEnabled)

	_, err := uc.Update(room.ID, 5, "admin", RoomUpdate{Name: strPtr("New Name"), GuestLoginEnabled: boolPtr(true)})
	require.NoError(t, err)

	fetched, _ := uc.GetByID(room.ID)
	assert.Equal(t, "New Name", fetched.Name)
	assert.Equal(t, "Original Desc", fetched.Description, "description should be preserved when empty string passed")
	assert.Equal(t, "#FF0000", fetched.Color, "color should be preserved when empty string passed")
	assert.True(t, fetched.GuestLoginEnabled)
}

func TestRoomUseCase_Update_ChangesGuestLogin(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	roomRepo := newTestRoomRepo(t, db)
	userRepo := newTestUserRepo(t, db)
	sessionRepo := newTestSessionRepo(t, db)
	uc := NewRoomUseCase(roomRepo, userRepo, sessionRepo)

	room, _ := uc.Create(1, "R", "D", "#000", 50, "")
	assert.True(t, room.GuestLoginEnabled)

	_, err := uc.Update(room.ID, 1, "admin", RoomUpdate{GuestLoginEnabled: boolPtr(false)})
	require.NoError(t, err)

	fetched, _ := uc.GetByID(room.ID)
	assert.False(t, fetched.GuestLoginEnabled, "guest_login_enabled should change to false")
}

func TestRoomUseCase_AddUser_TeacherRole_Preserved(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	roomRepo := newTestRoomRepo(t, db)
	userRepo := newTestUserRepo(t, db)
	sessionRepo := newTestSessionRepo(t, db)
	uc := NewRoomUseCase(roomRepo, userRepo, sessionRepo)

	room, _ := uc.Create(1, "R", "D", "#000", 50, "")

	u := &entity.User{Email: "teacher@test.com", DisplayName: "T", Role: "teacher"}
	require.NoError(t, userRepo.Create(u))

	err := uc.AddUser(room.ID, u.ID, 1, "teacher", 1, "admin")
	require.NoError(t, err)

	users, err := uc.GetUsers(room.ID)
	require.NoError(t, err)
	require.Len(t, users, 1)
	assert.Equal(t, "teacher", string(users[0].Role), "teacher role must survive AddUser → GetUsers")
}

func TestRoomUseCase_AddUser_StudentRole_Preserved(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	roomRepo := newTestRoomRepo(t, db)
	userRepo := newTestUserRepo(t, db)
	sessionRepo := newTestSessionRepo(t, db)
	uc := NewRoomUseCase(roomRepo, userRepo, sessionRepo)

	room, _ := uc.Create(1, "R", "D", "#000", 50, "")

	u := &entity.User{Email: "student@test.com", DisplayName: "S", Role: "student"}
	require.NoError(t, userRepo.Create(u))

	err := uc.AddUser(room.ID, u.ID, 1, "student", 1, "admin")
	require.NoError(t, err)

	users, err := uc.GetUsers(room.ID)
	require.NoError(t, err)
	require.Len(t, users, 1)
	assert.Equal(t, "student", string(users[0].Role))
}

func TestRoomUseCase_GetUsers_ReturnsAllFields(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	roomRepo := newTestRoomRepo(t, db)
	userRepo := newTestUserRepo(t, db)
	sessionRepo := newTestSessionRepo(t, db)
	uc := NewRoomUseCase(roomRepo, userRepo, sessionRepo)

	room, _ := uc.Create(1, "R", "D", "#000", 50, "")

	u := &entity.User{Email: "test@test.com", DisplayName: "Test User", Role: "teacher", IsActive: true}
	require.NoError(t, userRepo.Create(u))

	_ = uc.AddUser(room.ID, u.ID, 1, "teacher", 1, "admin")

	users, err := uc.GetUsers(room.ID)
	require.NoError(t, err)
	require.Len(t, users, 1)

	user := users[0]
	assert.Equal(t, "test@test.com", user.Email, "email must be returned")
	assert.Equal(t, "Test User", user.DisplayName, "display_name must be returned")
	assert.Equal(t, "teacher", string(user.Role), "role must be returned")
	assert.True(t, user.IsActive, "is_active must be returned")
}

func TestRoomUseCase_Settings_AllFieldsRoundtrip(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	roomRepo := newTestRoomRepo(t, db)
	userRepo := newTestUserRepo(t, db)
	sessionRepo := newTestSessionRepo(t, db)
	uc := NewRoomUseCase(roomRepo, userRepo, sessionRepo)

	room, _ := uc.Create(1, "R", "D", "#000", 50, "")

	s, _ := uc.GetSettings(room.ID)
	s.MaxUsers = 200
	s.RecordingEnabled = false
	s.AllowStudentVideo = true
	s.AllowStudentAudio = false
	s.AllowStudentScreenShare = true
	s.AllowStudentWhiteboard = true
	s.AllowStudentChat = false
	s.SessionAutoEndMinutes = 30
	s.WaitingRoomEnabled = true
	require.NoError(t, uc.UpdateSettings(room.ID, 1, "admin", s))

	fetched, _ := uc.GetSettings(room.ID)
	assert.Equal(t, 200, fetched.MaxUsers)
	assert.False(t, fetched.RecordingEnabled)
	assert.True(t, fetched.AllowStudentVideo)
	assert.False(t, fetched.AllowStudentAudio)
	assert.True(t, fetched.AllowStudentScreenShare)
	assert.True(t, fetched.AllowStudentWhiteboard)
	assert.False(t, fetched.AllowStudentChat)
	assert.Equal(t, 30, fetched.SessionAutoEndMinutes)
	assert.True(t, fetched.WaitingRoomEnabled)
}

func TestRoomUseCase_List_ReturnsAllFields(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	roomRepo := newTestRoomRepo(t, db)
	userRepo := newTestUserRepo(t, db)
	sessionRepo := newTestSessionRepo(t, db)
	uc := NewRoomUseCase(roomRepo, userRepo, sessionRepo)

	created, _ := uc.Create(7, "Field Room", "A description", "#123456", 50, "")

	rooms, _, err := uc.List(1, 10, "")
	require.NoError(t, err)
	require.Len(t, rooms, 1)

	fetched := rooms[0]
	assert.Equal(t, created.ID, fetched.ID)
	assert.Equal(t, int64(7), fetched.OwnerID, "owner_id must be in list result")
	assert.Equal(t, "Field Room", fetched.Name)
	assert.Equal(t, "A description", fetched.Description)
	assert.Equal(t, "#123456", fetched.Color)
	assert.Equal(t, "field-room", fetched.Slug)
	assert.True(t, fetched.GuestLoginEnabled)
}

func TestRoomUseCase_GetBySlug_ReturnsAllFields(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	roomRepo := newTestRoomRepo(t, db)
	userRepo := newTestUserRepo(t, db)
	sessionRepo := newTestSessionRepo(t, db)
	uc := NewRoomUseCase(roomRepo, userRepo, sessionRepo)

	created, _ := uc.Create(3, "Slug Full", "Desc", "#ABC", 50, "")

	fetched, err := uc.GetBySlug(created.Slug)
	require.NoError(t, err)
	assert.Equal(t, int64(3), fetched.OwnerID)
	assert.Equal(t, "Slug Full", fetched.Name)
	assert.Equal(t, "Desc", fetched.Description)
	assert.Equal(t, "#ABC", fetched.Color)
	assert.True(t, fetched.GuestLoginEnabled)
}

func TestRoomUseCase_Delete_RemovesRoom(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	roomRepo := newTestRoomRepo(t, db)
	userRepo := newTestUserRepo(t, db)
	sessionRepo := newTestSessionRepo(t, db)
	uc := NewRoomUseCase(roomRepo, userRepo, sessionRepo)

	room, _ := uc.Create(1, "Delete Me", "Desc", "#000", 50, "")
	require.NoError(t, uc.Delete(room.ID, 1, "admin"))

	_, err := uc.GetByID(room.ID)
	assert.Error(t, err)

	_, err = uc.GetBySlug(room.Slug)
	assert.Error(t, err)
}

func TestRoomUseCase_RegenerateCode(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	roomRepo := newTestRoomRepo(t, db)
	userRepo := newTestUserRepo(t, db)
	sessionRepo := newTestSessionRepo(t, db)
	uc := NewRoomUseCase(roomRepo, userRepo, sessionRepo)

	room, _ := uc.Create(1, "R", "D", "#000", 50, "")

	code, err := uc.RegenerateCode(room.ID, 1, "admin")
	require.NoError(t, err)
	assert.NotEmpty(t, code)

	// Verify the room can be found by the new code
	found, err := uc.JoinByCode(code)
	require.NoError(t, err)
	assert.Equal(t, room.ID, found.ID)
}

func TestRoomUseCase_RegenerateCode_Forbidden(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	roomRepo := newTestRoomRepo(t, db)
	userRepo := newTestUserRepo(t, db)
	sessionRepo := newTestSessionRepo(t, db)
	uc := NewRoomUseCase(roomRepo, userRepo, sessionRepo)

	room, _ := uc.Create(1, "R", "D", "#000", 50, "")

	_, err := uc.RegenerateCode(room.ID, 999, "student")
	assert.Error(t, err)
}

func TestRoomUseCase_GetUserRooms(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	roomRepo := newTestRoomRepo(t, db)
	userRepo := newTestUserRepo(t, db)
	sessionRepo := newTestSessionRepo(t, db)
	uc := NewRoomUseCase(roomRepo, userRepo, sessionRepo)

	_, _ = uc.Create(1, "Room A", "D", "#000", 50, "")
	_, _ = uc.Create(1, "Room B", "D", "#000", 50, "")
	_, _ = uc.Create(2, "Other Room", "D", "#000", 50, "")

	rooms, err := uc.GetUserRooms(1)
	require.NoError(t, err)
	assert.Len(t, rooms, 2)
}

// ---------- New tests covering hardened room feature ----------

func TestRoomUseCase_Update_SlugIsImmutable(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	roomRepo := newTestRoomRepo(t, db)
	userRepo := newTestUserRepo(t, db)
	sessionRepo := newTestSessionRepo(t, db)
	uc := NewRoomUseCase(roomRepo, userRepo, sessionRepo)

	room, err := uc.Create(1, "Original Name", "", "", 50, "")
	require.NoError(t, err)
	originalSlug := room.Slug

	updated, err := uc.Update(room.ID, 1, "admin", RoomUpdate{Name: strPtr("Brand New Name")})
	require.NoError(t, err)
	assert.Equal(t, "Brand New Name", updated.Name)
	assert.Equal(t, originalSlug, updated.Slug, "renaming a room must NOT change its slug — would break links")
}

func TestRoomUseCase_Update_ClearsDescriptionAndColor(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	roomRepo := newTestRoomRepo(t, db)
	userRepo := newTestUserRepo(t, db)
	sessionRepo := newTestSessionRepo(t, db)
	uc := NewRoomUseCase(roomRepo, userRepo, sessionRepo)

	room, _ := uc.Create(1, "R", "old desc", "#FF0000", 50, "")

	empty := ""
	_, err := uc.Update(room.ID, 1, "admin", RoomUpdate{Description: &empty, Color: &empty})
	require.NoError(t, err)

	fetched, _ := uc.GetByID(room.ID)
	assert.Equal(t, "", fetched.Description, "explicit empty must clear description")
	assert.Equal(t, "", fetched.Color, "explicit empty must clear color")
}

func TestRoomUseCase_Update_EmptyNameRejected(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	roomRepo := newTestRoomRepo(t, db)
	userRepo := newTestUserRepo(t, db)
	sessionRepo := newTestSessionRepo(t, db)
	uc := NewRoomUseCase(roomRepo, userRepo, sessionRepo)

	room, _ := uc.Create(1, "R", "", "", 50, "")
	_, err := uc.Update(room.ID, 1, "admin", RoomUpdate{Name: strPtr("")})
	assert.Error(t, err)
}

func TestRoomUseCase_Update_MaxUsersBounds(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	roomRepo := newTestRoomRepo(t, db)
	userRepo := newTestUserRepo(t, db)
	sessionRepo := newTestSessionRepo(t, db)
	uc := NewRoomUseCase(roomRepo, userRepo, sessionRepo)

	room, _ := uc.Create(1, "R", "", "", 50, "")

	for _, bad := range []int{0, -1, 1001, 99999} {
		_, err := uc.Update(room.ID, 1, "admin", RoomUpdate{MaxUsers: intPtr(bad)})
		assert.Error(t, err, "max_users=%d must be rejected", bad)
	}
	_, err := uc.Update(room.ID, 1, "admin", RoomUpdate{MaxUsers: intPtr(100)})
	assert.NoError(t, err)
}

func TestRoomUseCase_Create_DefaultsMaxUsers(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	roomRepo := newTestRoomRepo(t, db)
	userRepo := newTestUserRepo(t, db)
	sessionRepo := newTestSessionRepo(t, db)
	uc := NewRoomUseCase(roomRepo, userRepo, sessionRepo)

	room, err := uc.Create(1, "R", "", "", 0, "")
	require.NoError(t, err)
	assert.Equal(t, 50, room.MaxUsers, "MaxUsers=0 must be defaulted to 50")
}

func TestRoomUseCase_Create_CapsMaxUsers(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	roomRepo := newTestRoomRepo(t, db)
	userRepo := newTestUserRepo(t, db)
	sessionRepo := newTestSessionRepo(t, db)
	uc := NewRoomUseCase(roomRepo, userRepo, sessionRepo)

	room, err := uc.Create(1, "R", "", "", 99999, "")
	require.NoError(t, err)
	assert.Equal(t, 1000, room.MaxUsers, "MaxUsers must be capped at 1000")
}

func TestRoomUseCase_AddUser_RejectsInvalidRole(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	roomRepo := newTestRoomRepo(t, db)
	userRepo := newTestUserRepo(t, db)
	sessionRepo := newTestSessionRepo(t, db)
	uc := NewRoomUseCase(roomRepo, userRepo, sessionRepo)

	room, _ := uc.Create(1, "R", "", "", 50, "")
	err := uc.AddUser(room.ID, 2, 1, "hacker", 1, "admin")
	assert.Error(t, err, "invalid role must be rejected")
}

func TestRoomUseCase_AddUser_RejectsInvalidAccess(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	roomRepo := newTestRoomRepo(t, db)
	userRepo := newTestUserRepo(t, db)
	sessionRepo := newTestSessionRepo(t, db)
	uc := NewRoomUseCase(roomRepo, userRepo, sessionRepo)

	room, _ := uc.Create(1, "R", "", "", 50, "")
	for _, bad := range []int{-1, 4, 99} {
		err := uc.AddUser(room.ID, 2, 1, "student", bad, "admin")
		assert.Error(t, err, "access=%d must be rejected", bad)
	}
}

func TestRoomUseCase_AddUser_EnforcesMaxUsers(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	roomRepo := newTestRoomRepo(t, db)
	userRepo := newTestUserRepo(t, db)
	sessionRepo := newTestSessionRepo(t, db)
	uc := NewRoomUseCase(roomRepo, userRepo, sessionRepo)

	room, _ := uc.Create(1, "R", "", "", 2, "")
	require.NoError(t, uc.AddUser(room.ID, 10, 1, "student", 1, "admin"))
	require.NoError(t, uc.AddUser(room.ID, 11, 1, "student", 1, "admin"))

	err := uc.AddUser(room.ID, 12, 1, "student", 1, "admin")
	assert.Error(t, err, "3rd user must be rejected when max_users=2")

	// existing user re-add (role bump) must still work
	require.NoError(t, uc.AddUser(room.ID, 10, 1, "teacher", 1, "admin"))
}

func TestRoomUseCase_UpdateUserAccess_RejectsInvalid(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	roomRepo := newTestRoomRepo(t, db)
	userRepo := newTestUserRepo(t, db)
	sessionRepo := newTestSessionRepo(t, db)
	uc := NewRoomUseCase(roomRepo, userRepo, sessionRepo)

	room, _ := uc.Create(1, "R", "", "", 50, "")
	require.NoError(t, uc.AddUser(room.ID, 5, 1, "student", 1, "admin"))

	for _, bad := range []int{0, -1, 4, 99} {
		err := uc.UpdateUserAccess(room.ID, 5, 1, "admin", bad)
		assert.Error(t, err, "access=%d must be rejected", bad)
	}

	require.NoError(t, uc.UpdateUserAccess(room.ID, 5, 1, "admin", 3))
}

func TestRoomUseCase_UpdateUserAccess_UserNotInRoom(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	roomRepo := newTestRoomRepo(t, db)
	userRepo := newTestUserRepo(t, db)
	sessionRepo := newTestSessionRepo(t, db)
	uc := NewRoomUseCase(roomRepo, userRepo, sessionRepo)

	room, _ := uc.Create(1, "R", "", "", 50, "")
	err := uc.UpdateUserAccess(room.ID, 999, 1, "admin", 2)
	assert.Error(t, err, "updating access for user not in room must fail")
}

func TestRoomUseCase_UpdateSettings_RejectsBadBounds(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	roomRepo := newTestRoomRepo(t, db)
	userRepo := newTestUserRepo(t, db)
	sessionRepo := newTestSessionRepo(t, db)
	uc := NewRoomUseCase(roomRepo, userRepo, sessionRepo)

	room, _ := uc.Create(1, "R", "", "", 50, "")

	bad := []struct {
		name string
		s    entity.RoomSettings
	}{
		{"max_users=0", entity.RoomSettings{MaxUsers: 0, SessionAutoEndMinutes: 60}},
		{"max_users>1000", entity.RoomSettings{MaxUsers: 5000, SessionAutoEndMinutes: 60}},
		{"session<5", entity.RoomSettings{MaxUsers: 50, SessionAutoEndMinutes: 1}},
		{"session>1440", entity.RoomSettings{MaxUsers: 50, SessionAutoEndMinutes: 9999}},
	}
	for _, tc := range bad {
		t.Run(tc.name, func(t *testing.T) {
			err := uc.UpdateSettings(room.ID, 1, "admin", &tc.s)
			assert.Error(t, err)
		})
	}

	good := &entity.RoomSettings{MaxUsers: 100, SessionAutoEndMinutes: 60}
	require.NoError(t, uc.UpdateSettings(room.ID, 1, "admin", good))
}

func TestRoomUseCase_RegenerateCode_IsRandom(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	roomRepo := newTestRoomRepo(t, db)
	userRepo := newTestUserRepo(t, db)
	sessionRepo := newTestSessionRepo(t, db)
	uc := NewRoomUseCase(roomRepo, userRepo, sessionRepo)

	room, _ := uc.Create(1, "R", "", "", 50, "")

	codes := make(map[string]bool)
	for i := 0; i < 5; i++ {
		code, err := uc.RegenerateCode(room.ID, 1, "admin")
		require.NoError(t, err)
		assert.NotContains(t, code, "-", "code must not embed roomID-timestamp format")
		assert.Len(t, code, 16, "code must be 16 hex chars (8 random bytes)")
		assert.False(t, codes[code], "codes must be unique across regenerations")
		codes[code] = true
	}
}

func TestRoomUseCase_Update_NonOwnerNonAdminForbidden(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	roomRepo := newTestRoomRepo(t, db)
	userRepo := newTestUserRepo(t, db)
	sessionRepo := newTestSessionRepo(t, db)
	uc := NewRoomUseCase(roomRepo, userRepo, sessionRepo)

	room, _ := uc.Create(1, "R", "", "", 50, "")
	_, err := uc.Update(room.ID, 99, "teacher", RoomUpdate{Name: strPtr("hijack")})
	assert.Error(t, err, "non-owner teacher must not update room")
}

func TestRoomUseCase_Delete_NonOwnerNonAdminForbidden(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	roomRepo := newTestRoomRepo(t, db)
	userRepo := newTestUserRepo(t, db)
	sessionRepo := newTestSessionRepo(t, db)
	uc := NewRoomUseCase(roomRepo, userRepo, sessionRepo)

	room, _ := uc.Create(1, "R", "", "", 50, "")
	err := uc.Delete(room.ID, 99, "teacher")
	assert.Error(t, err)
}
