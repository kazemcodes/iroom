package entity

import "time"

type Message struct {
	ID        int64     `json:"id"`
	SessionID int64     `json:"session_id"`
	UserID    int64     `json:"user_id"`
	Content   string    `json:"content"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
}

type File struct {
	ID         int64     `json:"id"`
	SessionID  int64     `json:"session_id"`
	UploadedBy int64     `json:"uploaded_by"`
	Filename   string    `json:"filename"`
	Filepath   string    `json:"filepath"`
	Filesize   int64     `json:"filesize"`
	CreatedAt  time.Time `json:"created_at"`
}

type Recording struct {
	ID         int64     `json:"id"`
	SessionID  int64     `json:"session_id"`
	UploadedBy int64     `json:"uploaded_by"`
	Filename   string    `json:"filename"`
	Filepath   string    `json:"filepath"`
	Filesize   int64     `json:"filesize"`
	Duration   int       `json:"duration"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

type Ticket struct {
	ID              int64     `json:"id"`
	UserID          int64     `json:"user_id"`
	Title           string    `json:"title"`
	Category        string    `json:"category"`
	Status          string    `json:"status"`
	Priority        string    `json:"priority"`
	UserDisplayName string    `json:"user_display_name"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type TicketMessage struct {
	ID              int64     `json:"id"`
	TicketID        int64     `json:"ticket_id"`
	UserID          int64     `json:"user_id"`
	Content         string    `json:"content"`
	IsAdmin         bool      `json:"is_admin"`
	UserDisplayName string    `json:"user_display_name"`
	CreatedAt       time.Time `json:"created_at"`
}

type Announcement struct {
	ID            int64     `json:"id"`
	ClassID       int64     `json:"class_id"`
	AuthorID      int64     `json:"author_id"`
	Title         string    `json:"title"`
	Content       string    `json:"content"`
	IsPinned      bool      `json:"is_pinned"`
	IsSystemWide  bool      `json:"is_system_wide"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Poll struct {
	ID        int64     `json:"id"`
	SessionID int64     `json:"session_id"`
	Question  string    `json:"question"`
	Options   string    `json:"options"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type PollVote struct {
	ID          int64     `json:"id"`
	PollID      int64     `json:"poll_id"`
	UserID      int64     `json:"user_id"`
	OptionIndex int       `json:"option_index"`
	CreatedAt   time.Time `json:"created_at"`
}

type PollResults struct {
	PollID     int64          `json:"poll_id"`
	Question   string         `json:"question"`
	Options    []string       `json:"options"`
	TotalVotes int            `json:"total_votes"`
	Votes      map[int]int    `json:"votes"`
}

type Notification struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Message   string    `json:"message"`
	Data      string    `json:"data"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}

type SessionLog struct {
	ID              int64     `json:"id"`
	SessionID       int64     `json:"session_id"`
	UserID          int64     `json:"user_id"`
	JoinedAt        time.Time `json:"joined_at"`
	LeftAt          *time.Time `json:"left_at"`
	Duration        int       `json:"duration"`
	IPAddress       string    `json:"ip_address"`
	UserDisplayName string    `json:"user_display_name"`
}

type Settings struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type ActivityLog struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	Action     string    `json:"action"`
	EntityType string    `json:"entity_type"`
	EntityID   int64     `json:"entity_id"`
	Details    string    `json:"details"`
	IPAddress  string    `json:"ip_address"`
	CreatedAt  time.Time `json:"created_at"`
}

type Webhook struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	URL       string    `json:"url"`
	Secret    string    `json:"secret"`
	Events    []string  `json:"events"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type WebhookDelivery struct {
	ID           int64     `json:"id"`
	WebhookID    int64     `json:"webhook_id"`
	EventType    string    `json:"event_type"`
	Payload      string    `json:"payload"`
	StatusCode   int       `json:"status_code"`
	ResponseBody string    `json:"response_body"`
	Success      bool      `json:"success"`
	RetryCount   int       `json:"retry_count"`
	CreatedAt    time.Time `json:"created_at"`
}

type PasswordReset struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}
