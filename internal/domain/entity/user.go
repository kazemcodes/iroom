package entity

import "time"

type TokenClaims struct {
	UserID int64
	Email  string
	Role   string
}

type UserRole string

const (
	RoleAdmin   UserRole = "admin"
	RoleTeacher UserRole = "teacher"
	RoleStudent UserRole = "student"
)

type User struct {
	ID              int64     `json:"id"`
	Email           string    `json:"email"`
	PasswordHash    string    `json:"-"`
	DisplayName     string    `json:"display_name"`
	Role            UserRole  `json:"role"`
	Phone           string    `json:"phone"`
	AvatarURL       string    `json:"avatar_url,omitempty"`
	IsActive        bool      `json:"is_active"`
	TOTPSecret      string    `json:"-"`
	TOTPEnabled     bool      `json:"totp_enabled"`
	TOTPBackupCodes string    `json:"-"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
