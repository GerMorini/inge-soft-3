//go:build integration

package repository

import (
	"context"
	stderrors "errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gmorini/inge-soft-3/backend/internal/routines/dao"
	routineserrors "github.com/gmorini/inge-soft-3/backend/internal/routines/errors"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestExercisePersistenceIsOwnerScoped(t *testing.T) {
	pool := integrationPool(t)
	repository := New(pool)
	firstUser := seedUser(t, pool, "first")
	secondUser := seedUser(t, pool, "second")
	description := "Con barra"

	first, err := repository.CreateExercise(t.Context(), dao.CreateExerciseParams{UserID: firstUser, Name: "Sentadilla", Description: &description})
	if err != nil {
		t.Fatalf("create first exercise: %v", err)
	}
	second, err := repository.CreateExercise(t.Context(), dao.CreateExerciseParams{UserID: firstUser, Name: "Sentadilla"})
	if err != nil {
		t.Fatalf("create duplicate-name exercise: %v", err)
	}
	if second.Description != nil {
		t.Fatalf("optional description = %v", second.Description)
	}
	if _, err := repository.CreateExercise(t.Context(), dao.CreateExerciseParams{UserID: secondUser, Name: "Plancha"}); err != nil {
		t.Fatalf("seed foreign exercise: %v", err)
	}

	items, err := repository.ListExercises(t.Context(), firstUser)
	if err != nil {
		t.Fatalf("list exercises: %v", err)
	}
	if len(items) != 2 || items[0].ID != first.ID || items[1].ID != second.ID {
		t.Fatalf("items = %+v", items)
	}
	if _, err := repository.GetExercise(t.Context(), secondUser, first.ID); !stderrors.Is(err, routineserrors.ErrNotFound) {
		t.Fatalf("foreign lookup error = %v", err)
	}
	if _, err := repository.GetExercise(t.Context(), secondUser, 999999); !stderrors.Is(err, routineserrors.ErrNotFound) {
		t.Fatalf("missing lookup error = %v", err)
	}
}

func TestSessionRoutinePersistenceAndCascades(t *testing.T) {
	pool := integrationPool(t)
	repository := New(pool)
	userID := seedUser(t, pool, "owner")
	foreignUserID := seedUser(t, pool, "foreign")
	first, _ := repository.CreateExercise(t.Context(), dao.CreateExerciseParams{UserID: userID, Name: "Primero"})
	middle, _ := repository.CreateExercise(t.Context(), dao.CreateExerciseParams{UserID: userID, Name: "Segundo"})
	last, _ := repository.CreateExercise(t.Context(), dao.CreateExerciseParams{UserID: userID, Name: "Tercero"})
	foreign, _ := repository.CreateExercise(t.Context(), dao.CreateExerciseParams{UserID: foreignUserID, Name: "Ajeno"})

	tx, err := repository.Begin(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	found, err := repository.LockExercises(t.Context(), tx, userID, []int64{first.ID, middle.ID, last.ID})
	if err != nil || len(found) != 3 {
		t.Fatalf("lock exercises = %v, %v", found, err)
	}
	sessionID, err := repository.CreateSession(t.Context(), tx, dao.CreateSessionParams{UserID: userID, Name: "Sesión"})
	if err != nil {
		t.Fatal(err)
	}
	err = repository.AddSessionExercises(t.Context(), tx, userID, sessionID, []dao.SelectedExercise{
		{ExerciseID: first.ID, Series: 1, Repetitions: 2, Order: 1},
		{ExerciseID: middle.ID, Series: 3, Repetitions: 4, Order: 2},
		{ExerciseID: last.ID, Series: 5, Repetitions: 6, Order: 3},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(t.Context()); err != nil {
		t.Fatal(err)
	}

	detail, err := repository.GetSession(t.Context(), userID, sessionID)
	if err != nil || len(detail.Exercises) != 3 || detail.Exercises[1].Order != 2 {
		t.Fatalf("session detail = %+v, %v", detail, err)
	}

	badTx, _ := repository.Begin(t.Context())
	badSession, _ := repository.CreateSession(t.Context(), badTx, dao.CreateSessionParams{UserID: userID, Name: "Debe revertirse"})
	if err := repository.AddSessionExercises(t.Context(), badTx, userID, badSession, []dao.SelectedExercise{{ExerciseID: foreign.ID, Order: 1}}); err == nil {
		t.Fatal("cross-owner association succeeded")
	}
	_ = badTx.Rollback(t.Context())
	var partialCount int
	if err := pool.QueryRow(t.Context(), "SELECT count(*) FROM workout_sessions WHERE user_id = $1 AND id = $2", userID, badSession).Scan(&partialCount); err != nil || partialCount != 0 {
		t.Fatalf("partial session count = %d, %v", partialCount, err)
	}

	tx, _ = repository.Begin(t.Context())
	routineID, _ := repository.CreateRoutine(t.Context(), tx, dao.CreateRoutineParams{UserID: userID, Name: "Semana"})
	err = repository.AddRoutineSessions(t.Context(), tx, userID, routineID, []dao.SelectedSession{{SessionID: sessionID, Day: 1}, {SessionID: sessionID, Day: 7}})
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(t.Context()); err != nil {
		t.Fatal(err)
	}
	routine, err := repository.GetRoutine(t.Context(), userID, routineID)
	if err != nil || len(routine.Sessions) != 2 || len(routine.Sessions[0].Session.Exercises) != 3 {
		t.Fatalf("routine detail = %+v, %v", routine, err)
	}

	tx, _ = repository.Begin(t.Context())
	secondSessionID, _ := repository.CreateSession(t.Context(), tx, dao.CreateSessionParams{UserID: userID, Name: "Sesión reutilizada"})
	if err := repository.AddSessionExercises(t.Context(), tx, userID, secondSessionID, []dao.SelectedExercise{
		{ExerciseID: first.ID, Series: 9, Repetitions: 9, Order: 1},
		{ExerciseID: middle.ID, Series: 8, Repetitions: 8, Order: 2},
		{ExerciseID: last.ID, Series: 7, Repetitions: 7, Order: 3},
	}); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(t.Context()); err != nil {
		t.Fatal(err)
	}

	rollbackTx, _ := repository.Begin(t.Context())
	if err := repository.LockExercise(t.Context(), rollbackTx, userID, middle.ID); err != nil {
		t.Fatal(err)
	}
	rollbackAffected, err := repository.LockAffectedSessions(t.Context(), rollbackTx, userID, middle.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.DeleteExerciseInTx(t.Context(), rollbackTx, userID, middle.ID); err != nil {
		t.Fatal(err)
	}
	if err := repository.CompactSessionOrders(t.Context(), rollbackTx, userID, rollbackAffected); err != nil {
		t.Fatal(err)
	}
	if err := rollbackTx.Rollback(t.Context()); err != nil {
		t.Fatal(err)
	}
	detail, _ = repository.GetSession(t.Context(), userID, sessionID)
	if len(detail.Exercises) != 3 {
		t.Fatalf("rollback changed session: %+v", detail)
	}

	tx, _ = repository.Begin(t.Context())
	if err := repository.LockExercise(t.Context(), tx, userID, middle.ID); err != nil {
		t.Fatal(err)
	}
	affected, err := repository.LockAffectedSessions(t.Context(), tx, userID, middle.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.DeleteExerciseInTx(t.Context(), tx, userID, middle.ID); err != nil {
		t.Fatal(err)
	}
	if err := repository.CompactSessionOrders(t.Context(), tx, userID, affected); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(t.Context()); err != nil {
		t.Fatal(err)
	}
	detail, _ = repository.GetSession(t.Context(), userID, sessionID)
	if len(detail.Exercises) != 2 || detail.Exercises[0].Exercise.ID != first.ID || detail.Exercises[1].Exercise.ID != last.ID || detail.Exercises[1].Order != 2 {
		t.Fatalf("compacted detail = %+v", detail)
	}
	secondDetail, _ := repository.GetSession(t.Context(), userID, secondSessionID)
	if len(secondDetail.Exercises) != 2 || secondDetail.Exercises[0].Series != 9 || secondDetail.Exercises[1].Order != 2 {
		t.Fatalf("second compacted detail = %+v", secondDetail)
	}

	deleteExerciseAndCompact(t, repository, userID, first.ID)
	detail, _ = repository.GetSession(t.Context(), userID, sessionID)
	secondDetail, _ = repository.GetSession(t.Context(), userID, secondSessionID)
	if len(detail.Exercises) != 1 || detail.Exercises[0].Exercise.ID != last.ID || detail.Exercises[0].Order != 1 || len(secondDetail.Exercises) != 1 || secondDetail.Exercises[0].Order != 1 {
		t.Fatalf("first deletion details = %+v / %+v", detail, secondDetail)
	}
	deleteExerciseAndCompact(t, repository, userID, last.ID)
	detail, _ = repository.GetSession(t.Context(), userID, sessionID)
	secondDetail, _ = repository.GetSession(t.Context(), userID, secondSessionID)
	if len(detail.Exercises) != 0 || len(secondDetail.Exercises) != 0 {
		t.Fatalf("last deletion details = %+v / %+v", detail, secondDetail)
	}

	if err := repository.DeleteSession(t.Context(), userID, sessionID); err != nil {
		t.Fatal(err)
	}
	routine, _ = repository.GetRoutine(t.Context(), userID, routineID)
	if len(routine.Sessions) != 0 {
		t.Fatalf("routine assignments survived: %+v", routine.Sessions)
	}
	if _, err := repository.GetSession(t.Context(), userID, secondSessionID); err != nil {
		t.Fatalf("unrelated reusable session deleted: %v", err)
	}
	if err := repository.DeleteRoutine(t.Context(), foreignUserID, routineID); !stderrors.Is(err, routineserrors.ErrNotFound) {
		t.Fatalf("foreign delete error = %v", err)
	}
}

func deleteExerciseAndCompact(t *testing.T, repository *Repository, userID, exerciseID int64) {
	t.Helper()
	tx, err := repository.Begin(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(t.Context())
	if err := repository.LockExercise(t.Context(), tx, userID, exerciseID); err != nil {
		t.Fatal(err)
	}
	affected, err := repository.LockAffectedSessions(t.Context(), tx, userID, exerciseID)
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.DeleteExerciseInTx(t.Context(), tx, userID, exerciseID); err != nil {
		t.Fatal(err)
	}
	if err := repository.CompactSessionOrders(t.Context(), tx, userID, affected); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(t.Context()); err != nil {
		t.Fatal(err)
	}
}

func integrationPool(t *testing.T) *pgxpool.Pool {
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
		migration, err := os.ReadFile(filepath.Join("..", "..", "..", "migrations", name))
		if err != nil {
			t.Fatalf("read migration %s: %v", name, err)
		}
		if _, err := lockConnection.Exec(t.Context(), string(migration)); err != nil {
			t.Fatalf("apply migration %s: %v", name, err)
		}
	}
	if _, err := lockConnection.Exec(t.Context(), "TRUNCATE TABLE routine_sessions, session_exercises, routines, workout_sessions, exercises, users RESTART IDENTITY"); err != nil {
		t.Fatalf("clean database: %v", err)
	}
	return pool
}

func seedUser(t *testing.T, pool *pgxpool.Pool, suffix string) int64 {
	t.Helper()
	const query = `
		INSERT INTO users (first_name, last_name, phone, street, street_number, city, province, username, email, password_hash)
		VALUES ('Ada', 'Lovelace', '+5493515551234', 'San Martín', '123', 'Córdoba', 'Córdoba', $1, $2, '$argon2id$test')
		RETURNING id`
	var id int64
	if err := pool.QueryRow(t.Context(), query, "user_"+suffix, suffix+"@example.com").Scan(&id); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return id
}
