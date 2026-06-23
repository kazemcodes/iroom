package repository

import (
	"database/sql"

	"github.com/iroom/iroom/internal/domain/entity"
)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(u *entity.User) error {
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

func (r *UserRepo) GetByID(id int64) (*entity.User, error) {
	u := &entity.User{}
	err := r.db.QueryRow(
		`SELECT id, email, password_hash, display_name, role, phone, is_active, COALESCE(totp_secret,''), totp_enabled, COALESCE(totp_backup_codes,''), created_at, updated_at FROM users WHERE id = ?`, id,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.DisplayName, &u.Role, &u.Phone, &u.IsActive, &u.TOTPSecret, &u.TOTPEnabled, &u.TOTPBackupCodes, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *UserRepo) GetByEmail(email string) (*entity.User, error) {
	u := &entity.User{}
	err := r.db.QueryRow(
		`SELECT id, email, password_hash, display_name, role, phone, is_active, COALESCE(totp_secret,''), totp_enabled, COALESCE(totp_backup_codes,''), created_at, updated_at FROM users WHERE email = ?`, email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.DisplayName, &u.Role, &u.Phone, &u.IsActive, &u.TOTPSecret, &u.TOTPEnabled, &u.TOTPBackupCodes, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *UserRepo) List(page, perPage int, search string) ([]entity.User, int64, error) {
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
	query = `SELECT id, email, password_hash, display_name, role, phone, is_active, COALESCE(totp_secret,''), totp_enabled, COALESCE(totp_backup_codes,''), created_at, updated_at FROM users WHERE 1=1`
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

	var users []entity.User
	for rows.Next() {
		var u entity.User
		if err := rows.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.DisplayName, &u.Role, &u.Phone, &u.IsActive, &u.TOTPSecret, &u.TOTPEnabled, &u.TOTPBackupCodes, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}
	return users, total, nil
}

func (r *UserRepo) Update(u *entity.User) error {
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

func (r *UserRepo) UpdateRole(id int64, role string) error {
	_, err := r.db.Exec(
		`UPDATE users SET role = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		role, id,
	)
	return err
}

func (r *UserRepo) Delete(id int64) error {
	_, err := r.db.Exec(`DELETE FROM users WHERE id = ?`, id)
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

// TOTP-related methods

func (r *UserRepo) UpdateTOTPSecret(id int64, secret string) error {
	_, err := r.db.Exec(
		`UPDATE users SET totp_secret = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		secret, id,
	)
	return err
}

func (r *UserRepo) EnableTOTP(id int64, backupCodes string) error {
	_, err := r.db.Exec(
		`UPDATE users SET totp_enabled = TRUE, totp_backup_codes = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		backupCodes, id,
	)
	return err
}

func (r *UserRepo) DisableTOTP(id int64) error {
	_, err := r.db.Exec(
		`UPDATE users SET totp_secret = NULL, totp_enabled = FALSE, totp_backup_codes = NULL, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		id,
	)
	return err
}

func (r *UserRepo) UpdateTOTPBackupCodes(id int64, backupCodes string) error {
	_, err := r.db.Exec(
		`UPDATE users SET totp_backup_codes = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		backupCodes, id,
	)
	return err
}
