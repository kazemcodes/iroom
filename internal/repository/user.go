package repository

import (
	"database/sql"

	"github.com/iroom/iroom/internal/models"
)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(u *models.User) error {
	result, err := r.db.Exec(
		`INSERT INTO users (email, password_hash, display_name, role, phone) VALUES (?, ?, ?, ?, ?)`,
		u.Email, u.PasswordHash, u.DisplayName, u.Role, u.Phone,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	u.ID = id
	return nil
}

func (r *UserRepo) GetByID(id int64) (*models.User, error) {
	u := &models.User{}
	err := r.db.QueryRow(
		`SELECT id, email, password_hash, display_name, role, phone, is_active, created_at, updated_at FROM users WHERE id = ?`, id,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.DisplayName, &u.Role, &u.Phone, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *UserRepo) GetByEmail(email string) (*models.User, error) {
	u := &models.User{}
	err := r.db.QueryRow(
		`SELECT id, email, password_hash, display_name, role, phone, is_active, created_at, updated_at FROM users WHERE email = ?`, email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.DisplayName, &u.Role, &u.Phone, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *UserRepo) List(page, perPage int, search string) ([]models.User, int64, error) {
	var total int64
	query := `SELECT COUNT(*) FROM users WHERE 1=1`
	args := []interface{}{}

	if search != "" {
		query += ` AND (email LIKE ? OR display_name LIKE ? OR phone LIKE ?)`
		s := "%" + search + "%"
		args = append(args, s, s, s)
	}

	if err := r.db.QueryRow(query, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	query = `SELECT id, email, password_hash, display_name, role, phone, is_active, created_at, updated_at FROM users WHERE 1=1`
	if search != "" {
		query += ` AND (email LIKE ? OR display_name LIKE ? OR phone LIKE ?)`
	}
	query += ` ORDER BY id DESC LIMIT ? OFFSET ?`
	args = append(args, perPage, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.DisplayName, &u.Role, &u.Phone, &u.IsActive, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}
	return users, total, nil
}

func (r *UserRepo) Update(u *models.User) error {
	_, err := r.db.Exec(
		`UPDATE users SET display_name = ?, role = ?, phone = ?, is_active = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		u.DisplayName, u.Role, u.Phone, u.IsActive, u.ID,
	)
	return err
}

func (r *UserRepo) UpdatePassword(id int64, hash string) error {
	_, err := r.db.Exec(
		`UPDATE users SET password_hash = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		hash, id,
	)
	return err
}

func (r *UserRepo) UpdateAvatar(id int64, avatarURL string) error {
	_, err := r.db.Exec(
		`UPDATE users SET avatar_url = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		avatarURL, id,
	)
	return err
}

func (r *UserRepo) Count() (int64, error) {
	var count int64
	err := r.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&count)
	return count, err
}
