package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/iroom/iroom/internal/domain/entity"
)

type RoomRepo struct {
	db *sql.DB
}

func NewRoomRepo(db *sql.DB) *RoomRepo {
	return &RoomRepo{db: db}
}

func (r *RoomRepo) Create(room *entity.Room) error {
	// Loop on UNIQUE-constraint failure so concurrent creates of the same name
	// don't both observe count=0 then collide on insert.
	base := room.Slug
	for i := 0; i < 100; i++ {
		candidate := base
		if i > 0 {
			candidate = fmt.Sprintf("%s-%d", base, i)
		}
		result, err := r.db.Exec(
			`INSERT INTO rooms (owner_id, name, description, color, slug, guest_login_enabled, max_users, invite_code, is_archived)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			room.OwnerID, room.Name, room.Description, room.Color, candidate,
			room.GuestLoginEnabled, room.MaxUsers, room.InviteCode, room.IsArchived,
		)
		if err != nil {
			// UNIQUE constraint failure on slug → try next suffix
			if isUniqueConstraint(err) {
				continue
			}
			return err
		}
		id, err := result.LastInsertId()
		if err != nil {
			return err
		}
		room.ID = id
		room.Slug = candidate
		return nil
	}
	// give up — use timestamp as last-resort suffix
	candidate := fmt.Sprintf("%s-%d", base, time.Now().UnixMilli())
	result, err := r.db.Exec(
		`INSERT INTO rooms (owner_id, name, description, color, slug, guest_login_enabled, max_users, invite_code, is_archived)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		room.OwnerID, room.Name, room.Description, room.Color, candidate,
		room.GuestLoginEnabled, room.MaxUsers, room.InviteCode, room.IsArchived,
	)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	room.ID = id
	room.Slug = candidate
	return nil
}

func isUniqueConstraint(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	// sqlite3 driver returns: "UNIQUE constraint failed: rooms.slug"
	return strings.Contains(msg, "UNIQUE constraint failed") || strings.Contains(msg, "constraint failed: rooms.slug")
}

func (r *RoomRepo) GetByID(id int64) (*entity.Room, error) {
	room := &entity.Room{}
	err := r.db.QueryRow(
		`SELECT id, owner_id, name, description, color, slug, guest_login_enabled, max_users, invite_code, is_archived, created_at, updated_at
		 FROM rooms WHERE id = ?`, id,
	).Scan(&room.ID, &room.OwnerID, &room.Name, &room.Description, &room.Color,
		&room.Slug, &room.GuestLoginEnabled, &room.MaxUsers, &room.InviteCode,
		&room.IsArchived, &room.CreatedAt, &room.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return room, nil
}

func (r *RoomRepo) GetBySlug(slug string) (*entity.Room, error) {
	room := &entity.Room{}
	err := r.db.QueryRow(
		`SELECT id, owner_id, name, description, color, slug, guest_login_enabled, max_users, invite_code, is_archived, created_at, updated_at
		 FROM rooms WHERE slug = ?`, slug,
	).Scan(&room.ID, &room.OwnerID, &room.Name, &room.Description, &room.Color,
		&room.Slug, &room.GuestLoginEnabled, &room.MaxUsers, &room.InviteCode,
		&room.IsArchived, &room.CreatedAt, &room.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return room, nil
}

func (r *RoomRepo) GetByInviteCode(code string) (*entity.Room, error) {
	room := &entity.Room{}
	err := r.db.QueryRow(
		`SELECT id, owner_id, name, description, color, slug, guest_login_enabled, max_users, invite_code, is_archived, created_at, updated_at
		 FROM rooms WHERE invite_code = ?`, code,
	).Scan(&room.ID, &room.OwnerID, &room.Name, &room.Description, &room.Color,
		&room.Slug, &room.GuestLoginEnabled, &room.MaxUsers, &room.InviteCode,
		&room.IsArchived, &room.CreatedAt, &room.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return room, nil
}

func (r *RoomRepo) UpdateInviteCode(roomID int64, code string) error {
	_, err := r.db.Exec(
		`UPDATE rooms SET invite_code = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		code, roomID,
	)
	return err
}

func (r *RoomRepo) ListAll(page, perPage int, search string) ([]entity.Room, int64, error) {
	var total int64
	countQuery := `SELECT COUNT(*) FROM rooms WHERE 1=1`
	args := []interface{}{}
	if search != "" {
		countQuery += ` AND (name LIKE ? OR description LIKE ?)`
		s := "%" + search + "%"
		args = append(args, s, s)
	}
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	query := `SELECT id, owner_id, name, description, color, slug, guest_login_enabled, max_users, invite_code, is_archived, created_at, updated_at FROM rooms WHERE 1=1`
	if search != "" {
		query += ` AND (name LIKE ? OR description LIKE ?)`
	}
	query += ` ORDER BY id DESC LIMIT ? OFFSET ?`
	args = append(args, perPage, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var rooms []entity.Room
	for rows.Next() {
		var room entity.Room
		if err := rows.Scan(&room.ID, &room.OwnerID, &room.Name, &room.Description,
			&room.Color, &room.Slug, &room.GuestLoginEnabled, &room.MaxUsers,
			&room.InviteCode, &room.IsArchived, &room.CreatedAt, &room.UpdatedAt); err != nil {
			return nil, 0, err
		}
		rooms = append(rooms, room)
	}
	return rooms, total, nil
}

func (r *RoomRepo) ListByUser(userID int64) ([]entity.Room, error) {
	rows, err := r.db.Query(
		`SELECT id, owner_id, name, description, color, slug, guest_login_enabled, max_users, invite_code, is_archived, created_at, updated_at
		 FROM rooms WHERE owner_id = ? OR id IN (SELECT room_id FROM room_users WHERE user_id = ?)
		 ORDER BY id DESC`, userID, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []entity.Room
	for rows.Next() {
		var room entity.Room
		if err := rows.Scan(&room.ID, &room.OwnerID, &room.Name, &room.Description,
			&room.Color, &room.Slug, &room.GuestLoginEnabled, &room.MaxUsers,
			&room.InviteCode, &room.IsArchived, &room.CreatedAt, &room.UpdatedAt); err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}
	return rooms, nil
}

func (r *RoomRepo) Update(room *entity.Room) error {
	_, err := r.db.Exec(
		`UPDATE rooms SET name = ?, description = ?, color = ?, slug = ?, guest_login_enabled = ?, max_users = ?, invite_code = ?, is_archived = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		room.Name, room.Description, room.Color, room.Slug, room.GuestLoginEnabled,
		room.MaxUsers, room.InviteCode, room.IsArchived, room.ID,
	)
	return err
}

func (r *RoomRepo) Delete(id int64) error {
	_, err := r.db.Exec(`DELETE FROM rooms WHERE id = ?`, id)
	return err
}

func (r *RoomRepo) Count() (int64, error) {
	var count int64
	err := r.db.QueryRow(`SELECT COUNT(*) FROM rooms`).Scan(&count)
	return count, err
}

func (r *RoomRepo) AddUser(roomID, userID int64, role string, access int) error {
	_, err := r.db.Exec(
		`INSERT OR REPLACE INTO room_users (room_id, user_id, role, access) VALUES (?, ?, ?, ?)`,
		roomID, userID, role, access,
	)
	return err
}

func (r *RoomRepo) RemoveUser(roomID, userID int64) error {
	_, err := r.db.Exec(`DELETE FROM room_users WHERE room_id = ? AND user_id = ?`, roomID, userID)
	return err
}

func (r *RoomRepo) UpdateUserAccess(roomID, userID int64, access int) error {
	_, err := r.db.Exec(
		`UPDATE room_users SET access = ? WHERE room_id = ? AND user_id = ?`,
		access, roomID, userID,
	)
	return err
}

func (r *RoomRepo) GetUsers(roomID int64) ([]entity.User, error) {
	rows, err := r.db.Query(
		`SELECT u.id, u.email, u.display_name, u.role, u.is_active, u.created_at, u.updated_at
		 FROM users u JOIN room_users ru ON u.id = ru.user_id WHERE ru.room_id = ?`, roomID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []entity.User
	for rows.Next() {
		var u entity.User
		if err := rows.Scan(&u.ID, &u.Email, &u.DisplayName, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *RoomRepo) IsUserInRoom(roomID, userID int64) bool {
	var count int
	r.db.QueryRow(`SELECT COUNT(*) FROM room_users WHERE room_id = ? AND user_id = ?`, roomID, userID).Scan(&count)
	return count > 0
}

func (r *RoomRepo) GetUserCount(roomID int64) (int, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM room_users WHERE room_id = ?`, roomID).Scan(&count)
	return count, err
}

func (r *RoomRepo) GetSettings(roomID int64) (*entity.RoomSettings, error) {
	s := &entity.RoomSettings{}
	err := r.db.QueryRow(
		`SELECT room_id, max_users, recording_enabled, allow_student_video, allow_student_audio,
		 allow_student_screen_share, allow_student_whiteboard, allow_student_chat,
		 session_auto_end_minutes, waiting_room_enabled FROM room_settings WHERE room_id = ?`, roomID,
	).Scan(&s.RoomID, &s.MaxUsers, &s.RecordingEnabled, &s.AllowStudentVideo, &s.AllowStudentAudio,
		&s.AllowStudentScreenShare, &s.AllowStudentWhiteboard, &s.AllowStudentChat,
		&s.SessionAutoEndMinutes, &s.WaitingRoomEnabled)
	if err != nil {
		return &entity.RoomSettings{
			RoomID:                roomID,
			MaxUsers:              50,
			RecordingEnabled:      true,
			AllowStudentVideo:     false,
			AllowStudentAudio:     true,
			AllowStudentChat:      true,
			SessionAutoEndMinutes: 120,
		}, nil
	}
	return s, nil
}

func (r *RoomRepo) UpdateSettings(s *entity.RoomSettings) error {
	_, err := r.db.Exec(
		`INSERT OR REPLACE INTO room_settings (room_id, max_users, recording_enabled, allow_student_video, allow_student_audio,
		 allow_student_screen_share, allow_student_whiteboard, allow_student_chat, session_auto_end_minutes, waiting_room_enabled, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`,
		s.RoomID, s.MaxUsers, s.RecordingEnabled, s.AllowStudentVideo, s.AllowStudentAudio,
		s.AllowStudentScreenShare, s.AllowStudentWhiteboard, s.AllowStudentChat,
		s.SessionAutoEndMinutes, s.WaitingRoomEnabled,
	)
	return err
}
