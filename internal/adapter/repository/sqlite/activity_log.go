package repository

import (
	"database/sql"

	"github.com/iroom/iroom/internal/domain/entity"
)

type ActivityLogRepo struct {
	db *sql.DB
}

func NewActivityLogRepo(db *sql.DB) *ActivityLogRepo {
	return &ActivityLogRepo{db: db}
}

func (r *ActivityLogRepo) Create(log *entity.ActivityLog) error {
	result, err := r.db.Exec(
		`INSERT INTO activity_logs (user_id, action, entity_type, entity_id, details, ip_address) VALUES (?, ?, ?, ?, ?, ?)`,
		log.UserID, log.Action, log.EntityType, log.EntityID, log.Details, log.IPAddress,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	log.ID = id
	return nil
}

func (r *ActivityLogRepo) List(page, perPage int) ([]entity.ActivityLog, int64, error) {
	var total int64
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM activity_logs`).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	rows, err := r.db.Query(
		`SELECT id, user_id, action, entity_type, entity_id, details, ip_address, created_at FROM activity_logs ORDER BY id DESC LIMIT ? OFFSET ?`,
		perPage, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []entity.ActivityLog
	for rows.Next() {
		var l entity.ActivityLog
		if err := rows.Scan(&l.ID, &l.UserID, &l.Action, &l.EntityType, &l.EntityID, &l.Details, &l.IPAddress, &l.CreatedAt); err != nil {
			return nil, 0, err
		}
		logs = append(logs, l)
	}
	return logs, total, nil
}
