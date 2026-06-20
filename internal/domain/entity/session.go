package entity

import "time"

// SessionStatus tracks the lifecycle of a class session (meeting).
type SessionStatus string

const (
	SessionScheduled SessionStatus = "scheduled"
	SessionLive      SessionStatus = "live"
	SessionEnded     SessionStatus = "ended"
)

// Session represents a single room meeting/session.
// Created by teachers, students join via the classroom popup.
//
// Business rules:
//   - Status transitions: scheduled → live → ended (no going back)
//   - Only the room owner (teacher) or admin can start/end a session
//   - livekitRoom is auto-generated as "room-{roomID}" when started
//   - Students cannot join until status is "live"
//   - RoomID is the canonical parent reference; ClassID is kept for backward compat
type Session struct {
	ID           int64         `json:"id" db:"id"`
	RoomID       int64         `json:"room_id" db:"room_id"`   // Parent room
	ClassID      int64         `json:"class_id" db:"class_id"` // Deprecated: backward compat
	Title        string        `json:"title" db:"title"`
	ScheduledAt  time.Time     `json:"scheduled_at" db:"scheduled_at"`
	Duration     int           `json:"duration" db:"duration"` // Minutes
	Status       SessionStatus `json:"status" db:"status"`
	LivekitRoom  string        `json:"livekit_room" db:"livekit_room"`
	RecordingURL string        `json:"recording_url" db:"recording_url"`
	CreatedAt    time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at" db:"updated_at"`
}

// RecurringSession defines a repeating session pattern.
//
// day_of_week: 0=شنبه (Saturday), 1=یکشنبه (Sunday), ..., 6=جمعه (Friday)
type RecurringSession struct {
	ID                int64     `json:"id" db:"id"`
	ClassID           int64     `json:"class_id" db:"class_id"`
	Title             string    `json:"title" db:"title"`
	DayOfWeek         int       `json:"day_of_week" db:"day_of_week"`
	StartTime         string    `json:"start_time" db:"start_time"`
	Duration          int       `json:"duration" db:"duration"`
	WeekCount         int       `json:"week_count" db:"week_count"`
	SessionsGenerated int       `json:"sessions_generated" db:"sessions_generated"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}
