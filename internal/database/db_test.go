package database

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lib/pq"
)

// TestOpenSQLite_RunsMigrationsAndSeed is an end-to-end check of the default
// driver: a real file-backed database gets the full schema applied and the
// admin user seeded, exactly as on a normal boot.
func TestOpenSQLite_RunsMigrationsAndSeed(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")

	db, err := Open(Config{Driver: DriverSQLite, Path: path})
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	if db.Driver() != DriverSQLite {
		t.Errorf("Driver() = %v, want %v", db.Driver(), DriverSQLite)
	}

	// Every migration should be recorded.
	var applied int
	if err := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&applied); err != nil {
		t.Fatalf("count migrations: %v", err)
	}
	if applied == 0 {
		t.Error("expected migrations to be applied")
	}

	// Tables the repositories rely on must exist.
	for _, table := range []string{
		"users", "rooms", "room_users", "room_settings", "sessions",
		"messages", "polls", "poll_votes", "settings", "webhooks",
	} {
		var name string
		err := db.QueryRow(
			`SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?`, table,
		).Scan(&name)
		if err != nil {
			t.Errorf("table %q missing: %v", table, err)
		}
	}

	// Seeding is idempotent.
	if err := Seed(db); err != nil {
		t.Fatalf("Seed() error = %v", err)
	}
	if err := Seed(db); err != nil {
		t.Fatalf("second Seed() error = %v", err)
	}

	var users int
	if err := db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&users); err != nil {
		t.Fatalf("count users: %v", err)
	}
	if users != 1 {
		t.Errorf("user count = %d, want 1 (seed must be idempotent)", users)
	}

	// The database file must actually exist on disk.
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected database file at %s: %v", path, err)
	}
}

// TestOpenSQLite_ReopeningIsIdempotent verifies that restarting the app
// against an existing file does not re-run migrations or corrupt state.
func TestOpenSQLite_ReopeningIsIdempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "reopen.db")

	first, err := Open(Config{Driver: DriverSQLite, Path: path})
	if err != nil {
		t.Fatalf("first Open() error = %v", err)
	}
	if err := Seed(first); err != nil {
		t.Fatalf("Seed() error = %v", err)
	}
	first.Close()

	second, err := Open(Config{Driver: DriverSQLite, Path: path})
	if err != nil {
		t.Fatalf("second Open() error = %v", err)
	}
	defer second.Close()

	var users int
	if err := second.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&users); err != nil {
		t.Fatalf("count users: %v", err)
	}
	if users != 1 {
		t.Errorf("user count after reopen = %d, want 1", users)
	}
}

// TestConn_ExecAndQuery exercises the portable Conn surface used by repositories.
func TestConn_ExecAndQuery(t *testing.T) {
	db, err := Open(Config{Driver: DriverSQLite, Path: filepath.Join(t.TempDir(), "conn.db")})
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(
		`INSERT INTO settings (key, value) VALUES (?, ?)
		 ON CONFLICT (key) DO UPDATE SET value = excluded.value`,
		"greeting", "hello",
	); err != nil {
		t.Fatalf("upsert insert: %v", err)
	}

	// Re-running the same upsert must update, not duplicate.
	if _, err := db.Exec(
		`INSERT INTO settings (key, value) VALUES (?, ?)
		 ON CONFLICT (key) DO UPDATE SET value = excluded.value`,
		"greeting", "hello again",
	); err != nil {
		t.Fatalf("upsert update: %v", err)
	}

	var value string
	if err := db.QueryRow(`SELECT value FROM settings WHERE key = ?`, "greeting").Scan(&value); err != nil {
		t.Fatalf("select: %v", err)
	}
	if value != "hello again" {
		t.Errorf("value = %q, want %q", value, "hello again")
	}
}

func TestSQLiteDSN(t *testing.T) {
	// The pragmas are load-bearing: WAL gives concurrent readers, and
	// foreign_keys=ON makes ON DELETE CASCADE work. SQLite ignores unknown
	// DSN parameters silently, so assert the exact string we depend on.
	want := "app.db?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=ON"
	if got := sqliteDSN("app.db"); got != want {
		t.Errorf("sqliteDSN() = %q, want %q", got, want)
	}
}

