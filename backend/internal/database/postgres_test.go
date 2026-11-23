package database

import (
	"context"
	"os"
	"testing"
)

// Set TEST_DATABASE_URL to run these tests against a throwaway database.
func testURL(t *testing.T) string {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	return url
}

func TestMigrations(t *testing.T) {
	ms, err := Migrations()
	if err != nil {
		t.Fatal(err)
	}
	if len(ms) == 0 || ms[0].Version != "001_initial_schema" || ms[0].Up == "" || ms[0].Down == "" {
		t.Fatalf("unexpected migrations: %+v", ms)
	}
}

func TestConnectAndMigrateUpDown(t *testing.T) {
	ctx := context.Background()
	pool, err := Connect(ctx, testURL(t), 3)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	// start clean
	for {
		v, err := MigrateDown(ctx, pool)
		if err != nil {
			t.Fatal(err)
		}
		if v == "" {
			break
		}
	}

	n, err := MigrateUp(ctx, pool)
	if err != nil || n == 0 {
		t.Fatalf("first up: n=%d err=%v", n, err)
	}
	n, err = MigrateUp(ctx, pool)
	if err != nil || n != 0 {
		t.Fatalf("second up should be a no-op: n=%d err=%v", n, err)
	}
	v, err := MigrateDown(ctx, pool)
	if err != nil || v != "001_initial_schema" {
		t.Fatalf("down: v=%q err=%v", v, err)
	}
	if _, err := MigrateUp(ctx, pool); err != nil {
		t.Fatal(err)
	}
}

func TestMigrateUnknownVersion(t *testing.T) {
	ctx := context.Background()
	pool, err := Connect(ctx, testURL(t), 3)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	if _, err := MigrateUp(ctx, pool); err != nil {
		t.Fatal(err)
	}
	_, _ = pool.Exec(ctx, `INSERT INTO schema_migrations(version) VALUES ('999_missing')`)
	defer pool.Exec(ctx, `DELETE FROM schema_migrations WHERE version='999_missing'`)
	if _, err := MigrateDown(ctx, pool); err == nil {
		t.Fatal("expected error for missing migration file")
	}
}

func TestMigrateFailures(t *testing.T) {
	ctx := context.Background()
	pool, err := Connect(ctx, testURL(t), 3)
	if err != nil {
		t.Fatal(err)
	}
	pool.Close() // closed pool makes every call fail
	if _, err := MigrateUp(ctx, pool); err == nil {
		t.Fatal("MigrateUp on closed pool should fail")
	}
	if _, err := MigrateDown(ctx, pool); err == nil {
		t.Fatal("MigrateDown on closed pool should fail")
	}
}

func TestConnectErrors(t *testing.T) {
	ctx := context.Background()
	if _, err := Connect(ctx, "::not a url::", 1); err == nil {
		t.Fatal("bad url accepted")
	}
	if _, err := Connect(ctx, "postgres://u:p@127.0.0.1:1/db?sslmode=disable&connect_timeout=1", 2); err == nil {
		t.Fatal("unreachable db accepted")
	}
}
