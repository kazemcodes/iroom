package repository

import (
	"database/sql"

	"github.com/iroom/iroom/internal/domain/entity"
)

type SessionRepo struct {
	db *sql.DB
}

func NewSessionRepo(db *sql.DB) *SessionRepo {
	return &SessionRepo{db: db}
}

func (r *SessionRepo) Create(s *entity.Session) error {
	result, err := r.db.Exec(
		`INSERT INTO sessions (room_id, class_id, title, scheduled_at, duration) VALUES (?, ?, ?, ?, ?)`,
		s.RoomID, s.ClassID, s.Title, s.ScheduledAt, s.Duration,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	s.ID = id
	return nil
}

func (r *SessionRepo) GetByID(id int64) (*entity.Session, error) {
	s := &entity.Session{}
	err := r.db.QueryRow(
		`SELECT id, COALESCE(room_id, 0), COALESCE(class_id, 0), title, scheduled_at, duration, status, livekit_room, recording_url, created_at, updated_at
		 FROM sessions WHERE id = ?`, id,
	).Scan(&s.ID, &s.RoomID, &s.ClassID, &s.Title, &s.ScheduledAt, &s.Duration,
		&s.Status, &s.LivekitRoom, &s.RecordingURL, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *SessionRepo) ListByRoom(roomID int64) ([]entity.Session, error) {
	rows, err := r.db.Query(
		`SELECT id, COALESCE(room_id, 0), COALESCE(class_id, 0), title, scheduled_at, duration, status, livekit_room, recording_url, created_at, updated_at
		 FROM sessions WHERE room_id = ? ORDER BY scheduled_at DESC`, roomID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []entity.Session
	for rows.Next() {
		var s entity.Session
		if err := rows.Scan(&s.ID, &s.RoomID, &s.ClassID, &s.Title, &s.ScheduledAt, &s.Duration,
			&s.Status, &s.LivekitRoom, &s.RecordingURL, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}

func (r *SessionRepo) ListByClass(classID int64) ([]entity.Session, error) {
	return r.ListByRoom(classID)
}

func (r *SessionRepo) ListAll(page, perPage int, search string) ([]entity.Session, int64, error) {
	var total int64
	countQuery := `SELECT COUNT(*) FROM sessions WHERE 1=1`
	args := []interface{}{}

	if search != "" {
		countQuery += ` AND title LIKE ?`
		args = append(args, "%"+search+"%")
	}

	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	query := `SELECT id, COALESCE(room_id, 0), COALESCE(class_id, 0), title, scheduled_at, duration, status, livekit_room, recording_url, created_at, updated_at FROM sessions WHERE 1=1`
	if search != "" {
		query += ` AND title LIKE ?`
	}
	query += ` ORDER BY scheduled_at DESC LIMIT ? OFFSET ?`
	args = append(args, perPage, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var sessions []entity.Session
	for rows.Next() {
		var s entity.Session
		if err := rows.Scan(&s.ID, &s.RoomID, &s.ClassID, &s.Title, &s.ScheduledAt, &s.Duration,
			&s.Status, &s.LivekitRoom, &s.RecordingURL, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, 0, err
		}
		sessions = append(sessions, s)
	}
	return sessions, total, nil
}

func (r *SessionRepo) UpdateStatus(id int64, status, livekitRoom string) error {
	_, err := r.db.Exec(
		`UPDATE sessions SET status = ?, livekit_room = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		status, livekitRoom, id,
	)
	return err
}

func (r *SessionRepo) UpdateRecordingURL(id int64, url string) error {
	_, err := r.db.Exec(
		`UPDATE sessions SET recording_url = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		url, id,
	)
	return err
}

func (r *SessionRepo) Count() (int64, error) {
	var count int64
	err := r.db.QueryRow(`SELECT COUNT(*) FROM sessions`).Scan(&count)
	return count, err
}

func (r *SessionRepo) CountActive() (int64, error) {
	var count int64
	err := r.db.QueryRow(`SELECT COUNT(*) FROM sessions WHERE status = 'live'`).Scan(&count)
	return count, err
}

func (r *SessionRepo) CountActiveByRoom(roomID int64) (int, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM sessions WHERE room_id = ? AND status = 'live'`, roomID).Scan(&count)
	return count, err
}

func (r *SessionRepo) Delete(id int64) error {
	_, err := r.db.Exec(`DELETE FROM sessions WHERE id = ?`, id)
	return err
}

func (r *SessionRepo) CreateRecurring(rs *entity.RecurringSession) error {
	result, err := r.db.Exec(
		`INSERT INTO recurring_sessions (class_id, title, day_of_week, start_time, duration, week_count) VALUES (?, ?, ?, ?, ?, ?)`,
		rs.ClassID, rs.Title, rs.DayOfWeek, rs.StartTime, rs.Duration, rs.WeekCount,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	rs.ID = id
	return nil
}

func (r *SessionRepo) ListRecurringByClass(classID int64) ([]entity.RecurringSession, error) {
	rows, err := r.db.Query(
		`SELECT id, class_id, title, day_of_week, start_time, duration, week_count, created_at FROM recurring_sessions WHERE class_id = ? ORDER BY day_of_week, start_time`, classID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []entity.RecurringSession
	for rows.Next() {
		var rs entity.RecurringSession
		if err := rows.Scan(&rs.ID, &rs.ClassID, &rs.Title, &rs.DayOfWeek, &rs.StartTime, &rs.Duration, &rs.WeekCount, &rs.CreatedAt); err != nil {
			return nil, err
		}
		sessions = append(sessions, rs)
	}
	return sessions, nil
}

func (r *SessionRepo) GetRecurringByID(id int64) (*entity.RecurringSession, error) {
	rs := &entity.RecurringSession{}
	err := r.db.QueryRow(
		`SELECT id, class_id, title, day_of_week, start_time, duration, week_count, created_at FROM recurring_sessions WHERE id = ?`, id,
	).Scan(&rs.ID, &rs.ClassID, &rs.Title, &rs.DayOfWeek, &rs.StartTime, &rs.Duration, &rs.WeekCount, &rs.CreatedAt)
	if err != nil {
		return nil, err
	}
	return rs, nil
}

func (r *SessionRepo) DeleteRecurring(id int64) error {
	_, err := r.db.Exec(`DELETE FROM recurring_sessions WHERE id = ?`, id)
	return err
}

func (r *SessionRepo) GetRoomBySessionID(sessionID int64) (*entity.Room, error) {
	room := &entity.Room{}
	err := r.db.QueryRow(
		`SELECT r.id, r.owner_id, r.name, r.description, r.color, r.slug, r.guest_login_enabled, r.max_users, r.invite_code, r.is_archived, r.created_at, r.updated_at
		 FROM rooms r
		 JOIN sessions s ON s.room_id = r.id
		 WHERE s.id = ?`, sessionID,
	).Scan(&room.ID, &room.OwnerID, &room.Name, &room.Description, &room.Color,
		&room.Slug, &room.GuestLoginEnabled, &room.MaxUsers, &room.InviteCode,
		&room.IsArchived, &room.CreatedAt, &room.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return room, nil
}

// IsSessionLive checks if a session is currently in 'live' status.
func (r *SessionRepo) IsSessionLive(sessionID int64) bool {
	var status string
	err := r.db.QueryRow(`SELECT status FROM sessions WHERE id = ?`, sessionID).Scan(&status)
	return err == nil && status == "live"
}

// GetAutoEndMinutesBySessionID returns the session_auto_end_minutes setting
// for the room associated with the given session. Returns 0 if disabled.
func (r *SessionRepo) GetAutoEndMinutesBySessionID(sessionID int64) int {
	var minutes int
	err := r.db.QueryRow(
		`SELECT COALESCE(rs.session_auto_end_minutes, 120)
		 FROM room_settings rs
		 JOIN sessions s ON s.room_id = rs.room_id
		 WHERE s.id = ?`, sessionID,
	).Scan(&minutes)
	if err != nil || minutes <= 0 {
		return 0
	}
	return minutes
}
