package repository

import (
	"database/sql"

	"github.com/iroom/iroom/internal/models"
)

type AnnouncementRepo struct {
	db *sql.DB
}

func NewAnnouncementRepo(db *sql.DB) *AnnouncementRepo {
	return &AnnouncementRepo{db: db}
}

func (r *AnnouncementRepo) Create(a *models.Announcement) error {
	result, err := r.db.Exec(
		`INSERT INTO announcements (class_id, author_id, title, content, is_pinned, is_system_wide) VALUES (?, ?, ?, ?, ?, ?)`,
		a.ClassID, a.AuthorID, a.Title, a.Content, a.IsPinned, a.IsSystemWide,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	a.ID = id
	return nil
}

func (r *AnnouncementRepo) GetByID(id int64) (*models.Announcement, error) {
	a := &models.Announcement{}
	err := r.db.QueryRow(
		`SELECT id, class_id, author_id, title, content, is_pinned, is_system_wide, created_at, updated_at FROM announcements WHERE id = ?`, id,
	).Scan(&a.ID, &a.ClassID, &a.AuthorID, &a.Title, &a.Content, &a.IsPinned, &a.IsSystemWide, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (r *AnnouncementRepo) ListByClass(classID int64, page, perPage int) ([]*models.Announcement, int64, error) {
	var total int64
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM announcements WHERE class_id = ? OR is_system_wide = TRUE`, classID,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	rows, err := r.db.Query(
		`SELECT id, class_id, author_id, title, content, is_pinned, is_system_wide, created_at, updated_at 
		 FROM announcements 
		 WHERE class_id = ? OR is_system_wide = TRUE 
		 ORDER BY is_pinned DESC, created_at DESC 
		 LIMIT ? OFFSET ?`,
		classID, perPage, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var announcements []*models.Announcement
	for rows.Next() {
		a := &models.Announcement{}
		if err := rows.Scan(&a.ID, &a.ClassID, &a.AuthorID, &a.Title, &a.Content, &a.IsPinned, &a.IsSystemWide, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, 0, err
		}
		announcements = append(announcements, a)
	}
	return announcements, total, nil
}

func (r *AnnouncementRepo) Update(a *models.Announcement) error {
	_, err := r.db.Exec(
		`UPDATE announcements SET title = ?, content = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		a.Title, a.Content, a.ID,
	)
	return err
}

func (r *AnnouncementRepo) Delete(id int64) error {
	_, err := r.db.Exec(`DELETE FROM announcements WHERE id = ?`, id)
	return err
}

func (r *AnnouncementRepo) SetPinned(id int64, pinned bool) error {
	_, err := r.db.Exec(
		`UPDATE announcements SET is_pinned = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		pinned, id,
	)
	return err
}
