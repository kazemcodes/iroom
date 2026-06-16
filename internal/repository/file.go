package repository

import (
	"database/sql"

	"github.com/iroom/iroom/internal/models"
)

type FileRepo struct {
	db *sql.DB
}

func NewFileRepo(db *sql.DB) *FileRepo {
	return &FileRepo{db: db}
}

func (r *FileRepo) Create(f *models.File) error {
	result, err := r.db.Exec(
		`INSERT INTO files (session_id, uploaded_by, filename, filepath, filesize) VALUES (?, ?, ?, ?, ?)`,
		f.SessionID, f.UploadedBy, f.Filename, f.Filepath, f.Filesize,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	f.ID = id
	return nil
}

func (r *FileRepo) GetByID(id int64) (*models.File, error) {
	f := &models.File{}
	err := r.db.QueryRow(
		`SELECT id, session_id, uploaded_by, filename, filepath, filesize, created_at FROM files WHERE id = ?`, id,
	).Scan(&f.ID, &f.SessionID, &f.UploadedBy, &f.Filename, &f.Filepath, &f.Filesize, &f.CreatedAt)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (r *FileRepo) ListBySession(sessionID int64) ([]models.File, error) {
	rows, err := r.db.Query(
		`SELECT id, session_id, uploaded_by, filename, filepath, filesize, created_at FROM files WHERE session_id = ? ORDER BY created_at DESC`, sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []models.File
	for rows.Next() {
		var f models.File
		if err := rows.Scan(&f.ID, &f.SessionID, &f.UploadedBy, &f.Filename, &f.Filepath, &f.Filesize, &f.CreatedAt); err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	return files, nil
}
