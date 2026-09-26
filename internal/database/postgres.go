package database

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"sort"
	"strings"
	"time"
)

// postgresMigrations holds the Postgres schema. It is a single, idempotent
// baseline that mirrors the final state of the SQLite migrations.
//
// Postgres is the recommended driver for ephemeral/container hosting
// (Hugging Face Spaces, Fly.io, Render) where the local disk is wiped on
// every restart — the data lives in the external database instead.
//
//go:embed migrations/postgres/*.sql
var postgresMigrationsFS embed.FS

// openPostgres connects to an external PostgreSQL instance (Supabase, Neon,
// RDS, ...) and applies the embedded Postgres migrations.
func openPostgres(cfg Config) (DB, error) {
	dsn := cfg.URL
	if dsn == "" {
		return nil, errors.New("postgres: DATABASE_URL is required when DB_DRIVER=postgres")
	}
	dsn = applySSLMode(dsn, cfg.SSLMode)

	db, err := openSQL("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	// Postgres benefits from a real connection pool. Defaults are conservative
	// so the app stays within the small quotas of free hosting tiers, where
	// idle connections are the main way to exhaust the connection limit.
	maxOpen := cfg.MaxOpenConns
	if maxOpen <= 0 {
		maxOpen = 10
	}
	maxIdle := cfg.MaxIdleConns
	if maxIdle <= 0 {
		maxIdle = 5
	}
	if maxIdle > maxOpen {
		maxIdle = maxOpen
	}

	db.SetMaxOpenConns(maxOpen)
	db.SetMaxIdleConns(maxIdle)
	if cfg.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(secondsDuration(cfg.ConnMaxLifetime))
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	conn := NewConn(db, DriverPostgres)
	if err := Migrate(conn, postgresMigrationsFS, "migrations/postgres"); err != nil {
		db.Close()
		return nil, fmt.Errorf("postgres migrations: %w", err)
	}

	slog.Info("database initialized", "driver", DriverPostgres, "pool", maxOpen)
	return conn, nil
}

// Migrate applies every embedded migration that has not run yet.
//
// Migrations are plain .sql files applied in filename order and tracked in the
// schema_migrations table. Each file runs inside a transaction so a failure
// leaves the schema untouched.
func Migrate(db DB, migrationsFS fs.FS, dir string) error {
	// Bootstrap the migration tracking table.
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		filename   TEXT PRIMARY KEY,
		applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	entries, err := fs.ReadDir(migrationsFS, dir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && !isDotFile(entry.Name()) {
			names = append(names, entry.Name())
		}
	}
	// ReadDir already returns sorted entries, but sort explicitly so the
	// ordering contract ("applied in filename order") is guaranteed.
	sort.Strings(names)

	for _, name := range names {
		var exists int
		if err := db.QueryRow(
			`SELECT COUNT(*) FROM schema_migrations WHERE filename = ?`, name,
		).Scan(&exists); err != nil {
			return fmt.Errorf("check migration %s: %w", name, err)
		}
		if exists > 0 {
			slog.Debug("migration already applied, skipping", "file", name)
			continue
		}

		data, err := fs.ReadFile(migrationsFS, dir+"/"+name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}

		if err := applyMigration(db, name, string(data)); err != nil {
			return err
		}
		slog.Info("migration applied", "file", name, "driver", db.Driver())
	}

	return nil
}

// applyMigration runs a single migration atomically and records it as applied.
func applyMigration(db DB, name, script string) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin migration %s: %w", name, err)
	}
	// Rollback is a no-op once the transaction has been committed.
	defer tx.Rollback()

	// Migrations are authored per driver, so no placeholder rebinding here.
	if _, err := tx.Exec(script); err != nil {
		return fmt.Errorf("exec migration %s: %w", name, err)
	}

	if _, err := tx.Exec(
		`INSERT INTO schema_migrations (filename) VALUES (?)`, name,
	); err != nil {
		return fmt.Errorf("record migration %s: %w", name, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration %s: %w", name, err)
	}
	return nil
}

// isDotFile reports whether name is a dotfile (e.g. ".gitkeep").
func isDotFile(name string) bool {
	return len(name) > 0 && name[0] == '.'
}

// applySSLMode injects or overrides the sslmode parameter of a Postgres DSN.
//
// Hosted providers (Supabase, Neon) require TLS, but their published
// connection strings sometimes omit sslmode, so this makes it configurable
// without forcing users to hand-edit the URL. Accepts both URL-style DSNs
// (postgres://...) and key=value DSNs (host=... sslmode=...).
func applySSLMode(dsn, sslMode string) string {
	if sslMode == "" {
		return dsn
	}

	// Already specified: respect the explicit value in the DSN.
	if strings.Contains(dsn, "sslmode=") {
		return dsn
	}

	sep := "?"
	if strings.Contains(dsn, "?") {
		sep = "&"
	}
	return dsn + sep + "sslmode=" + sslMode
}

// openSQL opens a database/sql handle for the given driver and DSN.
func openSQL(driver, dsn string) (*sql.DB, error) {
	return sql.Open(driver, dsn)
}

// secondsDuration converts a whole-second count into a time.Duration.
func secondsDuration(seconds int) time.Duration {
	return time.Duration(seconds) * time.Second
}
