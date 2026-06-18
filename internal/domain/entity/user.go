package entity

import "time"

// TokenClaims holds the data encoded inside JWT tokens.
// Used for authentication and authorization across the system.
type TokenClaims struct {
	UserID int64  // Database ID of the authenticated user
	Email  string // Email address of the user
	Role   string // User role: "admin", "teacher", or "student"
}

// UserRole defines the access level of a user in the system.
// Determines what features and resources a user can access.
type UserRole string

const (
	// RoleAdmin has full system access. Can manage all users, classes,
	// sessions, and system settings. Only role that can access the admin panel.
	RoleAdmin UserRole = "admin"

	// RoleTeacher can create and manage their own classes and sessions.
	// Has access to the admin panel for class/session management.
	RoleTeacher UserRole = "teacher"

	// RoleStudent can only view and join classes they are enrolled in.
	// Has no admin panel access. This is the default role for new registrations.
	RoleStudent UserRole = "student"
)

// User represents a system user (admin, teacher, or student).
// Users authenticate via email+password or guest login for class joining.
//
// Business rules:
//   - Email must be unique across the system
//   - Password is hashed with bcrypt before storage (never stored in plain text)
//   - Guest users get auto-generated emails like "guest_123_456@iroom.local"
//   - TOTP fields are optional; only populated when 2FA is enabled
//   - isActive=false soft-deletes the user without removing from database
type User struct {
	ID              int64     `json:"id" db:"id"`
	Email           string    `json:"email" db:"email"`
	PasswordHash    string    `json:"-" db:"password_hash"`           // Never exposed in JSON
	DisplayName     string    `json:"display_name" db:"display_name"` // Persian display name
	Role            UserRole  `json:"role" db:"role"`
	Phone           string    `json:"phone" db:"phone"`
	AvatarURL       string    `json:"avatar_url,omitempty" db:"avatar_url"`
	IsActive        bool      `json:"is_active" db:"is_active"` // false = soft-deleted
	TOTPSecret      string    `json:"-" db:"totp_secret"`       // Never exposed
	TOTPEnabled     bool      `json:"totp_enabled" db:"totp_enabled"`
	TOTPBackupCodes string    `json:"-" db:"totp_backup_codes"` // Never exposed
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}
