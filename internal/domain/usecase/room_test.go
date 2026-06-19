package usecase

import (
	"testing"

	"github.com/iroom/iroom/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
			result := generateRoomSlug(tt.input)
			if tt.input == "" || tt.input == "@#$%^&*()" {
				assert.Regexp(t, `^room-\d+$`, result, "fallback slug should match pattern")
				return
			}
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateRoomSlugDoubleDashCollapse(t *testing.T) {
	slug := generateRoomSlug("a  b")
	assert.NotContains(t, slug, "--", "double dashes should be collapsed")
}

func TestRoomUseCase_Create_DefaultGuestLogin(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	roomRepo := newTestRoomRepo(t, db)
	userRepo := newTestUserRepo(t, db)
	sessionRepo := newTestSessionRepo(t, db)

	uc := NewRoomUseCase(roomRepo, userRepo, sessionRepo)

	room, err := uc.Create(1, "Test Room", "Description", "#FF0000")
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

	room, err := uc.Create(1, "", "Description", "#FF0000")
	require.NoError(t, err)
	assert.Regexp(t, `^room-\d+$`, room.Slug, "empty name gets fallback slug")
}

func TestRoomUseCase_GetByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	roomRepo := newTestRoomRepo(t, db)
	userRepo := newTestUserRepo(t, db)
	sessionRepo := newTestSessionRepo(t, db)

	uc := NewRoomUseCase(roomRepo, userRepo, sessionRepo)

	created, err := uc.Create(1, "My Room", "Desc", "#00FF00")
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

	created, err := uc.Create(1, "Slug Room", "Desc", "#0000FF")
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
		_, err := uc.Create(1, "Room "+string(rune('A'+i)), "Desc", "#FF0000")
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

	_, _ = uc.Create(1, "Math Class", "Algebra", "#FF0000")
	_, _ = uc.Create(1, "Science Lab", "Physics", "#00FF00")

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

	created, err := uc.Create(1, "Original", "Old Desc", "#FF0000")
	require.NoError(t, err)

	updated, err := uc.Update(created.ID, "Updated", "New Desc", "#00FF00", false)
	require.NoError(t, err)
	assert.Equal(t, "Updated", updated.Name)
	assert.Equal(t, "New Desc", updated.Description)
	assert.Equal(t, "#00FF00", updated.Color)
	assert.False(t, updated.GuestLoginEnabled)
}

func TestRoomUseCase_Update_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	roomRepo := newTestRoomRepo(t, db)
	userRepo := newTestUserRepo(t, db)
	sessionRepo := newTestSessionRepo(t, db)

	uc := NewRoomUseCase(roomRepo, userRepo, sessionRepo)

	_, err := uc.Update(9999, "X", "X", "X", true)
	assert.Error(t, err)
}

func TestRoomUseCase_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	roomRepo := newTestRoomRepo(t, db)
	userRepo := newTestUserRepo(t, db)
	sessionRepo := newTestSessionRepo(t, db)

	uc := NewRoomUseCase(roomRepo, userRepo, sessionRepo)

	created, err := uc.Create(1, "To Delete", "Desc", "#FF0000")
	require.NoError(t, err)

	err = uc.Delete(created.ID)
	require.NoError(t, err)

	_, err = uc.GetByID(created.ID)
	assert.Error(t, err)
}

func TestRoomUseCase_AddUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	roomRepo := newTestRoomRepo(t, db)
	userRepo := newTestUserRepo(t, db)
	sessionRepo := newTestSessionRepo(t, db)

	uc := NewRoomUseCase(roomRepo, userRepo, sessionRepo)

	room, _ := uc.Create(1, "Room", "Desc", "#FF0000")
	err := uc.AddUser(room.ID, 2, "student")
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

	room, _ := uc.Create(1, "Room", "Desc", "#FF0000")
	err := uc.AddUser(room.ID, 3, "")
	require.NoError(t, err)
}

func TestRoomUseCase_RemoveUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	roomRepo := newTestRoomRepo(t, db)
	userRepo := newTestUserRepo(t, db)
	sessionRepo := newTestSessionRepo(t, db)

	uc := NewRoomUseCase(roomRepo, userRepo, sessionRepo)

	room, _ := uc.Create(1, "Room", "Desc", "#FF0000")
	_ = uc.AddUser(room.ID, 4, "student")

	err := uc.RemoveUser(room.ID, 4)
	require.NoError(t, err)

	count, err := uc.GetUserCount(room.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestRoomUseCase_GetUsers_Empty(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	roomRepo := newTestRoomRepo(t, db)
	userRepo := newTestUserRepo(t, db)
	sessionRepo := newTestSessionRepo(t, db)

	uc := NewRoomUseCase(roomRepo, userRepo, sessionRepo)

	room, _ := uc.Create(1, "Room", "Desc", "#FF0000")

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

	room, _ := uc.Create(1, "Room", "Desc", "#FF0000")

	settings, err := uc.GetSettings(room.ID)
	require.NoError(t, err)
	assert.Equal(t, 50, settings.MaxUsers)
	assert.True(t, settings.RecordingEnabled)
	assert.True(t, settings.AllowStudentAudio)
	assert.True(t, settings.AllowStudentChat)
	assert.Equal(t, 120, settings.SessionAutoEndMinutes)

	settings.MaxUsers = 100
	settings.AllowStudentVideo = true
	err = uc.UpdateSettings(room.ID, settings)
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

	room, err := uc.Create(42, "Owner Room", "Desc", "#000")
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

	room, _ := uc.Create(5, "Original", "Original Desc", "#FF0000")
	assert.True(t, room.GuestLoginEnabled)

	// Update only name, leave other fields empty (usecase skips empty)
	_, err := uc.Update(room.ID, "New Name", "", "", true)
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

	room, _ := uc.Create(1, "R", "D", "#000")
	assert.True(t, room.GuestLoginEnabled)

	_, err := uc.Update(room.ID, "", "", "", false)
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

	room, _ := uc.Create(1, "R", "D", "#000")

	u := &entity.User{Email: "teacher@test.com", DisplayName: "T", Role: "teacher"}
	require.NoError(t, userRepo.Create(u))

	err := uc.AddUser(room.ID, u.ID, "teacher")
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

	room, _ := uc.Create(1, "R", "D", "#000")

	u := &entity.User{Email: "student@test.com", DisplayName: "S", Role: "student"}
	require.NoError(t, userRepo.Create(u))

	err := uc.AddUser(room.ID, u.ID, "student")
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

	room, _ := uc.Create(1, "R", "D", "#000")

	// Create a user in the user repo
	u := &entity.User{Email: "test@test.com", DisplayName: "Test User", Role: "teacher", IsActive: true}
	require.NoError(t, userRepo.Create(u))

	_ = uc.AddUser(room.ID, u.ID, "teacher")

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

	room, _ := uc.Create(1, "R", "D", "#000")

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
	require.NoError(t, uc.UpdateSettings(room.ID, s))

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

	created, _ := uc.Create(7, "Field Room", "A description", "#123456")

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

	created, _ := uc.Create(3, "Slug Full", "Desc", "#ABC")

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

	room, _ := uc.Create(1, "Delete Me", "Desc", "#000")
	require.NoError(t, uc.Delete(room.ID))

	_, err := uc.GetByID(room.ID)
	assert.Error(t, err)

	_, err = uc.GetBySlug(room.Slug)
	assert.Error(t, err)
}
