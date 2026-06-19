package repository

import (
	"database/sql"

	"github.com/iroom/iroom/internal/domain/entity"
)

type SessionLogRepo struct {
	db *sql.DB
}

func NewSessionLogRepo(db *sql.DB) *SessionLogRepo {
	return &SessionLogRepo{db: db}
}

func (r *SessionLogRepo) Create(l *entity.SessionLog) error {
	result, err := r.db.Exec(
		`INSERT INTO session_logs (session_id, user_id, joined_at, ip_address) VALUES (?, ?, ?, ?)`,
		l.SessionID, l.UserID, l.JoinedAt, l.IPAddress,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	l.ID = id
	return nil
}

func (r *SessionLogRepo) ListBySession(sessionID int64) ([]entity.SessionLog, error) {
	rows, err := r.db.Query(
		`SELECT sl.id, sl.session_id, sl.user_id, sl.joined_at, sl.left_at, sl.duration, sl.ip_address, u.display_name
		 FROM session_logs sl JOIN users u ON sl.user_id = u.id WHERE sl.session_id = ? ORDER BY sl.joined_at DESC`,
		sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []entity.SessionLog
	for rows.Next() {
		var l entity.SessionLog
		if err := rows.Scan(&l.ID, &l.SessionID, &l.UserID, &l.JoinedAt, &l.LeftAt, &l.Duration, &l.IPAddress, &l.UserDisplayName); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, nil
}

func (r *SessionLogRepo) GetUserTotalTime(userID int64) (int, error) {
	var total int
	err := r.db.QueryRow(`SELECT COALESCE(SUM(duration), 0) FROM session_logs WHERE user_id = ?`, userID).Scan(&total)
	return total, err
}

func (r *SessionLogRepo) UpdateLeave(id int64, leftAt string, duration int) error {
	_, err := r.db.Exec(
		`UPDATE session_logs SET left_at = ?, duration = ? WHERE id = ?`,
		leftAt, duration, id,
	)
	return err
}

func (r *SessionLogRepo) GetActiveLog(sessionID, userID int64) (*entity.SessionLog, error) {
	l := &entity.SessionLog{}
	err := r.db.QueryRow(
		`SELECT id, session_id, user_id, joined_at, left_at, duration, ip_address
		 FROM session_logs WHERE session_id = ? AND user_id = ? AND left_at IS NULL ORDER BY id DESC LIMIT 1`,
		sessionID, userID,
	).Scan(&l.ID, &l.SessionID, &l.UserID, &l.JoinedAt, &l.LeftAt, &l.Duration, &l.IPAddress)
	if err != nil {
		return nil, err
	}
	return l, nil
}
