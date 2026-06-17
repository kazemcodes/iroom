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

func (r *FileRepo) Delete(id int64) error {
	_, err := r.db.Exec(`DELETE FROM files WHERE id = ?`, id)
	return err
}

func (r *FileRepo) Search(query string, limit, offset int) ([]models.File, int64, error) {
	var total int64
	countQuery := `SELECT COUNT(*) FROM files WHERE filename LIKE ?`
	s := "%" + query + "%"
	if err := r.db.QueryRow(countQuery, s).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(
		`SELECT id, session_id, uploaded_by, filename, filepath, filesize, created_at FROM files WHERE filename LIKE ? ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		s, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var files []models.File
	for rows.Next() {
		var f models.File
		if err := rows.Scan(&f.ID, &f.SessionID, &f.UploadedBy, &f.Filename, &f.Filepath, &f.Filesize, &f.CreatedAt); err != nil {
			return nil, 0, err
		}
		files = append(files, f)
	}
	return files, total, nil
}
