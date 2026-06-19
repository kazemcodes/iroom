package repository

import (
	"database/sql"
	"testing"

	"github.com/iroom/iroom/internal/domain/entity"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS room_users (
		room_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		role TEXT DEFAULT 'student',
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
	`
	_, err = db.Exec(schema)
	require.NoError(t, err)
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)
	return db
}

func TestRoomRepo_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRoomRepo(db)

	room := &entity.Room{
		OwnerID:           1,
		Name:              "Test Room",
		Description:       "Desc",
		Color:             "#FF0000",
		Slug:              "test-room",
		GuestLoginEnabled: true,
	}

	require.NoError(t, repo.Create(room))
	assert.NotZero(t, room.ID)
}

func TestRoomRepo_GetByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRoomRepo(db)

	room := &entity.Room{
		OwnerID:     1,
		Name:        "Fetch Room",
		Slug:        "fetch-room",
	}
	require.NoError(t, repo.Create(room))

	fetched, err := repo.GetByID(room.ID)
	require.NoError(t, err)
	assert.Equal(t, "Fetch Room", fetched.Name)
	assert.Equal(t, "fetch-room", fetched.Slug)
	assert.False(t, fetched.CreatedAt.IsZero())
}

func TestRoomRepo_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRoomRepo(db)

	_, err := repo.GetByID(9999)
	assert.Error(t, err)
}

func TestRoomRepo_GetBySlug(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRoomRepo(db)

	room := &entity.Room{
		OwnerID: 1,
		Name:    "Slug Room",
		Slug:    "my-slug",
	}
	require.NoError(t, repo.Create(room))

	fetched, err := repo.GetBySlug("my-slug")
	require.NoError(t, err)
	assert.Equal(t, room.ID, fetched.ID)
}

func TestRoomRepo_GetBySlug_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRoomRepo(db)

	_, err := repo.GetBySlug("nonexistent")
	assert.Error(t, err)
}

func TestRoomRepo_ListAll(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRoomRepo(db)

	for i := 0; i < 5; i++ {
		require.NoError(t, repo.Create(&entity.Room{
			OwnerID: 1,
			Name:    "Room",
			Slug:    "room-" + string(rune('a'+i)),
		}))
	}

	rooms, total, err := repo.ListAll(1, 3, "")
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, rooms, 3)
}

func TestRoomRepo_ListAll_Search(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRoomRepo(db)

	require.NoError(t, repo.Create(&entity.Room{OwnerID: 1, Name: "Math Class", Slug: "math"}))
	require.NoError(t, repo.Create(&entity.Room{OwnerID: 1, Name: "Science Lab", Slug: "science"}))

	rooms, total, err := repo.ListAll(1, 10, "Math")
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, "Math Class", rooms[0].Name)
}

func TestRoomRepo_ListAll_SearchDescription(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRoomRepo(db)

	require.NoError(t, repo.Create(&entity.Room{OwnerID: 1, Name: "Room A", Description: "Advanced algebra", Slug: "room-a"}))
	require.NoError(t, repo.Create(&entity.Room{OwnerID: 1, Name: "Room B", Description: "Basic physics", Slug: "room-b"}))

	rooms, total, err := repo.ListAll(1, 10, "algebra")
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, "Room A", rooms[0].Name)
}

func TestRoomRepo_ListAll_Pagination(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRoomRepo(db)

	for i := 0; i < 10; i++ {
		require.NoError(t, repo.Create(&entity.Room{
			OwnerID: 1,
			Name:    "Room",
			Slug:    "room-pag-" + string(rune('a'+i)),
		}))
	}

	rooms, total, err := repo.ListAll(2, 3, "")
	require.NoError(t, err)
	assert.Equal(t, int64(10), total)
	assert.Len(t, rooms, 3)
}

func TestRoomRepo_Update(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRoomRepo(db)

	room := &entity.Room{OwnerID: 1, Name: "Original", Slug: "original", Color: "#000"}
	require.NoError(t, repo.Create(room))

	room.Name = "Updated"
	room.Description = "New"
	room.Color = "#FFF"
	room.GuestLoginEnabled = false

	require.NoError(t, repo.Update(room))

	fetched, _ := repo.GetByID(room.ID)
	assert.Equal(t, "Updated", fetched.Name)
	assert.Equal(t, "New", fetched.Description)
	assert.False(t, fetched.GuestLoginEnabled)
}

func TestRoomRepo_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRoomRepo(db)

	room := &entity.Room{OwnerID: 1, Name: "Delete Me", Slug: "delete-me"}
	require.NoError(t, repo.Create(room))

	require.NoError(t, repo.Delete(room.ID))

	_, err := repo.GetByID(room.ID)
	assert.Error(t, err)
}

func TestRoomRepo_Count(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRoomRepo(db)

	count, _ := repo.Count()
	assert.Equal(t, int64(0), count)

	_ = repo.Create(&entity.Room{OwnerID: 1, Name: "A", Slug: "a"})
	_ = repo.Create(&entity.Room{OwnerID: 1, Name: "B", Slug: "b"})

	count, _ = repo.Count()
	assert.Equal(t, int64(2), count)
}

func TestRoomRepo_AddUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRoomRepo(db)

	room := &entity.Room{OwnerID: 1, Name: "Room", Slug: "room-ru"}
	require.NoError(t, repo.Create(room))

	err := repo.AddUser(room.ID, 2, "teacher")
	require.NoError(t, err)

	assert.True(t, repo.IsUserInRoom(room.ID, 2))
}

func TestRoomRepo_AddUser_ReplaceRole(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRoomRepo(db)

	room := &entity.Room{OwnerID: 1, Name: "Room", Slug: "room-ru2"}
	require.NoError(t, repo.Create(room))

	_ = repo.AddUser(room.ID, 3, "student")
	_ = repo.AddUser(room.ID, 3, "teacher")

	count, _ := repo.GetUserCount(room.ID)
	assert.Equal(t, 1, count, "should not create duplicate user entries")
}

func TestRoomRepo_RemoveUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRoomRepo(db)

	room := &entity.Room{OwnerID: 1, Name: "Room", Slug: "room-ru3"}
	require.NoError(t, repo.Create(room))

	_ = repo.AddUser(room.ID, 4, "student")
	require.NoError(t, repo.RemoveUser(room.ID, 4))

	assert.False(t, repo.IsUserInRoom(room.ID, 4))
}

func TestRoomRepo_GetUsers(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	roomRepo := NewRoomRepo(db)
	userRepo := NewUserRepo(db)

	room := &entity.Room{OwnerID: 1, Name: "Room", Slug: "room-gu"}
	require.NoError(t, roomRepo.Create(room))

	u1 := &entity.User{Email: "a@test.com", DisplayName: "A", Role: "teacher"}
	u2 := &entity.User{Email: "b@test.com", DisplayName: "B", Role: "student"}
	require.NoError(t, userRepo.Create(u1))
	require.NoError(t, userRepo.Create(u2))

	_ = roomRepo.AddUser(room.ID, u1.ID, "teacher")
	_ = roomRepo.AddUser(room.ID, u2.ID, "student")

	users, err := roomRepo.GetUsers(room.ID)
	require.NoError(t, err)
	assert.Len(t, users, 2)
}

func TestRoomRepo_IsUserInRoom(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRoomRepo(db)

	room := &entity.Room{OwnerID: 1, Name: "Room", Slug: "room-check"}
	require.NoError(t, repo.Create(room))

	assert.False(t, repo.IsUserInRoom(room.ID, 99))
}

func TestRoomRepo_GetUserCount(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRoomRepo(db)

	room := &entity.Room{OwnerID: 1, Name: "Room", Slug: "room-uc"}
	require.NoError(t, repo.Create(room))

	count, _ := repo.GetUserCount(room.ID)
	assert.Equal(t, 0, count)

	_ = repo.AddUser(room.ID, 10, "student")
	_ = repo.AddUser(room.ID, 11, "student")

	count, _ = repo.GetUserCount(room.ID)
	assert.Equal(t, 2, count)
}

func TestRoomRepo_GetSettings_Default(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRoomRepo(db)

	room := &entity.Room{OwnerID: 1, Name: "Room", Slug: "room-s"}
	require.NoError(t, repo.Create(room))

	settings, err := repo.GetSettings(room.ID)
	require.NoError(t, err)
	assert.Equal(t, 50, settings.MaxUsers)
	assert.True(t, settings.RecordingEnabled)
	assert.Equal(t, 120, settings.SessionAutoEndMinutes)
}

func TestRoomRepo_UpdateSettings(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRoomRepo(db)

	room := &entity.Room{OwnerID: 1, Name: "Room", Slug: "room-us"}
	require.NoError(t, repo.Create(room))

	s := &entity.RoomSettings{
		RoomID:                    room.ID,
		MaxUsers:                  100,
		RecordingEnabled:          false,
		AllowStudentVideo:         true,
		AllowStudentAudio:         false,
		AllowStudentScreenShare:   true,
		AllowStudentWhiteboard:    true,
		AllowStudentChat:          false,
		SessionAutoEndMinutes:     45,
		WaitingRoomEnabled:        true,
	}
	require.NoError(t, repo.UpdateSettings(s))

	fetched, err := repo.GetSettings(room.ID)
	require.NoError(t, err)
	assert.Equal(t, 100, fetched.MaxUsers)
	assert.False(t, fetched.RecordingEnabled)
	assert.True(t, fetched.AllowStudentVideo)
	assert.False(t, fetched.AllowStudentAudio)
	assert.True(t, fetched.AllowStudentScreenShare)
	assert.True(t, fetched.AllowStudentWhiteboard)
	assert.False(t, fetched.AllowStudentChat)
	assert.Equal(t, 45, fetched.SessionAutoEndMinutes)
	assert.True(t, fetched.WaitingRoomEnabled)
}

func TestRoomRepo_ListByUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRoomRepo(db)

	r1 := &entity.Room{OwnerID: 1, Name: "Owned", Slug: "owned"}
	r2 := &entity.Room{OwnerID: 2, Name: "Assigned", Slug: "assigned"}
	r3 := &entity.Room{OwnerID: 3, Name: "Other", Slug: "other"}
	_ = repo.Create(r1)
	_ = repo.Create(r2)
	_ = repo.Create(r3)

	_ = repo.AddUser(r2.ID, 1, "student")

	rooms, err := repo.ListByUser(1)
	require.NoError(t, err)
	assert.Len(t, rooms, 2, "user 1 should see owned + assigned rooms")
}

func TestRoomRepo_Create_AllFieldsPreserved(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRoomRepo(db)

	room := &entity.Room{
		OwnerID:           42,
		Name:              "Full Room",
		Description:       "A detailed description",
		Color:             "#FF5733",
		Slug:              "full-room",
		GuestLoginEnabled: true,
	}
	require.NoError(t, repo.Create(room))
	require.NotZero(t, room.ID)

	fetched, err := repo.GetByID(room.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(42), fetched.OwnerID, "owner_id must be preserved")
	assert.Equal(t, "Full Room", fetched.Name)
	assert.Equal(t, "A detailed description", fetched.Description)
	assert.Equal(t, "#FF5733", fetched.Color)
	assert.Equal(t, "full-room", fetched.Slug)
	assert.True(t, fetched.GuestLoginEnabled)
	assert.False(t, fetched.CreatedAt.IsZero())
	assert.False(t, fetched.UpdatedAt.IsZero())
}

func TestRoomRepo_SlugUniqueness(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRoomRepo(db)

	require.NoError(t, repo.Create(&entity.Room{OwnerID: 1, Name: "A", Slug: "same-slug"}))
	err := repo.Create(&entity.Room{OwnerID: 1, Name: "B", Slug: "same-slug"})
	assert.Error(t, err, "duplicate slug should fail")
}

func TestRoomRepo_GuestLoginEnabled_Roundtrip(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRoomRepo(db)

	room := &entity.Room{OwnerID: 1, Name: "R", Slug: "gl", GuestLoginEnabled: true}
	require.NoError(t, repo.Create(room))

	fetched, _ := repo.GetByID(room.ID)
	assert.True(t, fetched.GuestLoginEnabled)

	fetched.GuestLoginEnabled = false
	require.NoError(t, repo.Update(fetched))

	fetched2, _ := repo.GetByID(room.ID)
	assert.False(t, fetched2.GuestLoginEnabled, "guest_login_enabled=false must survive update")
}

func TestRoomRepo_Update_PreservesOwnerID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRoomRepo(db)

	room := &entity.Room{OwnerID: 99, Name: "Original", Slug: "preserve-owner"}
	require.NoError(t, repo.Create(room))

	fetched, _ := repo.GetByID(room.ID)
	fetched.Name = "Changed"
	require.NoError(t, repo.Update(fetched))

	fetched2, _ := repo.GetByID(room.ID)
	assert.Equal(t, int64(99), fetched2.OwnerID, "update must not change owner_id")
	assert.Equal(t, "Changed", fetched2.Name)
}

func TestRoomRepo_Delete_CascadesRoomUsers(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRoomRepo(db)

	room := &entity.Room{OwnerID: 1, Name: "Cascade", Slug: "cascade"}
	require.NoError(t, repo.Create(room))

	_ = repo.AddUser(room.ID, 10, "student")
	_ = repo.AddUser(room.ID, 11, "teacher")
	assert.Equal(t, 2, mustCount(t, db, "SELECT COUNT(*) FROM room_users WHERE room_id = ?", room.ID))

	require.NoError(t, repo.Delete(room.ID))

	assert.Equal(t, 0, mustCount(t, db, "SELECT COUNT(*) FROM room_users WHERE room_id = ?", room.ID),
		"room_users should be cascade-deleted when room is deleted")
}

func TestRoomRepo_Delete_CascadesRoomSettings(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRoomRepo(db)

	room := &entity.Room{OwnerID: 1, Name: "CascadeSettings", Slug: "cascade-settings"}
	require.NoError(t, repo.Create(room))

	require.NoError(t, repo.UpdateSettings(&entity.RoomSettings{RoomID: room.ID, MaxUsers: 99}))
	assert.Equal(t, 1, mustCount(t, db, "SELECT COUNT(*) FROM room_settings WHERE room_id = ?", room.ID))

	require.NoError(t, repo.Delete(room.ID))

	assert.Equal(t, 0, mustCount(t, db, "SELECT COUNT(*) FROM room_settings WHERE room_id = ?", room.ID),
		"room_settings should be cascade-deleted when room is deleted")
}

func mustCount(t *testing.T, db *sql.DB, query string, args ...interface{}) int {
	t.Helper()
	var count int
	err := db.QueryRow(query, args...).Scan(&count)
	require.NoError(t, err)
	return count
}

func TestRoomRepo_GetUsers_ReturnsCorrectFields(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	roomRepo := NewRoomRepo(db)
	userRepo := NewUserRepo(db)

	room := &entity.Room{OwnerID: 1, Name: "Room", Slug: "room-fields"}
	require.NoError(t, roomRepo.Create(room))

	u1 := &entity.User{Email: "teacher@test.com", DisplayName: "Teacher Ali", Role: "teacher", IsActive: true}
	u2 := &entity.User{Email: "student@test.com", DisplayName: "Student Sara", Role: "student", IsActive: true}
	require.NoError(t, userRepo.Create(u1))
	require.NoError(t, userRepo.Create(u2))

	_ = roomRepo.AddUser(room.ID, u1.ID, "teacher")
	_ = roomRepo.AddUser(room.ID, u2.ID, "student")

	users, err := roomRepo.GetUsers(room.ID)
	require.NoError(t, err)
	require.Len(t, users, 2)

	userMap := make(map[string]entity.User)
	for _, u := range users {
		userMap[u.Email] = u
	}

	teacher := userMap["teacher@test.com"]
	assert.Equal(t, "Teacher Ali", teacher.DisplayName, "display_name must be returned")
	assert.Equal(t, "teacher", string(teacher.Role), "role must be returned")
	assert.True(t, teacher.IsActive)

	student := userMap["student@test.com"]
	assert.Equal(t, "Student Sara", student.DisplayName)
	assert.Equal(t, "student", string(student.Role))
}

func TestRoomRepo_SettingsAllDefaults(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRoomRepo(db)

	room := &entity.Room{OwnerID: 1, Name: "Defaults", Slug: "defaults"}
	require.NoError(t, repo.Create(room))

	s, err := repo.GetSettings(room.ID)
	require.NoError(t, err)

	assert.Equal(t, 50, s.MaxUsers)
	assert.True(t, s.RecordingEnabled)
	assert.False(t, s.AllowStudentVideo, "video should be off by default")
	assert.True(t, s.AllowStudentAudio, "audio should be on by default")
	assert.False(t, s.AllowStudentScreenShare, "screen share should be off by default")
	assert.False(t, s.AllowStudentWhiteboard, "whiteboard should be off by default")
	assert.True(t, s.AllowStudentChat, "chat should be on by default")
	assert.Equal(t, 120, s.SessionAutoEndMinutes)
	assert.False(t, s.WaitingRoomEnabled, "waiting room should be off by default")
}

func TestRoomRepo_SettingsRoundtrip_AllFields(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRoomRepo(db)

	room := &entity.Room{OwnerID: 1, Name: "FullSettings", Slug: "full-settings"}
	require.NoError(t, repo.Create(room))

	s := &entity.RoomSettings{
		RoomID:                  room.ID,
		MaxUsers:                200,
		RecordingEnabled:        false,
		AllowStudentVideo:       true,
		AllowStudentAudio:       false,
		AllowStudentScreenShare: true,
		AllowStudentWhiteboard:  true,
		AllowStudentChat:        false,
		SessionAutoEndMinutes:   30,
		WaitingRoomEnabled:      true,
	}
	require.NoError(t, repo.UpdateSettings(s))

	fetched, err := repo.GetSettings(room.ID)
	require.NoError(t, err)

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

func TestRoomRepo_AddUser_TeacherRolePreserved(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	roomRepo := NewRoomRepo(db)
	userRepo := NewUserRepo(db)

	room := &entity.Room{OwnerID: 1, Name: "R", Slug: "role-preserve"}
	require.NoError(t, roomRepo.Create(room))

	u := &entity.User{Email: "t@test.com", DisplayName: "T", Role: "teacher"}
	require.NoError(t, userRepo.Create(u))

	require.NoError(t, roomRepo.AddUser(room.ID, u.ID, "teacher"))

	users, err := roomRepo.GetUsers(room.ID)
	require.NoError(t, err)
	require.Len(t, users, 1)
	assert.Equal(t, "teacher", string(users[0].Role), "teacher role must be preserved through AddUser → GetUsers")
}

func TestRoomRepo_AddUser_StudentRolePreserved(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	roomRepo := NewRoomRepo(db)
	userRepo := NewUserRepo(db)

	room := &entity.Room{OwnerID: 1, Name: "R", Slug: "student-role"}
	require.NoError(t, roomRepo.Create(room))

	u := &entity.User{Email: "s@test.com", DisplayName: "S", Role: "student"}
	require.NoError(t, userRepo.Create(u))

	require.NoError(t, roomRepo.AddUser(room.ID, u.ID, "student"))

	users, err := roomRepo.GetUsers(room.ID)
	require.NoError(t, err)
	require.Len(t, users, 1)
	assert.Equal(t, "student", string(users[0].Role))
}

func TestRoomRepo_ListAll_Empty(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRoomRepo(db)

	rooms, total, err := repo.ListAll(1, 10, "")
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, rooms)
}

func TestRoomRepo_ListAll_SearchNoMatch(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRoomRepo(db)

	require.NoError(t, repo.Create(&entity.Room{OwnerID: 1, Name: "Physics", Slug: "physics"}))

	rooms, total, err := repo.ListAll(1, 10, "xyz")
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, rooms)
}

func TestRoomRepo_GetBySlug_AllFieldsPreserved(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewRoomRepo(db)

	room := &entity.Room{
		OwnerID:           7,
		Name:              "Slug Room",
		Description:       "desc",
		Color:             "#ABCDEF",
		Slug:              "slug-room",
		GuestLoginEnabled: false,
	}
	require.NoError(t, repo.Create(room))

	fetched, err := repo.GetBySlug("slug-room")
	require.NoError(t, err)
	assert.Equal(t, int64(7), fetched.OwnerID)
	assert.Equal(t, "Slug Room", fetched.Name)
	assert.Equal(t, "desc", fetched.Description)
	assert.Equal(t, "#ABCDEF", fetched.Color)
	assert.False(t, fetched.GuestLoginEnabled)
}
