package repository

import (
	"database/sql"

	"github.com/iroom/iroom/internal/domain/entity"
)

type MessageRepo struct {
	db *sql.DB
}

func NewMessageRepo(db *sql.DB) *MessageRepo {
	return &MessageRepo{db: db}
}

func (r *MessageRepo) Create(m *entity.Message) error {
	result, err := r.db.Exec(
		`INSERT INTO messages (session_id, user_id, content, type) VALUES (?, ?, ?, ?)`,
		m.SessionID, m.UserID, m.Content, m.Type,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	m.ID = id
	return nil
}

func (r *MessageRepo) ListBySession(sessionID int64, limit, offset int) ([]entity.Message, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.db.Query(
		`SELECT id, session_id, user_id, content, type, created_at FROM messages WHERE session_id = ? ORDER BY id DESC LIMIT ? OFFSET ?`,
		sessionID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []entity.Message
	for rows.Next() {
		var m entity.Message
		if err := rows.Scan(&m.ID, &m.SessionID, &m.UserID, &m.Content, &m.Type, &m.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, nil
}

func (r *MessageRepo) Count() (int64, error) {
	var count int64
	err := r.db.QueryRow(`SELECT COUNT(*) FROM messages`).Scan(&count)
	return count, err
}
