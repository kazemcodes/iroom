package database

import (
	"database/sql"
	"embed"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

//go:embed migrations/*.sql
var sqliteMigrationsFS embed.FS

// sqliteDriverName is the database/sql driver name registered by modernc.org/sqlite.
const sqliteDriverName = "sqlite"

// sqliteDSN builds the connection string for a SQLite database file.
//
// The parameter style works with modernc.org/sqlite and keeps WAL journaling
// plus foreign key enforcement enabled, matching the previous cgo driver.
func sqliteDSN(path string) string {
	return path + "?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=ON"
}

// NewTestDB returns a private in-memory SQLite connection for tests.
//
// It applies the same pragmas as production (so foreign key and cascade
// behaviour matches what the repositories rely on) but skips the migrations,
// letting each test define exactly the schema it needs. A private cache keeps
// separate tests from sharing one ":memory:" database.
func NewTestDB() (*sql.DB, error) {
	dsn := "file::memory:?cache=private" +
		"&_journal_mode=MEMORY" + // WAL is unavailable for in-memory databases
		"&_busy_timeout=5000" +
		"&_foreign_keys=ON"
	return sql.Open(sqliteDriverName, dsn)
}

// openSQLite opens (creating if needed) the SQLite database file, enables WAL
// and foreign keys, and applies the embedded SQLite migrations.
//
// SQLite is the default driver: it needs no external service, which makes it
// ideal for local development and single-node deployments. It is implemented
// with modernc.org/sqlite, a pure-Go translation of SQLite, so the project
// builds without CGO (no C toolchain required) and cross-compiles cleanly.
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

	db, err := openSQL(sqliteDriverName, sqliteDSN(dbPath))
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
