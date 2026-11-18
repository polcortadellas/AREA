package db

import (
	"database/sql"
	"errors"
	"os"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// testURL returns DATABASE_URL, or skips: the test wipes the public schema's tables.
func testURL(t *testing.T) string {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL not set (use a throwaway PostgreSQL)")
	}
	return url
}

func countTables(t *testing.T, url string) int {
	t.Helper()
	conn, err := sql.Open("pgx", url)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	var n int
	err = conn.QueryRowContext(t.Context(), `SELECT count(*) FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name <> 'schema_migrations'`).Scan(&n)
	if err != nil {
		t.Fatal(err)
	}
	return n
}

// M-07 "done when": up, down and up again run without errors.
func TestMigrateUpDownUp(t *testing.T) {
	url := testURL(t)

	m, err := newMigrator(url)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _, _ = m.Close() }()
	if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) { // clean start
		t.Fatal(err)
	}

	if err := Migrate(url); err != nil {
		t.Fatalf("first up: %v", err)
	}
	if n := countTables(t, url); n != 8 {
		t.Fatalf("after up: %d tables, want 8", n)
	}
	if err := Migrate(url); err != nil { // already current: must not fail
		t.Fatalf("second up (no change): %v", err)
	}
	if err := m.Down(); err != nil {
		t.Fatalf("down: %v", err)
	}
	if n := countTables(t, url); n != 0 {
		t.Fatalf("after down: %d tables, want 0", n)
	}
	if err := Migrate(url); err != nil {
		t.Fatalf("up after down: %v", err)
	}
}

func TestMigrateBadURL(t *testing.T) {
	if err := Migrate("postgres://nobody:x@127.0.0.1:1/none?sslmode=disable&connect_timeout=1"); err == nil {
		t.Fatal("want an error for an unreachable database")
	}
}
