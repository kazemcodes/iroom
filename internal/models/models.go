package models

import "time"

type User struct {
	ID           int64     `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	DisplayName  string    `json:"display_name" db:"display_name"`
	Role         string    `json:"role" db:"role"`
	Phone        string    `json:"phone" db:"phone"`
	IsActive     bool      `json:"is_active" db:"is_active"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type Class struct {
	ID           int64     `json:"id" db:"id"`
	TeacherID    int64     `json:"teacher_id" db:"teacher_id"`
	Name         string    `json:"name" db:"name"`
	Description  string    `json:"description" db:"description"`
	Color        string    `json:"color" db:"color"`
	MaxStudents  int       `json:"max_students" db:"max_students"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type ClassStudent struct {
	ClassID   int64 `json:"class_id" db:"class_id"`
	StudentID int64 `json:"student_id" db:"student_id"`
}

type Session struct {
	ID           int64     `json:"id" db:"id"`
	ClassID      int64     `json:"class_id" db:"class_id"`
	Title        string    `json:"title" db:"title"`
	ScheduledAt  time.Time `json:"scheduled_at" db:"scheduled_at"`
	Duration     int       `json:"duration" db:"duration"`
	Status       string    `json:"status" db:"status"`
	LivekitRoom  string    `json:"livekit_room" db:"livekit_room"`
	RecordingURL string    `json:"recording_url" db:"recording_url"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type Message struct {
	ID        int64     `json:"id" db:"id"`
	SessionID int64     `json:"session_id" db:"session_id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	Content   string    `json:"content" db:"content"`
	Type      string    `json:"type" db:"type"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type File struct {
	ID         int64     `json:"id" db:"id"`
	SessionID  int64     `json:"session_id" db:"session_id"`
	UploadedBy int64     `json:"uploaded_by" db:"uploaded_by"`
	Filename   string    `json:"filename" db:"filename"`
	Filepath   string    `json:"filepath" db:"filepath"`
	Filesize   int64     `json:"filesize" db:"filesize"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// Request/Response types

type RegisterRequest struct {
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,min=6"`
	DisplayName string `json:"display_name" validate:"required"`
	Phone       string `json:"phone"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type CreateClassRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	Color       string `json:"color"`
	MaxStudents int    `json:"max_students"`
}

type EnrollRequest struct {
	StudentID int64 `json:"student_id" validate:"required"`
}

type CreateSessionRequest struct {
	ClassID     int64  `json:"class_id" validate:"required"`
	Title       string `json:"title" validate:"required"`
	ScheduledAt string `json:"scheduled_at" validate:"required"`
	Duration    int    `json:"duration"`
}

type SendMessageRequest struct {
	Content string `json:"content" validate:"required"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type Recording struct {
	ID         int64     `json:"id" db:"id"`
	SessionID  int64     `json:"session_id" db:"session_id"`
	UploadedBy int64     `json:"uploaded_by" db:"uploaded_by"`
	Filename   string    `json:"filename" db:"filename"`
	Filepath   string    `json:"filepath" db:"filepath"`
	Filesize   int64     `json:"filesize" db:"filesize"`
	Duration   int       `json:"duration" db:"duration"`
	Status     string    `json:"status" db:"status"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

type PaginatedResponse struct {
	Items      interface{} `json:"items"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PerPage    int         `json:"per_page"`
	TotalPages int         `json:"total_pages"`
}
