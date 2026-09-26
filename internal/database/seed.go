/**
 * Seed — Creates initial admin user on first database run.
 *
 * Only runs if the users table is empty. Uses "iroomteppasword123" as the default password.
 * The admin user can then log in at /auth and change it later.
 *
 * Default admin: admin@iroom.local / iroomteppasword123
 */
package database

import (
	"log/slog"

	"github.com/iroom/iroom/internal/pkg/hash"
)

const DefaultAdminPassword = "iroomteppasword123"

// Seed creates the initial admin user on the first run of an empty database.
// It is safe to call on every boot: it no-ops when users already exist.
func Seed(db DB) error {
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	hashedPassword, err := hash.Hash(DefaultAdminPassword)
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

	slog.Info("seeded admin user", "email", "admin@iroom.local", "default_password", DefaultAdminPassword)
	return nil
}
