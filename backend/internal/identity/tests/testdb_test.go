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

func applyMigration(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	path := filepath.Join("..", "..", "..", "migrations", "001_create_users.up.sql")
	sql, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	lockConnection, err := pool.Acquire(context.Background())
	if err != nil {
		t.Fatalf("acquire migration connection: %v", err)
	}
	if _, err := lockConnection.Exec(context.Background(), "SELECT pg_advisory_lock(31082026)"); err != nil {
		t.Fatalf("lock migration: %v", err)
	}
	t.Cleanup(func() {
		if _, err := lockConnection.Exec(context.Background(), "SELECT pg_advisory_unlock(31082026)"); err != nil {
			t.Errorf("unlock migration: %v", err)
		}
		lockConnection.Release()
	})
	if _, err := lockConnection.Exec(context.Background(), string(sql)); err != nil {
		t.Fatalf("apply migration: %v", err)
	}
	if _, err := lockConnection.Exec(context.Background(), "TRUNCATE TABLE users RESTART IDENTITY"); err != nil {
		t.Fatalf("clean users: %v", err)
	}
}
