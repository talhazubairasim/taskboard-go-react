// Package database connects to PostgreSQL and applies schema migrations.
package database

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

// Connect opens a pool and waits (up to attempts × 1s) for the database to accept connections.
func Connect(ctx context.Context, url string, attempts int) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("invalid database url: %w", err)
	}
	return ConnectConfig(ctx, cfg, attempts)
}

// ConnectConfig is Connect with a prepared pool configuration (e.g. with a query tracer).
func ConnectConfig(ctx context.Context, cfg *pgxpool.Config, attempts int) (*pgxpool.Pool, error) {
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}
	for i := 0; ; i++ {
		pctx, cancel := context.WithTimeout(ctx, 2*time.Second)
		err = pool.Ping(pctx)
		cancel()
		if err == nil {
			return pool, nil
		}
		if i+1 >= attempts {
			pool.Close()
			return nil, fmt.Errorf("database not reachable: %w", err)
		}
		time.Sleep(time.Second)
	}
}

// Migration is one numbered schema change.
type Migration struct {
	Version string
	Up      string
	Down    string
}

// Migrations returns the embedded migrations in version order.
func Migrations() ([]Migration, error) {
	entries, err := fs.ReadDir(migrationFS, "migrations")
	if err != nil {
		return nil, err
	}
	byVersion := map[string]*Migration{}
	for _, e := range entries {
		name := e.Name()
		var version, kind string
		switch {
		case strings.HasSuffix(name, ".up.sql"):
			version, kind = strings.TrimSuffix(name, ".up.sql"), "up"
		case strings.HasSuffix(name, ".down.sql"):
			version, kind = strings.TrimSuffix(name, ".down.sql"), "down"
		default:
			continue
		}
		body, err := migrationFS.ReadFile("migrations/" + name)
		if err != nil {
			return nil, err
		}
		m := byVersion[version]
		if m == nil {
			m = &Migration{Version: version}
			byVersion[version] = m
		}
		if kind == "up" {
			m.Up = string(body)
		} else {
			m.Down = string(body)
		}
	}
	out := make([]Migration, 0, len(byVersion))
	for _, m := range byVersion {
		out = append(out, *m)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Version < out[j].Version })
	return out, nil
}

const createVersionTable = `CREATE TABLE IF NOT EXISTS schema_migrations (
	version TEXT PRIMARY KEY,
	applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)`

// MigrateUp applies every migration that has not been applied yet and returns how many ran.
func MigrateUp(ctx context.Context, pool *pgxpool.Pool) (int, error) {
	if _, err := pool.Exec(ctx, createVersionTable); err != nil {
		return 0, fmt.Errorf("create schema_migrations: %w", err)
	}
	migrations, err := Migrations()
	if err != nil {
		return 0, err
	}
	applied := 0
	for _, m := range migrations {
		var exists bool
		if err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version=$1)`, m.Version).Scan(&exists); err != nil {
			return applied, err
		}
		if exists {
			continue
		}
		tx, err := pool.Begin(ctx)
		if err != nil {
			return applied, err
		}
		if _, err := tx.Exec(ctx, m.Up); err != nil {
			_ = tx.Rollback(ctx)
			return applied, fmt.Errorf("migration %s: %w", m.Version, err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations(version) VALUES ($1)`, m.Version); err != nil {
			_ = tx.Rollback(ctx)
			return applied, err
		}
		if err := tx.Commit(ctx); err != nil {
			return applied, err
		}
		applied++
	}
	return applied, nil
}

// MigrateDown rolls back the most recently applied migration. It returns the
// rolled-back version, or "" when nothing was applied.
func MigrateDown(ctx context.Context, pool *pgxpool.Pool) (string, error) {
	if _, err := pool.Exec(ctx, createVersionTable); err != nil {
		return "", err
	}
	var version string
	err := pool.QueryRow(ctx, `SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1`).Scan(&version)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	migrations, err := Migrations()
	if err != nil {
		return "", err
	}
	for _, m := range migrations {
		if m.Version != version {
			continue
		}
		tx, err := pool.Begin(ctx)
		if err != nil {
			return "", err
		}
		if _, err := tx.Exec(ctx, m.Down); err != nil {
			_ = tx.Rollback(ctx)
			return "", fmt.Errorf("rollback %s: %w", version, err)
		}
		if _, err := tx.Exec(ctx, `DELETE FROM schema_migrations WHERE version=$1`, version); err != nil {
			_ = tx.Rollback(ctx)
			return "", err
		}
		return version, tx.Commit(ctx)
	}
	return "", fmt.Errorf("no migration file for applied version %s", version)
}
