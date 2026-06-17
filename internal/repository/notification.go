package repository

import (
	"database/sql"

	"github.com/iroom/iroom/internal/models"
)

type NotificationRepo struct {
	db *sql.DB
}

func NewNotificationRepo(db *sql.DB) *NotificationRepo {
	return &NotificationRepo{db: db}
}

func (r *NotificationRepo) Create(n *models.Notification) error {
	result, err := r.db.Exec(
		`INSERT INTO notifications (user_id, type, title, message, data) VALUES (?, ?, ?, ?, ?)`,
		n.UserID, n.Type, n.Title, n.Message, n.Data,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	n.ID = id
	return nil
}

func (r *NotificationRepo) ListByUser(userID int64, limit, offset int) ([]*models.Notification, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, type, title, message, data, is_read, created_at FROM notifications WHERE user_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []*models.Notification
	for rows.Next() {
		n := &models.Notification{}
		if err := rows.Scan(&n.ID, &n.UserID, &n.Type, &n.Title, &n.Message, &n.Data, &n.IsRead, &n.CreatedAt); err != nil {
			return nil, err
		}
		notifications = append(notifications, n)
	}
	return notifications, nil
}

func (r *NotificationRepo) MarkRead(id, userID int64) error {
	_, err := r.db.Exec(`UPDATE notifications SET is_read = TRUE WHERE id = ? AND user_id = ?`, id, userID)
	return err
}

func (r *NotificationRepo) MarkAllRead(userID int64) error {
	_, err := r.db.Exec(`UPDATE notifications SET is_read = TRUE WHERE user_id = ?`, userID)
	return err
}

func (r *NotificationRepo) CountByUser(userID int64) (int64, error) {
	var count int64
	err := r.db.QueryRow(`SELECT COUNT(*) FROM notifications WHERE user_id = ?`, userID).Scan(&count)
	return count, err
}

func (r *NotificationRepo) CountUnread(userID int64) (int64, error) {
	var count int64
	err := r.db.QueryRow(`SELECT COUNT(*) FROM notifications WHERE user_id = ? AND is_read = FALSE`, userID).Scan(&count)
	return count, err
}
