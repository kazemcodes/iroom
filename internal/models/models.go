package models

import "time"

type User struct {
	ID              int64     `json:"id" db:"id"`
	Email           string    `json:"email" db:"email"`
	PasswordHash    string    `json:"-" db:"password_hash"`
	DisplayName     string    `json:"display_name" db:"display_name"`
	Role            string    `json:"role" db:"role"`
	Phone           string    `json:"phone" db:"phone"`
	AvatarURL       string    `json:"avatar_url,omitempty" db:"avatar_url"`
	IsActive        bool      `json:"is_active" db:"is_active"`
	TOTPSecret      string    `json:"-" db:"totp_secret"`
	TOTPEnabled     bool      `json:"totp_enabled" db:"totp_enabled"`
	TOTPBackupCodes string    `json:"-" db:"totp_backup_codes"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

type Class struct {
	ID           int64     `json:"id" db:"id"`
	TeacherID    int64     `json:"teacher_id" db:"teacher_id"`
	Name         string    `json:"name" db:"name"`
	Description  string    `json:"description" db:"description"`
	Color        string    `json:"color" db:"color"`
	MaxStudents  int       `json:"max_students" db:"max_students"`
	InviteCode   string    `json:"invite_code,omitempty" db:"invite_code"`
	IsArchived   bool      `json:"is_archived" db:"is_archived"`
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

type Notification struct {
	ID        int64     `json:"id" db:"id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	Type      string    `json:"type" db:"type"`
	Title     string    `json:"title" db:"title"`
	Message   string    `json:"message,omitempty" db:"message"`
	Data      string    `json:"data,omitempty" db:"data"`
	IsRead    bool      `json:"is_read" db:"is_read"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type Announcement struct {
	ID           int64     `json:"id" db:"id"`
	ClassID      *int64    `json:"class_id,omitempty" db:"class_id"`
	AuthorID     int64     `json:"author_id" db:"author_id"`
	Title        string    `json:"title" db:"title"`
	Content      string    `json:"content" db:"content"`
	IsPinned     bool      `json:"is_pinned" db:"is_pinned"`
	IsSystemWide bool      `json:"is_system_wide" db:"is_system_wide"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type RecurringSession struct {
	ID         int64     `json:"id" db:"id"`
	ClassID    int64     `json:"class_id" db:"class_id"`
	Title      string    `json:"title" db:"title"`
	DayOfWeek  int       `json:"day_of_week" db:"day_of_week"`
	StartTime  string    `json:"start_time" db:"start_time"`
	Duration   int       `json:"duration" db:"duration"`
	WeekCount  int       `json:"week_count" db:"week_count"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

type CreateAnnouncementRequest struct {
	Title    string `json:"title" validate:"required"`
	Content  string `json:"content" validate:"required"`
	IsPinned bool   `json:"is_pinned"`
}

type UpdateAnnouncementRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type CreateRecurringSessionRequest struct {
	ClassID   int64  `json:"class_id" validate:"required"`
	Title     string `json:"title" validate:"required"`
	DayOfWeek int    `json:"day_of_week" validate:"required,min=0,max=6"`
	StartTime string `json:"start_time" validate:"required"`
	Duration  int    `json:"duration"`
	WeekCount int    `json:"week_count"`
}

type Poll struct {
	ID        int64     `json:"id" db:"id"`
	SessionID int64     `json:"session_id" db:"session_id"`
	Question  string    `json:"question" db:"question"`
	Options   []string  `json:"options" db:"-"`
	OptionsJSON string  `json:"-" db:"options"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type PollVote struct {
	ID          int64     `json:"id" db:"id"`
	PollID      int64     `json:"poll_id" db:"poll_id"`
	UserID      int64     `json:"user_id" db:"user_id"`
	OptionIndex int       `json:"option_index" db:"option_index"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type PollResults struct {
	PollID     int64    `json:"poll_id"`
	Question   string   `json:"question"`
	Options    []string `json:"options"`
	Votes      []int    `json:"votes"` // count per option
	TotalVotes int      `json:"total_votes"`
}

type CreatePollRequest struct {
	Question string   `json:"question" validate:"required"`
	Options  []string `json:"options" validate:"required,min=2"`
}

type VoteRequest struct {
	OptionIndex int `json:"option_index" validate:"required,min=0"`
}

type PaginatedResponse struct {
	Items      interface{} `json:"items"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PerPage    int         `json:"per_page"`
	TotalPages int         `json:"total_pages"`
}

// Webhook models

type Webhook struct {
	ID        int64     `json:"id" db:"id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	URL       string    `json:"url" db:"url"`
	Secret    string    `json:"-" db:"secret"`
	Events    []string  `json:"events" db:"-"`
	EventsJSON string   `json:"-" db:"events"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type WebhookDelivery struct {
	ID           int64     `json:"id" db:"id"`
	WebhookID    int64     `json:"webhook_id" db:"webhook_id"`
	EventType    string    `json:"event_type" db:"event_type"`
	Payload      string    `json:"payload" db:"payload"`
	StatusCode   *int      `json:"status_code,omitempty" db:"status_code"`
	ResponseBody string    `json:"response_body,omitempty" db:"response_body"`
	Success      bool      `json:"success" db:"success"`
	RetryCount   int       `json:"retry_count" db:"retry_count"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

type CreateWebhookRequest struct {
	URL    string   `json:"url" validate:"required,url"`
	Events []string `json:"events" validate:"required,min=1"`
}

type UpdateWebhookRequest struct {
	URL      string   `json:"url,omitempty" validate:"omitempty,url"`
	Events   []string `json:"events,omitempty" validate:"omitempty,min=1"`
	IsActive *bool    `json:"is_active,omitempty"`
}

type WebhookEvent struct {
	Type      string      `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// Webhook event types
const (
	WebhookEventSessionStarted  = "session.started"
	WebhookEventSessionEnded    = "session.ended"
	WebhookEventUserRegistered  = "user.registered"
	WebhookEventTicketCreated   = "ticket.created"
)
