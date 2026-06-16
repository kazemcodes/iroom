package database

import (
	"database/sql"
	"log/slog"

	"github.com/iroom/iroom/internal/pkg/hash"
)

func Seed(db *sql.DB) error {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	hashedPassword, err := hash.Hash("admin123")
	if err != nil {
		return err
	}

	_, err = db.Exec(
		`INSERT INTO users (email, password_hash, display_name, role, phone) VALUES (?, ?, ?, ?, ?)`,
		"admin@iroom.local", hashedPassword, "مدیر سیستم", "admin", "09120000000",
	)
	if err != nil {
		return err
	}

	slog.Info("seeded admin user", "email", "admin@iroom.local")
	return nil
}
