package repository

import (
	"database/sql"

	"github.com/iroom/iroom/internal/models"
)

type TicketRepo struct {
	db *sql.DB
}

func NewTicketRepo(db *sql.DB) *TicketRepo {
	return &TicketRepo{db: db}
}

func (r *TicketRepo) Create(t *models.Ticket) error {
	result, err := r.db.Exec(
		`INSERT INTO tickets (user_id, title, category, priority) VALUES (?, ?, ?, ?)`,
		t.UserID, t.Title, t.Category, t.Priority,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	t.ID = id
	return nil
}

func (r *TicketRepo) GetByID(id int64) (*models.Ticket, error) {
	t := &models.Ticket{}
	err := r.db.QueryRow(
		`SELECT t.id, t.user_id, t.title, t.category, t.status, t.priority, t.created_at, t.updated_at, u.display_name
		 FROM tickets t JOIN users u ON t.user_id = u.id WHERE t.id = ?`, id,
	).Scan(&t.ID, &t.UserID, &t.Title, &t.Category, &t.Status, &t.Priority, &t.CreatedAt, &t.UpdatedAt, &t.UserDisplayName)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *TicketRepo) ListByUser(userID int64, page, perPage int) ([]models.Ticket, int64, error) {
	var total int64
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM tickets WHERE user_id = ?`, userID).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	rows, err := r.db.Query(
		`SELECT t.id, t.user_id, t.title, t.category, t.status, t.priority, t.created_at, t.updated_at, u.display_name
		 FROM tickets t JOIN users u ON t.user_id = u.id WHERE t.user_id = ? ORDER BY t.id DESC LIMIT ? OFFSET ?`,
		userID, perPage, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tickets []models.Ticket
	for rows.Next() {
		var t models.Ticket
		if err := rows.Scan(&t.ID, &t.UserID, &t.Title, &t.Category, &t.Status, &t.Priority, &t.CreatedAt, &t.UpdatedAt, &t.UserDisplayName); err != nil {
			return nil, 0, err
		}
		tickets = append(tickets, t)
	}
	return tickets, total, nil
}

func (r *TicketRepo) ListAll(page, perPage int, search string) ([]models.Ticket, int64, error) {
	var total int64
	countQuery := `SELECT COUNT(*) FROM tickets WHERE 1=1`
	args := []interface{}{}

	if search != "" {
		countQuery += ` AND title LIKE ?`
		args = append(args, "%"+search+"%")
	}

	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	query := `SELECT t.id, t.user_id, t.title, t.category, t.status, t.priority, t.created_at, t.updated_at, u.display_name
		 FROM tickets t JOIN users u ON t.user_id = u.id WHERE 1=1`
	if search != "" {
		query += ` AND t.title LIKE ?`
	}
	query += ` ORDER BY t.id DESC LIMIT ? OFFSET ?`
	args = append(args, perPage, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tickets []models.Ticket
	for rows.Next() {
		var t models.Ticket
		if err := rows.Scan(&t.ID, &t.UserID, &t.Title, &t.Category, &t.Status, &t.Priority, &t.CreatedAt, &t.UpdatedAt, &t.UserDisplayName); err != nil {
			return nil, 0, err
		}
		tickets = append(tickets, t)
	}
	return tickets, total, nil
}

func (r *TicketRepo) UpdateStatus(id int64, status string) error {
	_, err := r.db.Exec(
		`UPDATE tickets SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		status, id,
	)
	return err
}

func (r *TicketRepo) SendMessage(m *models.TicketMessage) error {
	result, err := r.db.Exec(
		`INSERT INTO ticket_messages (ticket_id, user_id, content, is_admin) VALUES (?, ?, ?, ?)`,
		m.TicketID, m.UserID, m.Content, m.IsAdmin,
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

func (r *TicketRepo) ListMessages(ticketID int64) ([]models.TicketMessage, error) {
	rows, err := r.db.Query(
		`SELECT m.id, m.ticket_id, m.user_id, m.content, m.is_admin, m.created_at, u.display_name
		 FROM ticket_messages m JOIN users u ON m.user_id = u.id WHERE m.ticket_id = ? ORDER BY m.created_at ASC`,
		ticketID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []models.TicketMessage
	for rows.Next() {
		var m models.TicketMessage
		if err := rows.Scan(&m.ID, &m.TicketID, &m.UserID, &m.Content, &m.IsAdmin, &m.CreatedAt, &m.UserDisplayName); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, nil
}

type SessionLogRepo struct {
	db *sql.DB
}

func NewSessionLogRepo(db *sql.DB) *SessionLogRepo {
	return &SessionLogRepo{db: db}
}

func (r *SessionLogRepo) Create(l *models.SessionLog) error {
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

func (r *SessionLogRepo) ListBySession(sessionID int64) ([]models.SessionLog, error) {
	rows, err := r.db.Query(
		`SELECT sl.id, sl.session_id, sl.user_id, sl.joined_at, sl.left_at, sl.duration, sl.ip_address, u.display_name
		 FROM session_logs sl JOIN users u ON sl.user_id = u.id WHERE sl.session_id = ? ORDER BY sl.joined_at DESC`,
		sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.SessionLog
	for rows.Next() {
		var l models.SessionLog
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

func (r *SessionLogRepo) GetActiveLog(sessionID, userID int64) (*models.SessionLog, error) {
	l := &models.SessionLog{}
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
