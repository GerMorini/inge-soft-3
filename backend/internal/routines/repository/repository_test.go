//go:build integration

package repository

import (
	"context"
	stderrors "errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	routines "github.com/gmorini/inge-soft-3/backend/internal/routines"
	routineserrors "github.com/gmorini/inge-soft-3/backend/internal/routines/errors"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestExercisePersistenceIsOwnerScoped(t *testing.T) {
	pool := integrationPool(t)
	repository := New(pool)
	firstUser := seedUser(t, pool, "first")
	secondUser := seedUser(t, pool, "second")
	description := "Con barra"

	first, err := repository.CreateExercise(t.Context(), routines.ExerciseWrite{UserID: firstUser, Name: "Sentadilla", Description: &description})
	if err != nil {
		t.Fatalf("create first exercise: %v", err)
	}
	second, err := repository.CreateExercise(t.Context(), routines.ExerciseWrite{UserID: firstUser, Name: "Sentadilla"})
	if err != nil {
		t.Fatalf("create duplicate-name exercise: %v", err)
	}
	if second.Description != nil {
		t.Fatalf("optional description = %v", second.Description)
	}
	if _, err := repository.CreateExercise(t.Context(), routines.ExerciseWrite{UserID: secondUser, Name: "Plancha"}); err != nil {
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
	first, _ := repository.CreateExercise(t.Context(), routines.ExerciseWrite{UserID: userID, Name: "Primero"})
	middle, _ := repository.CreateExercise(t.Context(), routines.ExerciseWrite{UserID: userID, Name: "Segundo"})
	last, _ := repository.CreateExercise(t.Context(), routines.ExerciseWrite{UserID: userID, Name: "Tercero"})
	foreign, _ := repository.CreateExercise(t.Context(), routines.ExerciseWrite{UserID: foreignUserID, Name: "Ajeno"})

	tx, err := repository.Begin(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	found, err := repository.LockExercises(t.Context(), tx, userID, []int64{first.ID, middle.ID, last.ID})
	if err != nil || len(found) != 3 {
		t.Fatalf("lock exercises = %v, %v", found, err)
	}
	sessionID, err := repository.CreateSession(t.Context(), tx, routines.SessionWrite{UserID: userID, Name: "Sesión"})
	if err != nil {
		t.Fatal(err)
	}
	err = repository.AddSessionExercises(t.Context(), tx, userID, sessionID, []routines.SelectedExercise{
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
	summaries, err := repository.ListSessions(t.Context(), userID)
	if err != nil || len(summaries) != 1 || summaries[0].ExerciseCount != 3 {
		t.Fatalf("session summaries = %+v, %v", summaries, err)
	}
	foreignSummaries, err := repository.ListSessions(t.Context(), foreignUserID)
	if err != nil || len(foreignSummaries) != 0 {
		t.Fatalf("foreign session summaries = %+v, %v", foreignSummaries, err)
	}

	badTx, _ := repository.Begin(t.Context())
	badSession, _ := repository.CreateSession(t.Context(), badTx, routines.SessionWrite{UserID: userID, Name: "Debe revertirse"})
	if err := repository.AddSessionExercises(t.Context(), badTx, userID, badSession, []routines.SelectedExercise{{ExerciseID: foreign.ID, Order: 1}}); err == nil {
		t.Fatal("cross-owner association succeeded")
	}
	_ = badTx.Rollback(t.Context())
	var partialCount int
	if err := pool.QueryRow(t.Context(), "SELECT count(*) FROM workout_sessions WHERE user_id = $1 AND id = $2", userID, badSession).Scan(&partialCount); err != nil || partialCount != 0 {
		t.Fatalf("partial session count = %d, %v", partialCount, err)
	}

	tx, _ = repository.Begin(t.Context())
	routineID, _ := repository.CreateRoutine(t.Context(), tx, routines.RoutineWrite{UserID: userID, Name: "Semana"})
	err = repository.AddRoutineSessions(t.Context(), tx, userID, routineID, []routines.SelectedSession{{SessionID: sessionID, Day: 1}, {SessionID: sessionID, Day: 7}})
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
	secondSessionID, _ := repository.CreateSession(t.Context(), tx, routines.SessionWrite{UserID: userID, Name: "Sesión reutilizada"})
	if err := repository.AddSessionExercises(t.Context(), tx, userID, secondSessionID, []routines.SelectedExercise{
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

func TestUpdatePrimitivesAndFlatDetails(t *testing.T) {
	pool := integrationPool(t)
	repository := New(pool)
	userID := seedUser(t, pool, "updates")
	first, err := repository.CreateExercise(t.Context(), routines.ExerciseWrite{UserID: userID, Name: "Remo"})
	if err != nil {
		t.Fatal(err)
	}
	second, _ := repository.CreateExercise(t.Context(), routines.ExerciseWrite{UserID: userID, Name: "Plancha"})
	description := "Actualizado"
	updatedExercise, err := repository.UpdateExercise(t.Context(), userID, first.ID, routines.ExerciseWrite{Name: "Remo sentado", Description: &description})
	if err != nil || updatedExercise.ID != first.ID || updatedExercise.Name != "Remo sentado" {
		t.Fatalf("updated exercise = %+v, %v", updatedExercise, err)
	}

	tx, _ := repository.Begin(t.Context())
	sessionID, _ := repository.CreateSession(t.Context(), tx, routines.SessionWrite{UserID: userID, Name: "Vacía"})
	if err := tx.Commit(t.Context()); err != nil {
		t.Fatal(err)
	}
	emptySession, err := repository.GetSession(t.Context(), userID, sessionID)
	if err != nil || len(emptySession.Exercises) != 0 {
		t.Fatalf("empty session = %+v, %v", emptySession, err)
	}

	tx, _ = repository.Begin(t.Context())
	exists, err := repository.SessionExists(t.Context(), tx, userID, sessionID)
	if err != nil || !exists {
		t.Fatalf("session exists = %v, %v", exists, err)
	}
	found, err := repository.LockExercises(t.Context(), tx, userID, []int64{first.ID, second.ID})
	if err != nil || len(found) != 2 {
		t.Fatalf("locked exercises = %v, %v", found, err)
	}
	if err := repository.LockSession(t.Context(), tx, userID, sessionID); err != nil {
		t.Fatal(err)
	}
	if err := repository.UpdateSessionFields(t.Context(), tx, userID, sessionID, routines.SessionWrite{Name: "Fuerza", Description: &description}); err != nil {
		t.Fatal(err)
	}
	if err := repository.ClearSessionExercises(t.Context(), tx, userID, sessionID); err != nil {
		t.Fatal(err)
	}
	if err := repository.AddSessionExercises(t.Context(), tx, userID, sessionID, []routines.SelectedExercise{
		{ExerciseID: second.ID, Series: 4, Repetitions: 20, Order: 1},
		{ExerciseID: first.ID, Series: 3, Repetitions: 8, Order: 2},
	}); err != nil {
		t.Fatal(err)
	}
	updatedSession, err := repository.GetSessionTx(t.Context(), tx, userID, sessionID)
	if err != nil || len(updatedSession.Exercises) != 2 || updatedSession.Exercises[0].Exercise.ID != second.ID {
		t.Fatalf("updated session = %+v, %v", updatedSession, err)
	}
	if err := tx.Commit(t.Context()); err != nil {
		t.Fatal(err)
	}

	tx, _ = repository.Begin(t.Context())
	routineID, _ := repository.CreateRoutine(t.Context(), tx, routines.RoutineWrite{UserID: userID, Name: "Vacía"})
	if err := tx.Commit(t.Context()); err != nil {
		t.Fatal(err)
	}
	emptyRoutine, err := repository.GetRoutine(t.Context(), userID, routineID)
	if err != nil || len(emptyRoutine.Sessions) != 0 {
		t.Fatalf("empty routine = %+v, %v", emptyRoutine, err)
	}

	tx, _ = repository.Begin(t.Context())
	exists, err = repository.RoutineExists(t.Context(), tx, userID, routineID)
	if err != nil || !exists {
		t.Fatalf("routine exists = %v, %v", exists, err)
	}
	found, err = repository.LockSessions(t.Context(), tx, userID, []int64{sessionID})
	if err != nil || len(found) != 1 {
		t.Fatalf("locked sessions = %v, %v", found, err)
	}
	if err := repository.LockRoutine(t.Context(), tx, userID, routineID); err != nil {
		t.Fatal(err)
	}
	if err := repository.UpdateRoutineFields(t.Context(), tx, userID, routineID, routines.RoutineWrite{Name: "Semana"}); err != nil {
		t.Fatal(err)
	}
	if err := repository.ClearRoutineSessions(t.Context(), tx, userID, routineID); err != nil {
		t.Fatal(err)
	}
	if err := repository.AddRoutineSessions(t.Context(), tx, userID, routineID, []routines.SelectedSession{{SessionID: sessionID, Day: 2}}); err != nil {
		t.Fatal(err)
	}
	updatedRoutine, err := repository.GetRoutineTx(t.Context(), tx, userID, routineID)
	if err != nil || len(updatedRoutine.Sessions) != 1 || updatedRoutine.Sessions[0].Session.Name != "Fuerza" {
		t.Fatalf("updated routine = %+v, %v", updatedRoutine, err)
	}
	if err := tx.Commit(t.Context()); err != nil {
		t.Fatal(err)
	}

	propagated, err := repository.GetRoutine(t.Context(), userID, routineID)
	if err != nil || propagated.Sessions[0].Session.Exercises[1].Exercise.Name != "Remo sentado" {
		t.Fatalf("propagated routine = %+v, %v", propagated, err)
	}
}

func TestSessionDetailReadsOnlyCommittedSnapshots(t *testing.T) {
	pool := integrationPool(t)
	repository := New(pool)
	userID := seedUser(t, pool, "snapshot")
	first, _ := repository.CreateExercise(t.Context(), routines.ExerciseWrite{UserID: userID, Name: "Primero"})
	second, _ := repository.CreateExercise(t.Context(), routines.ExerciseWrite{UserID: userID, Name: "Segundo"})

	seedTx, _ := repository.Begin(t.Context())
	sessionID, _ := repository.CreateSession(t.Context(), seedTx, routines.SessionWrite{UserID: userID, Name: "Estado anterior"})
	if err := repository.AddSessionExercises(t.Context(), seedTx, userID, sessionID, []routines.SelectedExercise{{ExerciseID: first.ID, Series: 1, Repetitions: 2, Order: 1}}); err != nil {
		t.Fatal(err)
	}
	if err := seedTx.Commit(t.Context()); err != nil {
		t.Fatal(err)
	}

	writer, _ := repository.Begin(t.Context())
	if err := repository.LockSession(t.Context(), writer, userID, sessionID); err != nil {
		t.Fatal(err)
	}
	if err := repository.UpdateSessionFields(t.Context(), writer, userID, sessionID, routines.SessionWrite{Name: "Estado nuevo"}); err != nil {
		t.Fatal(err)
	}
	if err := repository.ClearSessionExercises(t.Context(), writer, userID, sessionID); err != nil {
		t.Fatal(err)
	}
	if err := repository.AddSessionExercises(t.Context(), writer, userID, sessionID, []routines.SelectedExercise{{ExerciseID: second.ID, Series: 3, Repetitions: 4, Order: 1}}); err != nil {
		t.Fatal(err)
	}

	beforeCommit, err := repository.GetSession(t.Context(), userID, sessionID)
	if err != nil || beforeCommit.Name != "Estado anterior" || len(beforeCommit.Exercises) != 1 || beforeCommit.Exercises[0].Exercise.ID != first.ID {
		t.Fatalf("detail during update = %+v, %v", beforeCommit, err)
	}
	if err := writer.Commit(t.Context()); err != nil {
		t.Fatal(err)
	}
	afterCommit, err := repository.GetSession(t.Context(), userID, sessionID)
	if err != nil || afterCommit.Name != "Estado nuevo" || len(afterCommit.Exercises) != 1 || afterCommit.Exercises[0].Exercise.ID != second.ID {
		t.Fatalf("detail after update = %+v, %v", afterCommit, err)
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