// TestSQLite_PragmasApplied guards against a driver swap silently dropping
// WAL or foreign key enforcement — both fail "successfully" if missing.
func TestSQLite_PragmasApplied(t *testing.T) {
	db, err := NewTestDB()
	if err != nil {
		t.Fatalf("NewTestDB() error = %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(`CREATE TABLE p (id INTEGER PRIMARY KEY)`); err != nil {
		t.Fatalf("create p: %v", err)
	}
	if _, err := db.Exec(
		`CREATE TABLE c (id INTEGER PRIMARY KEY, pid INTEGER REFERENCES p(id) ON DELETE CASCADE)`,
	); err != nil {
		t.Fatalf("create c: %v", err)
	}

	// A child row with a missing parent must be rejected.
	if _, err := db.Exec(`INSERT INTO c (id, pid) VALUES (1, 999)`); err == nil {
		t.Error("foreign keys are NOT enforced: orphan row was accepted")
	}

	// ON DELETE CASCADE must remove children.
	if _, err := db.Exec(`INSERT INTO p (id) VALUES (1)`); err != nil {
		t.Fatalf("insert p: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO c (id, pid) VALUES (2, 1)`); err != nil {
		t.Fatalf("insert c: %v", err)
	}
	if _, err := db.Exec(`DELETE FROM p WHERE id = 1`); err != nil {
		t.Fatalf("delete p: %v", err)
	}
	var remaining int
	if err := db.QueryRow(`SELECT COUNT(*) FROM c WHERE pid = 1`).Scan(&remaining); err != nil {
		t.Fatalf("count c: %v", err)
	}
	if remaining != 0 {
		t.Errorf("ON DELETE CASCADE did not fire: %d child rows remain", remaining)
	}
}

// TestSQLite_UpsertAndLastInsertId covers the two repository behaviours that
// differ most between SQLite drivers.
func TestSQLite_UpsertAndLastInsertId(t *testing.T) {
	db, err := NewTestDB()
	if err != nil {
		t.Fatalf("NewTestDB() error = %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(`CREATE TABLE t (id INTEGER PRIMARY KEY, v TEXT)`); err != nil {
		t.Fatalf("create t: %v", err)
	}

	res, err := db.Exec(`INSERT INTO t (v) VALUES (?)`, "x")
	if err != nil {
		t.Fatalf("insert: %v", err)
	}
	if id, err := res.LastInsertId(); err != nil || id != 1 {
		t.Errorf("LastInsertId() = %d, %v; want 1, <nil>", id, err)
	}

	// The repositories rely on this exact upsert form.
	for _, val := range []string{"a", "b"} {
		if _, err := db.Exec(
			`INSERT INTO t (id, v) VALUES (?, ?) ON CONFLICT(id) DO UPDATE SET v = excluded.v`,
			1, val,
		); err != nil {
			t.Fatalf("upsert %q: %v", val, err)
		}
	}

	var count int
	var v string
	if err := db.QueryRow(`SELECT COUNT(*) FROM t WHERE id = 1`).Scan(&count); err != nil {
		t.Fatalf("count: %v", err)
	}
	if err := db.QueryRow(`SELECT v FROM t WHERE id = 1`).Scan(&v); err != nil {
		t.Fatalf("select: %v", err)
	}
	if count != 1 || v != "b" {
		t.Errorf("upsert result count=%d value=%q; want 1 and %q", count, v, "b")
	}
}

func TestRebind_SQLiteLeavesQuestionMarks(t *testing.T) {
	q := `SELECT * FROM users WHERE id = ? AND email = ?`
	if got := Rebind(DriverSQLite, q); got != q {
		t.Fatalf("sqlite rebind should be a no-op:\n got: %s\nwant: %s", got, q)
	}
}

func TestRebind_PostgresConvertsToPositional(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "single placeholder",
			in:   `SELECT id FROM users WHERE id = ?`,
			want: `SELECT id FROM users WHERE id = $1`,
		},
		{
			name: "multiple placeholders",
			in:   `SELECT * FROM users WHERE id = ? AND email = ? AND role = ?`,
			want: `SELECT * FROM users WHERE id = $1 AND email = $2 AND role = $3`,
		},
		{
			name: "limit and offset",
			in:   `SELECT id FROM users ORDER BY id LIMIT ? OFFSET ?`,
			want: `SELECT id FROM users ORDER BY id LIMIT $1 OFFSET $2`,
		},
		{
			name: "no placeholders",
			in:   `SELECT COUNT(*) FROM users`,
			want: `SELECT COUNT(*) FROM users`,
		},
		{
			// A literal '?' inside a string must not be renumbered, otherwise
			// the query would reference a parameter that doesn't exist.
			name: "question mark inside string literal is ignored",
			in:   `SELECT * FROM polls WHERE question LIKE '%?%' AND id = ?`,
			want: `SELECT * FROM polls WHERE question LIKE '%?%' AND id = $1`,
		},
		{
			name: "question mark inside quoted identifier is ignored",
			in:   `SELECT "we?ird" FROM users WHERE id = ?`,
			want: `SELECT "we?ird" FROM users WHERE id = $1`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Rebind(DriverPostgres, tt.in); got != tt.want {
				t.Errorf("Rebind() =\n  %s\nwant:\n  %s", got, tt.want)
			}
		})
	}
}

