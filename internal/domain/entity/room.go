package entity

import "time"

// Room represents a persistent virtual classroom.
// Merges the old Class entity and Room entity into one concept.
// Rooms are owned by teachers/admins and contain assigned users and live sessions.
//
// Business rules:
//   - Only the room owner (teacher) or admin can modify/delete rooms
//   - guest_login_enabled controls whether guests can join via link
//   - slug is URL-friendly identifier for /room/:slug links
//   - invite_code allows students to self-enroll via POST /rooms/join/:code
//   - max_users limits concurrent participants (from old Class.max_students)
type Room struct {
	ID                int64     `json:"id" db:"id"`
	OwnerID           int64     `json:"owner_id" db:"owner_id"`
	Name              string    `json:"name" db:"name"`
	Description       string    `json:"description" db:"description"`
	Color             string    `json:"color" db:"color"`
	Slug              string    `json:"slug" db:"slug"`
	GuestLoginEnabled bool      `json:"guest_login_enabled" db:"guest_login_enabled"`
	MaxUsers          int       `json:"max_users" db:"max_users"`
	InviteCode        string    `json:"invite_code,omitempty" db:"invite_code"`
	IsArchived        bool      `json:"is_archived" db:"is_archived"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// RoomUser maps users to rooms with roles and access levels.
//
// Role: "teacher" or "student"
// Access levels:
//   1 = Regular student (default)
//   2 = Presenter (can share screen/whiteboard)
//   3 = Operator (full control within the room)
type RoomUser struct {
	RoomID int64  `json:"room_id" db:"room_id"`
	UserID int64  `json:"user_id" db:"user_id"`
	Role   string `json:"role" db:"role"`
	Access int    `json:"access" db:"access"`
}

// RoomSettings holds per-room configuration.
type RoomSettings struct {
	RoomID                    int64 `json:"room_id" db:"room_id"`
	MaxUsers                  int   `json:"max_users" db:"max_users"`
	RecordingEnabled          bool  `json:"recording_enabled" db:"recording_enabled"`
	AllowStudentVideo         bool  `json:"allow_student_video" db:"allow_student_video"`
	AllowStudentAudio         bool  `json:"allow_student_audio" db:"allow_student_audio"`
	AllowStudentScreenShare   bool  `json:"allow_student_screen_share" db:"allow_student_screen_share"`
	AllowStudentWhiteboard    bool  `json:"allow_student_whiteboard" db:"allow_student_whiteboard"`
	AllowStudentChat          bool  `json:"allow_student_chat" db:"allow_student_chat"`
	SessionAutoEndMinutes     int   `json:"session_auto_end_minutes" db:"session_auto_end_minutes"`
	WaitingRoomEnabled        bool  `json:"waiting_room_enabled" db:"waiting_room_enabled"`
}
