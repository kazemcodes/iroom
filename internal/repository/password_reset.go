package repository

import (
	"database/sql"
	"time"
)

type PasswordResetRepo struct {
	db *sql.DB
}

func NewPasswordResetRepo(db *sql.DB) *PasswordResetRepo {
	return &PasswordResetRepo{db: db}
}

func (r *PasswordResetRepo) Create(userID int64, token string, expiresAt time.Time) error {
	_, err := r.db.Exec(
		`INSERT INTO password_resets (user_id, token, expires_at) VALUES (?, ?, ?)`,
		userID, token, expiresAt,
	)
	return err
}

func (r *PasswordResetRepo) GetByToken(token string) (int64, time.Time, error) {
	var userID int64
	var expiresAt time.Time
	err := r.db.QueryRow(
		`SELECT user_id, expires_at FROM password_resets WHERE token = ? AND used_at IS NULL`, token,
	).Scan(&userID, &expiresAt)
	if err != nil {
		return 0, time.Time{}, err
	}
	return userID, expiresAt, nil
}

func (r *PasswordResetRepo) MarkUsed(token string) error {
	_, err := r.db.Exec(
		`UPDATE password_resets SET used_at = CURRENT_TIMESTAMP WHERE token = ?`,
		token,
	)
	return err
}
