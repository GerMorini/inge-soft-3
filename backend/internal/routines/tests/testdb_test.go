//go:build integration

package tests

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func openTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not configured")
	}
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		t.Fatalf("parse TEST_DATABASE_URL: %v", err)
	}
	if !strings.HasSuffix(config.ConnConfig.Database, "_test") {
		t.Fatalf("refusing database without _test suffix: %q", config.ConnConfig.Database)
	}
	pool, err := pgxpool.New(t.Context(), databaseURL)
	if err != nil {
		t.Fatalf("open test pool: %v", err)
	}
	if err := pool.Ping(t.Context()); err != nil {
		pool.Close()
		t.Fatalf("ping test database: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func applyMigrations(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	connection, err := pool.Acquire(context.Background())
	if err != nil {
		t.Fatalf("acquire migration connection: %v", err)
	}
	if _, err := connection.Exec(context.Background(), "SELECT pg_advisory_lock(31082026)"); err != nil {
		t.Fatalf("lock migrations: %v", err)
	}
	t.Cleanup(func() {
		if _, err := connection.Exec(context.Background(), "SELECT pg_advisory_unlock(31082026)"); err != nil {
			t.Errorf("unlock migrations: %v", err)
		}
		connection.Release()
	})
	for _, name := range []string{"001_create_users.up.sql", "002_create_exercise_routines.up.sql"} {
		migration, err := os.ReadFile(filepath.Join("..", "..", "..", "migrations", name))
		if err != nil {
			t.Fatalf("read migration %s: %v", name, err)
		}
		if _, err := connection.Exec(context.Background(), string(migration)); err != nil {
			t.Fatalf("apply migration %s: %v", name, err)
		}
	}
	if _, err := connection.Exec(context.Background(), "TRUNCATE TABLE routine_sessions, session_exercises, routines, workout_sessions, exercises, users RESTART IDENTITY"); err != nil {
		t.Fatalf("clean database: %v", err)
	}
}
