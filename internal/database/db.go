/**
 * Database — driver-agnostic database access layer.
 *
 * This package owns all database setup and exposes a small, portable API so
 * the rest of the application never touches a concrete *sql.DB or any
 * vendor-specific SQL:
 *
 *   1. Opens a connection pool for the configured driver (SQLite or Postgres)
 *   2. Runs pending migrations from embedded, driver-specific SQL files
 *   3. Provides dialect helpers (placeholder rebinding, upserts, error checks)
 *
 * The DB interface implemented by *Conn is what repositories depend on, which
 * means a repository written against database.DB works unchanged on any
 * supported driver.
 *
 * Usage:
 *   conn, err := database.Open(cfg.Database)
 *   defer conn.Close()
 *   database.Seed(conn)
 */
package database

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/lib/pq"             // registers the "postgres" driver
	_ "github.com/mattn/go-sqlite3" // registers the "sqlite3" driver
)

// Driver identifies a supported database backend.
type Driver string

const (
	// DriverSQLite is the embedded, file-backed database (default).
	DriverSQLite Driver = "sqlite"
	// DriverPostgres is an external PostgreSQL server (Supabase, Neon, RDS...).
	DriverPostgres Driver = "postgres"
)

// Config holds the connection settings for the selected driver.
type Config struct {
	// Driver selects the backend. Defaults to sqlite when empty.
	Driver Driver
	// Path is the SQLite database file path (driver=sqlite).
	Path string
	// URL is the PostgreSQL connection string (driver=postgres).
	URL string
	// SSLMode overrides the sslmode parameter of the Postgres URL.
	// Common values: require, disable, verify-full. Empty = leave as-is.
	SSLMode string
	// MaxOpenConns caps the connection pool. 0 = driver default.
	MaxOpenConns int
	// MaxIdleConns caps idle connections. 0 = driver default.
	MaxIdleConns int
	// ConnMaxLifetime recycles connections after this duration. 0 = no limit.
	ConnMaxLifetime int // seconds
}

// DB is the portable database contract used by repositories and services.
//
// It intentionally mirrors the small subset of *sql.DB the application needs
// so that swapping SQLite for Postgres requires no changes at call sites.
type DB interface {
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
	Prepare(query string) (*sql.Stmt, error)
	Begin() (*sql.Tx, error)
	Ping() error
	Close() error

	// Driver reports which backend this connection talks to.
	Driver() Driver
	// Rebind converts "?" placeholders to the driver's native form.
	Rebind(query string) string
	// IsUniqueViolation reports whether err is a UNIQUE constraint failure.
	IsUniqueViolation(err error) bool
}

// ErrUnsupportedDriver is returned when DB_DRIVER names an unknown backend.
var ErrUnsupportedDriver = errors.New("database: unsupported driver")

// Open connects to the configured database, applies migrations and returns a
// ready-to-use portable connection.
func Open(cfg Config) (DB, error) {
	driver := cfg.Driver
	if driver == "" {
		driver = DriverSQLite
	}

	switch driver {
	case DriverSQLite:
		return openSQLite(cfg)
	case DriverPostgres:
		return openPostgres(cfg)
	default:
		return nil, fmt.Errorf("%w: %q (supported: %s, %s)",
			ErrUnsupportedDriver, driver, DriverSQLite, DriverPostgres)
	}
}

// Ping verifies the connection is alive.
func Ping(db DB) error {
	if db == nil {
		return errors.New("database: nil connection")
	}
	return db.Ping()
}

// Conn is the concrete database.DB implementation. It wraps a *sql.DB and
// applies the active dialect's rules to every query (notably placeholder
// rebinding) so that callers can write portable "?"-style SQL everywhere.
type Conn struct {
	db     *sql.DB
	driver Driver
}

// NewConn wraps an existing *sql.DB in a portable Conn. Primarily useful for
// tests that need an in-memory database.
func NewConn(db *sql.DB, driver Driver) *Conn {
	return &Conn{db: db, driver: driver}
}

// SQLDB exposes the underlying *sql.DB for the rare cases that need it
// (e.g. driver-level tuning or transactions).
func (c *Conn) SQLDB() *sql.DB { return c.db }

func (c *Conn) Exec(query string, args ...any) (sql.Result, error) {
	return c.db.Exec(c.Rebind(query), args...)
}

func (c *Conn) Query(query string, args ...any) (*sql.Rows, error) {
	return c.db.Query(c.Rebind(query), args...)
}

func (c *Conn) QueryRow(query string, args ...any) *sql.Row {
	return c.db.QueryRow(c.Rebind(query), args...)
}

func (c *Conn) Prepare(query string) (*sql.Stmt, error) {
	return c.db.Prepare(c.Rebind(query))
}

func (c *Conn) Begin() (*sql.Tx, error) { return c.db.Begin() }

func (c *Conn) Ping() error { return c.db.Ping() }

func (c *Conn) Close() error { return c.db.Close() }

func (c *Conn) Driver() Driver { return c.driver }

func (c *Conn) Rebind(query string) string { return Rebind(c.driver, query) }

func (c *Conn) IsUniqueViolation(err error) bool { return isUniqueViolation(c.driver, err) }

// Rebind converts a query written with "?" placeholders into the placeholder
// syntax the driver expects. SQLite uses "?" natively; Postgres requires
// positional "$1", "$2", ... parameters.
//
// Any query that already contains "$n" placeholders is returned untouched so
// callers can safely mix styles.
func Rebind(driver Driver, query string) string {
	if driver != DriverPostgres {
		return query
	}
	if !strings.Contains(query, "?") {
		return query
	}

	var b strings.Builder
	b.Grow(len(query) + 8)

	n := 0
	inSingleQuote := false
	inDoubleQuote := false

	for i := 0; i < len(query); i++ {
		c := query[i]

		switch {
		case c == '\'' && !inDoubleQuote:
			inSingleQuote = !inSingleQuote
			b.WriteByte(c)
		case c == '"' && !inSingleQuote:
			inDoubleQuote = !inDoubleQuote
			b.WriteByte(c)
		case c == '?' && !inSingleQuote && !inDoubleQuote:
			n++
			b.WriteByte('$')
			b.WriteString(strconv.Itoa(n))
		default:
			b.WriteByte(c)
		}
	}

	return b.String()
}

// isUniqueViolation reports whether err is a UNIQUE constraint failure.
// Each driver surfaces this differently, so the check is centralised here.
func isUniqueViolation(driver Driver, err error) bool {
	if err == nil {
		return false
	}

	switch driver {
	case DriverPostgres:
		// lib/pq exposes the SQLSTATE code: 23505 = unique_violation.
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			return pqErr.Code == "23505"
		}
		// Fall back to the message for wrapped/unknown error types.
		return strings.Contains(err.Error(), "duplicate key value")
	default:
		// SQLite (mattn/go-sqlite3) reports constraint failures in the message.
		msg := err.Error()
		return strings.Contains(msg, "UNIQUE constraint failed") ||
			strings.Contains(msg, "constraint failed: rooms.slug")
	}
}
