package database

import (
	"embed"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

//go:embed migrations/*.sql
var sqliteMigrationsFS embed.FS

// openSQLite opens (creating if needed) the SQLite database file, enables WAL
// and foreign keys, and applies the embedded SQLite migrations.
//
// SQLite is the default driver: it needs no external service, which makes it
// ideal for local development and single-node deployments.
func openSQLite(cfg Config) (DB, error) {
	dbPath := cfg.Path
	if dbPath == "" {
		dbPath = "iroom.db"
	}

	// Ensure the parent directory exists (":memory:" and bare filenames skip this).
	if dir := filepath.Dir(dbPath); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("create db dir: %w", err)
		}
	}

	dsn := dbPath + "?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=ON"
	db, err := openSQL("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	// SQLite tolerates only a single writer; a pool of one avoids
	// "database is locked" errors under concurrent requests.
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0) // SQLite connections don't benefit from recycling

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	conn := NewConn(db, DriverSQLite)
	if err := Migrate(conn, sqliteMigrationsFS, "migrations"); err != nil {
		db.Close()
		return nil, fmt.Errorf("sqlite migrations: %w", err)
	}

	slog.Info("database initialized", "driver", DriverSQLite, "path", dbPath)
	return conn, nil
}
