//go:build integration

package repository

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/gmorini/inge-soft-3/backend/internal/identity/dao"
	identityerrors "github.com/gmorini/inge-soft-3/backend/internal/identity/errors"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestRepository_CreateUser(t *testing.T) {
	pool := integrationPool(t)
	repository := New(pool)
	apartment := "2 B"

	created, err := repository.CreateUser(t.Context(), dao.CreateUserParams{
		FirstName:    "Ada",
		LastName:     "Lovelace",
		Phone:        "+5493515551234",
		Street:       "San Martín",
		StreetNumber: "123 Bis",
		Apartment:    &apartment,
		City:         "Córdoba",
		Province:     "Córdoba",
		Username:     "ada_01",
		Email:        "ada@example.com",
		PasswordHash: "$argon2id$v=19$m=19456,t=2,p=1$c2FsdHNhbHRzYWx0MTIzNA$aGFzaGhhc2hoYXNoaGFzaGhhc2hoYXNoaGFzaGhhc2g",
	})
	if err != nil {
		t.Fatalf("CreateUser() error: %v", err)
	}
	if created.ID < 1 || created.Username != "ada_01" || created.Email != "ada@example.com" {
		t.Fatalf("CreateUser() = %+v", created)
	}

	var storedHash string
	if err := pool.QueryRow(t.Context(), "SELECT password_hash FROM users WHERE id = $1", created.ID).Scan(&storedHash); err != nil {
		t.Fatalf("load password hash: %v", err)
	}
	if strings.Contains(storedHash, "Segura!@123") {
		t.Fatal("stored password contains plaintext")
	}
}

func TestRepository_FindCredentialsByUsername(t *testing.T) {
	pool := integrationPool(t)
	repository := New(pool)
	created, err := repository.CreateUser(t.Context(), dao.CreateUserParams{
		FirstName:    "Ada",
		LastName:     "Lovelace",
		Phone:        "+5493515551234",
		Street:       "San Martín",
		StreetNumber: "123",
		City:         "Córdoba",
		Province:     "Córdoba",
		Username:     "ada_01",
		Email:        "ada@example.com",
		PasswordHash: "$argon2id$v=19$m=19456,t=2,p=1$c2FsdHNhbHRzYWx0MTIzNA$aGFzaGhhc2hoYXNoaGFzaGhhc2hoYXNoaGFzaGhhc2g",
	})
	if err != nil {
		t.Fatalf("CreateUser() error: %v", err)
	}

	credentials, err := repository.FindCredentialsByUsername(t.Context(), "ada_01")
	if err != nil {
		t.Fatalf("FindCredentialsByUsername() error: %v", err)
	}
	if credentials.ID != created.ID || credentials.Username != "ada_01" {
		t.Fatalf("credentials = %+v", credentials)
	}
	if _, err := repository.FindCredentialsByUsername(t.Context(), "missing"); err == nil {
		t.Fatal("missing username returned no error")
	}
}

func TestRepository_CreateUserConflicts(t *testing.T) {
	pool := integrationPool(t)
	repository := New(pool)
	base := dao.CreateUserParams{
		FirstName: "Ada", LastName: "Lovelace", Phone: "+5493515551234", Street: "San Martín",
		StreetNumber: "123", City: "Córdoba", Province: "Córdoba", Username: "ada_01",
		Email: "ada@example.com", PasswordHash: "$argon2id$v=19$m=19456,t=2,p=1$c2FsdHNhbHRzYWx0MTIzNA$aGFzaGhhc2hoYXNoaGFzaGhhc2hoYXNoaGFzaGhhc2g",
	}
	if _, err := repository.CreateUser(t.Context(), base); err != nil {
		t.Fatalf("seed user: %v", err)
	}

	usernameConflict := base
	usernameConflict.Email = "other@example.com"
	_, err := repository.CreateUser(t.Context(), usernameConflict)
	var conflict *identityerrors.ConflictError
	if !errors.As(err, &conflict) || conflict.Field != "username" {
		t.Fatalf("username conflict = %v", err)
	}

	emailConflict := base
	emailConflict.Username = "other"
	_, err = repository.CreateUser(t.Context(), emailConflict)
	if !errors.As(err, &conflict) || conflict.Field != "email" {
		t.Fatalf("email conflict = %v", err)
	}
}

func TestRepository_CreateUserConcurrentUniqueness(t *testing.T) {
	pool := integrationPool(t)
	repository := New(pool)
	base := dao.CreateUserParams{
		FirstName: "Ada", LastName: "Lovelace", Phone: "+5493515551234", Street: "San Martín",
		StreetNumber: "123", City: "Córdoba", Province: "Córdoba", Username: "same_user",
		Email: "same@example.com", PasswordHash: "$argon2id$v=19$m=19456,t=2,p=1$c2FsdHNhbHRzYWx0MTIzNA$aGFzaGhhc2hoYXNoaGFzaGhhc2hoYXNoaGFzaGhhc2g",
	}

	start := make(chan struct{})
	results := make(chan error, 2)
	var ready sync.WaitGroup
	ready.Add(2)
	for range 2 {
		go func() {
			ready.Done()
			<-start
			_, err := repository.CreateUser(t.Context(), base)
			results <- err
		}()
	}
	ready.Wait()
	close(start)

	successes, conflicts := 0, 0
	for range 2 {
		err := <-results
		if err == nil {
			successes++
			continue
		}
		var conflict *identityerrors.ConflictError
		if errors.As(err, &conflict) {
			conflicts++
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("successes = %d, conflicts = %d", successes, conflicts)
	}
}

func integrationPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not configured")
	}
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		t.Fatalf("parse TEST_DATABASE_URL: %v", err)
	}
	if !strings.HasSuffix(cfg.ConnConfig.Database, "_test") {
		t.Fatalf("refusing database without _test suffix: %q", cfg.ConnConfig.Database)
	}
	pool, err := pgxpool.New(t.Context(), databaseURL)
	if err != nil {
		t.Fatalf("open test pool: %v", err)
	}
	t.Cleanup(pool.Close)

	lockConnection, err := pool.Acquire(t.Context())
	if err != nil {
		t.Fatalf("acquire migration connection: %v", err)
	}
	if _, err := lockConnection.Exec(t.Context(), "SELECT pg_advisory_lock(31082026)"); err != nil {
		t.Fatalf("lock migration: %v", err)
	}
	t.Cleanup(func() {
		if _, err := lockConnection.Exec(context.Background(), "SELECT pg_advisory_unlock(31082026)"); err != nil {
			t.Errorf("unlock migration: %v", err)
		}
		lockConnection.Release()
	})
	for _, name := range []string{"001_create_users.up.sql", "002_create_exercise_routines.up.sql"} {
		migrationPath := filepath.Join("..", "..", "..", "migrations", name)
		migration, err := os.ReadFile(migrationPath)
		if err != nil {
			t.Fatalf("read migration %s: %v", name, err)
		}
		if _, err := lockConnection.Exec(t.Context(), string(migration)); err != nil {
			t.Fatalf("apply migration %s: %v", name, err)
		}
	}
	if _, err := lockConnection.Exec(t.Context(), "TRUNCATE TABLE routine_sessions, session_exercises, routines, workout_sessions, exercises, users RESTART IDENTITY"); err != nil {
		t.Fatalf("clean users: %v", err)
	}
	return pool
}
