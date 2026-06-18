package entity

import "time"

type SessionStatus string

const (
	SessionScheduled SessionStatus = "scheduled"
	SessionLive      SessionStatus = "live"
	SessionEnded     SessionStatus = "ended"
)

type Session struct {
	ID           int64         `json:"id"`
	ClassID      int64         `json:"class_id"`
	Title        string        `json:"title"`
	ScheduledAt  time.Time     `json:"scheduled_at"`
	Duration     int           `json:"duration"`
	Status       SessionStatus `json:"status"`
	LivekitRoom  string        `json:"livekit_room"`
	RecordingURL string        `json:"recording_url"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

type RecurringSession struct {
	ID                int64     `json:"id"`
	ClassID           int64     `json:"class_id"`
	Title             string    `json:"title"`
	DayOfWeek         int       `json:"day_of_week"`
	StartTime         string    `json:"start_time"`
	Duration          int       `json:"duration"`
	WeekCount         int       `json:"week_count"`
	SessionsGenerated int       `json:"sessions_generated"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