func TestIsUniqueViolation(t *testing.T) {
	t.Run("sqlite constraint message", func(t *testing.T) {
		err := errString("UNIQUE constraint failed: rooms.slug")
		if !isUniqueViolation(DriverSQLite, err) {
			t.Error("expected sqlite unique violation to be detected")
		}
	})

	t.Run("sqlite unrelated error", func(t *testing.T) {
		if isUniqueViolation(DriverSQLite, errString("no such table: rooms")) {
			t.Error("unrelated sqlite error must not be treated as unique violation")
		}
	})

	t.Run("postgres sqlstate 23505", func(t *testing.T) {
		err := &pq.Error{Code: "23505"}
		if !isUniqueViolation(DriverPostgres, err) {
			t.Error("expected postgres 23505 to be detected as unique violation")
		}
	})

	t.Run("postgres other sqlstate", func(t *testing.T) {
		err := &pq.Error{Code: "23503"} // foreign_key_violation
		if isUniqueViolation(DriverPostgres, err) {
			t.Error("FK violation must not be treated as unique violation")
		}
	})

	t.Run("nil error", func(t *testing.T) {
		if isUniqueViolation(DriverPostgres, nil) {
			t.Error("nil error must not be a unique violation")
		}
	})
}

func TestApplySSLMode(t *testing.T) {
	tests := []struct {
		name string
		dsn  string
		mode string
		want string
	}{
		{
			name: "adds sslmode when absent",
			dsn:  "postgres://u:p@host:5432/db",
			mode: "require",
			want: "postgres://u:p@host:5432/db?sslmode=require",
		},
		{
			name: "appends with ampersand when query exists",
			dsn:  "postgres://u:p@host:5432/db?connect_timeout=10",
			mode: "require",
			want: "postgres://u:p@host:5432/db?connect_timeout=10&sslmode=require",
		},
		{
			name: "does not override an explicit sslmode",
			dsn:  "postgres://u:p@host:5432/db?sslmode=disable",
			mode: "require",
			want: "postgres://u:p@host:5432/db?sslmode=disable",
		},
		{
			name: "no-op when mode is empty",
			dsn:  "postgres://u:p@host:5432/db",
			mode: "",
			want: "postgres://u:p@host:5432/db",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := applySSLMode(tt.dsn, tt.mode); got != tt.want {
				t.Errorf("applySSLMode() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestOpen_UnsupportedDriver(t *testing.T) {
	_, err := Open(Config{Driver: Driver("mysql")})
	if err == nil {
		t.Fatal("expected an error for an unsupported driver")
	}
}

func TestOpen_PostgresRequiresURL(t *testing.T) {
	_, err := Open(Config{Driver: DriverPostgres})
	if err == nil {
		t.Fatal("expected an error when DATABASE_URL is missing")
	}
}

// errString is a minimal error implementation for table-driven tests.
type errString string

func (e errString) Error() string { return string(e) }
