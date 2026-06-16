package models

import "time"

type Ticket struct {
	ID              int64     `json:"id" db:"id"`
	UserID          int64     `json:"user_id" db:"user_id"`
	Title           string    `json:"title" db:"title"`
	Category        string    `json:"category" db:"category"`
	Status          string    `json:"status" db:"status"`
	Priority        string    `json:"priority" db:"priority"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
	UserDisplayName string    `json:"user_display_name" db:"user_display_name"`
}

type TicketMessage struct {
	ID              int64     `json:"id" db:"id"`
	TicketID        int64     `json:"ticket_id" db:"ticket_id"`
	UserID          int64     `json:"user_id" db:"user_id"`
	Content         string    `json:"content" db:"content"`
	IsAdmin         bool      `json:"is_admin" db:"is_admin"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UserDisplayName string    `json:"user_display_name" db:"user_display_name"`
}

type SessionLog struct {
	ID              int64      `json:"id" db:"id"`
	SessionID       int64      `json:"session_id" db:"session_id"`
	UserID          int64      `json:"user_id" db:"user_id"`
	JoinedAt        time.Time  `json:"joined_at" db:"joined_at"`
	LeftAt          *time.Time `json:"left_at" db:"left_at"`
	Duration        int        `json:"duration" db:"duration"`
	IPAddress       string     `json:"ip_address" db:"ip_address"`
	UserDisplayName string     `json:"user_display_name" db:"user_display_name"`
}

type CreateTicketRequest struct {
	Title    string `json:"title" validate:"required"`
	Message  string `json:"message"`
	Category string `json:"category"`
	Priority string `json:"priority"`
}

type ReplyTicketRequest struct {
	Content string `json:"content" validate:"required"`
}
