/**
 * Seed — Creates initial admin user on first database run.
 *
 * Only runs if the users table is empty. Generates a random 12-character
 * password and logs it to stdout. The admin user can then log in at /auth.
 *
 * Default admin: admin@iroom.local (password shown in logs)
 */
package database

import (
	"crypto/rand"
	"database/sql"
	"fmt"
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

	password := generateRandomPassword(12)
	hashedPassword, err := hash.Hash(password)
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

	slog.Info("seeded admin user", "email", "admin@iroom.local", "initial_password", password)
	return nil
}

func generateRandomPassword(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%"
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "Admin@2024Secure"
	}
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return fmt.Sprintf("%s", b)
}
