package repository

import (
	"database/sql"

	"github.com/iroom/iroom/internal/models"
)

type SessionRepo struct {
	db *sql.DB
}

func NewSessionRepo(db *sql.DB) *SessionRepo {
	return &SessionRepo{db: db}
}

func (r *SessionRepo) Create(s *models.Session) error {
	result, err := r.db.Exec(
		`INSERT INTO sessions (class_id, title, scheduled_at, duration) VALUES (?, ?, ?, ?)`,
		s.ClassID, s.Title, s.ScheduledAt, s.Duration,
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

func (r *SessionRepo) GetByID(id int64) (*models.Session, error) {
	s := &models.Session{}
	err := r.db.QueryRow(
		`SELECT id, class_id, title, scheduled_at, duration, status, livekit_room, recording_url, created_at, updated_at FROM sessions WHERE id = ?`, id,
	).Scan(&s.ID, &s.ClassID, &s.Title, &s.ScheduledAt, &s.Duration, &s.Status, &s.LivekitRoom, &s.RecordingURL, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *SessionRepo) ListByClass(classID int64) ([]models.Session, error) {
	rows, err := r.db.Query(
		`SELECT id, class_id, title, scheduled_at, duration, status, livekit_room, recording_url, created_at, updated_at FROM sessions WHERE class_id = ? ORDER BY scheduled_at DESC`, classID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []models.Session
	for rows.Next() {
		var s models.Session
		if err := rows.Scan(&s.ID, &s.ClassID, &s.Title, &s.ScheduledAt, &s.Duration, &s.Status, &s.LivekitRoom, &s.RecordingURL, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}

func (r *SessionRepo) ListAll(page, perPage int, search string) ([]models.Session, int64, error) {
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
	query := `SELECT id, class_id, title, scheduled_at, duration, status, livekit_room, recording_url, created_at, updated_at FROM sessions WHERE 1=1`
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

	var sessions []models.Session
	for rows.Next() {
		var s models.Session
		if err := rows.Scan(&s.ID, &s.ClassID, &s.Title, &s.ScheduledAt, &s.Duration, &s.Status, &s.LivekitRoom, &s.RecordingURL, &s.CreatedAt, &s.UpdatedAt); err != nil {
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

func (r *SessionRepo) Delete(id int64) error {
	_, err := r.db.Exec(`DELETE FROM sessions WHERE id = ?`, id)
	return err
}

// Recurring session methods

func (r *SessionRepo) CreateRecurring(rs *models.RecurringSession) error {
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

func (r *SessionRepo) ListRecurringByClass(classID int64) ([]models.RecurringSession, error) {
	rows, err := r.db.Query(
		`SELECT id, class_id, title, day_of_week, start_time, duration, week_count, created_at FROM recurring_sessions WHERE class_id = ? ORDER BY day_of_week, start_time`, classID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []models.RecurringSession
	for rows.Next() {
		var rs models.RecurringSession
		if err := rows.Scan(&rs.ID, &rs.ClassID, &rs.Title, &rs.DayOfWeek, &rs.StartTime, &rs.Duration, &rs.WeekCount, &rs.CreatedAt); err != nil {
			return nil, err
		}
		sessions = append(sessions, rs)
	}
	return sessions, nil
}

func (r *SessionRepo) GetRecurringByID(id int64) (*models.RecurringSession, error) {
	rs := &models.RecurringSession{}
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

func (r *SessionRepo) GetClassBySessionID(sessionID int64) (*models.Class, error) {
	class := &models.Class{}
	err := r.db.QueryRow(
		`SELECT c.id, c.teacher_id, c.name, c.description, c.color, c.max_students, c.invite_code, c.is_archived, c.created_at, c.updated_at 
		 FROM classes c 
		 JOIN sessions s ON s.class_id = c.id 
		 WHERE s.id = ?`, sessionID,
	).Scan(&class.ID, &class.TeacherID, &class.Name, &class.Description, &class.Color, &class.MaxStudents, &class.InviteCode, &class.IsArchived, &class.CreatedAt, &class.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return class, nil
}
