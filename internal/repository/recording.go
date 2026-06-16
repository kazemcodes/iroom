package repository

import (
	"database/sql"

	"github.com/iroom/iroom/internal/models"
)

type RecordingRepo struct {
	db *sql.DB
}

func NewRecordingRepo(db *sql.DB) *RecordingRepo {
	return &RecordingRepo{db: db}
}

func (r *RecordingRepo) Create(rec *models.Recording) error {
	result, err := r.db.Exec(
		`INSERT INTO recordings (session_id, uploaded_by, filename, filepath, filesize, duration, status) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		rec.SessionID, rec.UploadedBy, rec.Filename, rec.Filepath, rec.Filesize, rec.Duration, rec.Status,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	rec.ID = id
	return nil
}

func (r *RecordingRepo) GetByID(id int64) (*models.Recording, error) {
	rec := &models.Recording{}
	err := r.db.QueryRow(
		`SELECT id, session_id, uploaded_by, filename, filepath, filesize, duration, status, created_at FROM recordings WHERE id = ?`, id,
	).Scan(&rec.ID, &rec.SessionID, &rec.UploadedBy, &rec.Filename, &rec.Filepath, &rec.Filesize, &rec.Duration, &rec.Status, &rec.CreatedAt)
	if err != nil {
		return nil, err
	}
	return rec, nil
}

func (r *RecordingRepo) ListBySession(sessionID int64) ([]models.Recording, error) {
	rows, err := r.db.Query(
		`SELECT id, session_id, uploaded_by, filename, filepath, filesize, duration, status, created_at FROM recordings WHERE session_id = ? ORDER BY created_at DESC`, sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recordings []models.Recording
	for rows.Next() {
		var rec models.Recording
		if err := rows.Scan(&rec.ID, &rec.SessionID, &rec.UploadedBy, &rec.Filename, &rec.Filepath, &rec.Filesize, &rec.Duration, &rec.Status, &rec.CreatedAt); err != nil {
			return nil, err
		}
		recordings = append(recordings, rec)
	}
	return recordings, nil
}

func (r *RecordingRepo) ListAll(page, perPage int, search string) ([]models.Recording, int64, error) {
	var total int64
	countQuery := `SELECT COUNT(*) FROM recordings WHERE 1=1`
	args := []interface{}{}

	if search != "" {
		countQuery += ` AND filename LIKE ?`
		args = append(args, "%"+search+"%")
	}

	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	query := `SELECT id, session_id, uploaded_by, filename, filepath, filesize, duration, status, created_at FROM recordings WHERE 1=1`
	if search != "" {
		query += ` AND filename LIKE ?`
	}
	query += ` ORDER BY created_at DESC LIMIT ? OFFSET ?`
	args = append(args, perPage, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var recordings []models.Recording
	for rows.Next() {
		var rec models.Recording
		if err := rows.Scan(&rec.ID, &rec.SessionID, &rec.UploadedBy, &rec.Filename, &rec.Filepath, &rec.Filesize, &rec.Duration, &rec.Status, &rec.CreatedAt); err != nil {
			return nil, 0, err
		}
		recordings = append(recordings, rec)
	}
	return recordings, total, nil
}

func (r *RecordingRepo) Count() (int64, error) {
	var count int64
	err := r.db.QueryRow(`SELECT COUNT(*) FROM recordings`).Scan(&count)
	return count, err
}

func (r *RecordingRepo) Delete(id int64) error {
	_, err := r.db.Exec(`DELETE FROM recordings WHERE id = ?`, id)
	return err
}
