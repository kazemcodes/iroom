package repository

import (
	"database/sql"

	"github.com/iroom/iroom/internal/domain/entity"
)

type RoomRepo struct {
	db *sql.DB
}

func NewRoomRepo(db *sql.DB) *RoomRepo {
	return &RoomRepo{db: db}
}

func (r *RoomRepo) Create(room *entity.Room) error {
	result, err := r.db.Exec(
		`INSERT INTO rooms (owner_id, name, description, color, slug, guest_login_enabled) VALUES (?, ?, ?, ?, ?, ?)`,
		room.OwnerID, room.Name, room.Description, room.Color, room.Slug, room.GuestLoginEnabled,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	room.ID = id
	return nil
}

func (r *RoomRepo) GetByID(id int64) (*entity.Room, error) {
	room := &entity.Room{}
	err := r.db.QueryRow(
		`SELECT id, owner_id, name, description, color, slug, guest_login_enabled, created_at, updated_at FROM rooms WHERE id = ?`, id,
	).Scan(&room.ID, &room.OwnerID, &room.Name, &room.Description, &room.Color, &room.Slug, &room.GuestLoginEnabled, &room.CreatedAt, &room.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return room, nil
}

func (r *RoomRepo) GetBySlug(slug string) (*entity.Room, error) {
	room := &entity.Room{}
	err := r.db.QueryRow(
		`SELECT id, owner_id, name, description, color, slug, guest_login_enabled, created_at, updated_at FROM rooms WHERE slug = ?`, slug,
	).Scan(&room.ID, &room.OwnerID, &room.Name, &room.Description, &room.Color, &room.Slug, &room.GuestLoginEnabled, &room.CreatedAt, &room.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return room, nil
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
	query := `SELECT id, owner_id, name, description, color, slug, guest_login_enabled, created_at, updated_at FROM rooms WHERE 1=1`
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
		if err := rows.Scan(&room.ID, &room.OwnerID, &room.Name, &room.Description, &room.Color, &room.Slug, &room.GuestLoginEnabled, &room.CreatedAt, &room.UpdatedAt); err != nil {
			return nil, 0, err
		}
		rooms = append(rooms, room)
	}
	return rooms, total, nil
}

func (r *RoomRepo) ListByUser(userID int64) ([]entity.Room, error) {
	rows, err := r.db.Query(
		`SELECT id, owner_id, name, description, color, slug, guest_login_enabled, created_at, updated_at
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
		if err := rows.Scan(&room.ID, &room.OwnerID, &room.Name, &room.Description, &room.Color, &room.Slug, &room.GuestLoginEnabled, &room.CreatedAt, &room.UpdatedAt); err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}
	return rooms, nil
}

func (r *RoomRepo) Update(room *entity.Room) error {
	_, err := r.db.Exec(
		`UPDATE rooms SET name = ?, description = ?, color = ?, guest_login_enabled = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		room.Name, room.Description, room.Color, room.GuestLoginEnabled, room.ID,
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

// Room users

func (r *RoomRepo) AddUser(roomID, userID int64, role string) error {
	_, err := r.db.Exec(
		`INSERT OR REPLACE INTO room_users (room_id, user_id, role) VALUES (?, ?, ?)`,
		roomID, userID, role,
	)
	return err
}

func (r *RoomRepo) RemoveUser(roomID, userID int64) error {
	_, err := r.db.Exec(`DELETE FROM room_users WHERE room_id = ? AND user_id = ?`, roomID, userID)
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
