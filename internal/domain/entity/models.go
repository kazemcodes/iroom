package entity

import "time"

// Message represents a chat message sent within a session.
// Messages are broadcast to all connected participants via WebSocket.
type Message struct {
	ID        int64     `json:"id" db:"id"`
	SessionID int64     `json:"session_id" db:"session_id"` // Parent session
	UserID    int64     `json:"user_id" db:"user_id"`       // Sender
	Content   string    `json:"content" db:"content"`       // Message text (max 10000 chars)
	Type      string    `json:"type" db:"type"`             // "text", "file", or "system"
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// File represents an uploaded file within a session.
// Files are stored on disk and referenced by filepath.
type File struct {
	ID         int64     `json:"id" db:"id"`
	SessionID  int64     `json:"session_id" db:"session_id"` // Parent session
	UploadedBy int64     `json:"uploaded_by" db:"uploaded_by"` // User who uploaded
	Filename   string    `json:"filename" db:"filename"`     // Original filename
	Filepath   string    `json:"filepath" db:"filepath"`     // Storage path on disk
	Filesize   int64     `json:"filesize" db:"filesize"`     // Size in bytes
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// Recording represents a session recording.
// Recordings are uploaded after a session ends and can be played back.
type Recording struct {
	ID         int64     `json:"id" db:"id"`
	SessionID  int64     `json:"session_id" db:"session_id"` // Parent session
	UploadedBy int64     `json:"uploaded_by" db:"uploaded_by"` // User who uploaded
	Filename   string    `json:"filename" db:"filename"`
	Filepath   string    `json:"filepath" db:"filepath"`     // Storage path
	Filesize   int64     `json:"filesize" db:"filesize"`
	Duration   int       `json:"duration" db:"duration"`     // Seconds
	Status     string    `json:"status" db:"status"`         // "processing", "ready", "failed"
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// Announcement represents a class announcement.
// Can be pinned to the top of the list or marked as system-wide.
type Announcement struct {
	ID           int64     `json:"id" db:"id"`
	ClassID      int64     `json:"class_id" db:"class_id"`   // Parent class (0 = system-wide)
	AuthorID     int64     `json:"author_id" db:"author_id"` // Teacher who posted
	Title        string    `json:"title" db:"title"`
	Content      string    `json:"content" db:"content"`
	IsPinned     bool      `json:"is_pinned" db:"is_pinned"`       // Pinned to top
	IsSystemWide bool      `json:"is_system_wide" db:"is_system_wide"` // Visible in all classes
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// Poll represents a session poll/vote.
// Options are stored as a comma-separated string.
type Poll struct {
	ID        int64     `json:"id" db:"id"`
	SessionID int64     `json:"session_id" db:"session_id"` // Parent session
	Question  string    `json:"question" db:"question"`
	Options   string    `json:"options" db:"options"` // Comma-separated options
	IsActive  bool      `json:"is_active" db:"is_active"` // false = closed
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// PollVote represents a single vote in a poll.
// Each user can vote once per poll.
type PollVote struct {
	ID          int64     `json:"id" db:"id"`
	PollID      int64     `json:"poll_id" db:"poll_id"`
	UserID      int64     `json:"user_id" db:"user_id"`     // Voter
	OptionIndex int       `json:"option_index" db:"option_index"` // 0-based index
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// PollResults holds aggregated vote results for a poll.
// Used by the frontend to display vote counts and percentages.
type PollResults struct {
	PollID     int64       `json:"poll_id"`
	Question   string      `json:"question"`
	Options    []string    `json:"options"`
	TotalVotes int         `json:"total_votes"`
	Votes      map[int]int `json:"votes"` // option_index → count
}

// Notification represents a user notification.
// Created for events like new messages, session starts, etc.
type Notification struct {
	ID        int64     `json:"id" db:"id"`
	UserID    int64     `json:"user_id" db:"user_id"` // Recipient
	Type      string    `json:"type" db:"type"`       // e.g. "message", "session_start"
	Title     string    `json:"title" db:"title"`
	Content   string    `json:"content" db:"content"`
	Message   string    `json:"message" db:"message"`
	Data      string    `json:"data" db:"data"`       // JSON-encoded extra data
	IsRead    bool      `json:"is_read" db:"is_read"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// SessionLog tracks user join/leave times for attendance.
// Used for the attendance report and session analytics.
type SessionLog struct {
	ID              int64      `json:"id" db:"id"`
	SessionID       int64      `json:"session_id" db:"session_id"`
	UserID          int64      `json:"user_id" db:"user_id"`
	JoinedAt        time.Time  `json:"joined_at" db:"joined_at"`
	LeftAt          *time.Time `json:"left_at" db:"left_at"` // nil = still connected
	Duration        int        `json:"duration" db:"duration"` // Seconds in session
	IPAddress       string     `json:"ip_address" db:"ip_address"`
	UserDisplayName string     `json:"user_display_name" db:"user_display_name"` // Joined from users
}

// Settings holds a key-value system configuration entry.
// Used for runtime settings like SMTP config, maintenance mode, etc.
type Settings struct {
	Key   string `json:"key" db:"key"`
	Value string `json:"value" db:"value"`
}

// ActivityLog records user actions for the audit trail.
// Captures who did what, when, and from which IP address.
type ActivityLog struct {
	ID         int64     `json:"id" db:"id"`
	UserID     int64     `json:"user_id" db:"user_id"`
	Action     string    `json:"action" db:"action"`         // e.g. "login", "create_class"
	EntityType string    `json:"entity_type" db:"entity_type"` // e.g. "user", "class"
	EntityID   int64     `json:"entity_id" db:"entity_id"`
	Details    string    `json:"details" db:"details"`       // Human-readable description
	IPAddress  string    `json:"ip_address" db:"ip_address"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// Webhook defines an HTTP callback URL that receives event notifications.
// Events are specified as a string array (e.g. ["session.started", "user.registered"]).
type Webhook struct {
	ID        int64     `json:"id" db:"id"`
	UserID    int64     `json:"user_id" db:"user_id"` // Owner
	URL       string    `json:"url" db:"url"`         // Target URL
	Secret    string    `json:"secret" db:"secret"`   // HMAC signing secret
	Events    []string  `json:"events" db:"events"`   // Subscribed event types
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// WebhookDelivery records the result of a webhook delivery attempt.
// Includes retry count, status code, and response for debugging.
type WebhookDelivery struct {
	ID           int64     `json:"id" db:"id"`
	WebhookID    int64     `json:"webhook_id" db:"webhook_id"` // Parent webhook
	EventType    string    `json:"event_type" db:"event_type"`
	Payload      string    `json:"payload" db:"payload"`           // JSON payload sent
	StatusCode   int       `json:"status_code" db:"status_code"`   // HTTP response code
	ResponseBody string    `json:"response_body" db:"response_body"` // HTTP response body
	Success      bool      `json:"success" db:"success"`
	RetryCount   int       `json:"retry_count" db:"retry_count"`   // 0 = first attempt
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// PasswordReset holds a password reset token.
// Tokens expire after 30 minutes and are single-use.
type PasswordReset struct {
	ID        int64     `json:"id" db:"id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	Token     string    `json:"token" db:"token"` // Secure random hex string
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// WebhookEvent is the payload sent to webhook endpoints.
// Wraps the event type, timestamp, and arbitrary data.
type WebhookEvent struct {
	Type      string      `json:"type"`      // e.g. "session.started"
	Timestamp time.Time   `json:"timestamp"` // When the event occurred
	Data      interface{} `json:"data"`      // Event-specific payload
}
