package entity

import "time"

// SessionStatus tracks the lifecycle of a class session (meeting).
type SessionStatus string

const (
	// SessionScheduled means the session is planned but not yet started.
	// Students cannot join until status changes to "live".
	SessionScheduled SessionStatus = "scheduled"

	// SessionLive means the session is active and participants can join.
	// WebRTC connections and chat are enabled in this state.
	SessionLive SessionStatus = "live"

	// SessionEnded means the session has concluded.
	// Recording playback is still available but no new connections allowed.
	SessionEnded SessionStatus = "ended"
)

// Session represents a single class meeting/session.
// Created by teachers, students join via the classroom popup.
//
// Business rules:
//   - Status transitions: scheduled → live → ended (no going back)
//   - Only the class owner (teacher) or admin can start/end a session
//   - livekitRoom is auto-generated as "room-{sessionID}" when started
//   - Students cannot join until status is "live"
type Session struct {
	ID           int64         `json:"id" db:"id"`
	ClassID      int64         `json:"class_id" db:"class_id"` // Parent class
	Title        string        `json:"title" db:"title"`       // Session name
	ScheduledAt  time.Time     `json:"scheduled_at" db:"scheduled_at"`
	Duration     int           `json:"duration" db:"duration"` // Minutes
	Status       SessionStatus `json:"status" db:"status"`
	LivekitRoom  string        `json:"livekit_room" db:"livekit_room"` // WebRTC room ID
	RecordingURL string        `json:"recording_url" db:"recording_url"` // Path to recording file
	CreatedAt    time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at" db:"updated_at"`
}

// RecurringSession defines a repeating session pattern.
// Used to auto-generate sessions on a weekly schedule.
//
// day_of_week: 0=شنبه (Saturday), 1=یکشنبه (Sunday), ..., 6=جمعه (Friday)
type RecurringSession struct {
	ID                int64     `json:"id" db:"id"`
	ClassID           int64     `json:"class_id" db:"class_id"`
	Title             string    `json:"title" db:"title"`
	DayOfWeek         int       `json:"day_of_week" db:"day_of_week"` // 0=Sat, 6=Fri
	StartTime         string    `json:"start_time" db:"start_time"`   // HH:MM format
	Duration          int       `json:"duration" db:"duration"`       // Minutes
	WeekCount         int       `json:"week_count" db:"week_count"`   // How many weeks to generate
	SessionsGenerated int       `json:"sessions_generated" db:"sessions_generated"` // Count of generated sessions
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}
