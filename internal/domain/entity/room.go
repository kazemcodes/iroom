package entity

import "time"

// Room represents a persistent virtual classroom.
// Replaces the old Class entity. Rooms are owned by admins and contain
// assigned users and live sessions.
//
// Business rules:
//   - Only admins can create/modify rooms
//   - guest_login_enabled controls whether guests can join via link
//   - slug is URL-friendly identifier for /room/:slug links
type Room struct {
	ID                int64     `json:"id" db:"id"`
	OwnerID           int64     `json:"owner_id" db:"owner_id"`
	Name              string    `json:"name" db:"name"`
	Description       string    `json:"description" db:"description"`
	Color             string    `json:"color" db:"color"`
	Slug              string    `json:"slug" db:"slug"`
	GuestLoginEnabled bool      `json:"guest_login_enabled" db:"guest_login_enabled"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// RoomUser maps users to rooms with roles.
type RoomUser struct {
	RoomID int64  `json:"room_id" db:"room_id"`
	UserID int64  `json:"user_id" db:"user_id"`
	Role   string `json:"role" db:"role"` // "teacher" or "student"
}
