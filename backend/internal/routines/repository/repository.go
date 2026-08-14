package repository

import (
	"context"
	stderrors "errors"
	"fmt"

	routines "github.com/gmorini/inge-soft-3/backend/internal/routines"
	"github.com/gmorini/inge-soft-3/backend/internal/routines/dao"
	routineserrors "github.com/gmorini/inge-soft-3/backend/internal/routines/errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

const sessionDetailQuery = `
	SELECT s.id, s.name, s.description,
	       e.id, e.name, e.description, e.image_url, e.video_url,
	       se.series_count, se.repetition_count, se.execution_order
	FROM workout_sessions s
	LEFT JOIN session_exercises se
	  ON se.user_id = s.user_id AND se.session_id = s.id
	LEFT JOIN exercises e
	  ON e.user_id = se.user_id AND e.id = se.exercise_id
	WHERE s.user_id = $1 AND s.id = $2
	ORDER BY se.execution_order, e.id`

const routineDetailQuery = `
	SELECT r.id, r.name, r.description,
	       rs.day_of_week, s.id, s.name, s.description,
	       e.id, e.name, e.description, e.image_url, e.video_url,
	       se.series_count, se.repetition_count, se.execution_order
	FROM routines r
	LEFT JOIN routine_sessions rs
	  ON rs.user_id = r.user_id AND rs.routine_id = r.id
	LEFT JOIN workout_sessions s
	  ON s.user_id = rs.user_id AND s.id = rs.session_id
	LEFT JOIN session_exercises se
	  ON se.user_id = s.user_id AND se.session_id = s.id
	LEFT JOIN exercises e
	  ON e.user_id = se.user_id AND e.id = se.exercise_id
	WHERE r.user_id = $1 AND r.id = $2
	ORDER BY rs.day_of_week, s.id, se.execution_order, e.id`

func New(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Begin(ctx context.Context) (pgx.Tx, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin routines transaction: %w", err)
	}
	return tx, nil
}

func (r *Repository) CreateExercise(ctx context.Context, params routines.ExerciseWrite) (routines.Exercise, error) {
	const query = `
		INSERT INTO exercises (user_id, name, description, image_url, video_url)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, description, image_url, video_url`
	var exercise routines.Exercise
	err := r.db.QueryRow(ctx, query, params.UserID, params.Name, params.Description, params.ImageURL, params.VideoURL).
		Scan(&exercise.ID, &exercise.Name, &exercise.Description, &exercise.ImageURL, &exercise.VideoURL)
	if err != nil {
		return routines.Exercise{}, fmt.Errorf("create exercise: %w", err)
	}
	return exercise, nil
}

func (r *Repository) ListExercises(ctx context.Context, userID int64) ([]routines.Exercise, error) {
	const query = `
		SELECT id, name, description, image_url, video_url
		FROM exercises
		WHERE user_id = $1
		ORDER BY id`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list exercises: %w", err)
	}
	defer rows.Close()

	exercises := make([]routines.Exercise, 0)
	for rows.Next() {
		var exercise routines.Exercise
		if err := rows.Scan(&exercise.ID, &exercise.Name, &exercise.Description, &exercise.ImageURL, &exercise.VideoURL); err != nil {
			return nil, fmt.Errorf("scan exercise: %w", err)
		}
		exercises = append(exercises, exercise)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate exercises: %w", err)
	}
	return exercises, nil
}

func (r *Repository) GetExercise(ctx context.Context, userID, exerciseID int64) (routines.Exercise, error) {
	const query = `
		SELECT id, name, description, image_url, video_url
		FROM exercises
		WHERE user_id = $1 AND id = $2`
	var exercise routines.Exercise
	err := r.db.QueryRow(ctx, query, userID, exerciseID).
		Scan(&exercise.ID, &exercise.Name, &exercise.Description, &exercise.ImageURL, &exercise.VideoURL)
	if stderrors.Is(err, pgx.ErrNoRows) {
		return routines.Exercise{}, routineserrors.ErrNotFound
	}
	if err != nil {
		return routines.Exercise{}, fmt.Errorf("get exercise: %w", err)
	}
	return exercise, nil
}

func (r *Repository) UpdateExercise(ctx context.Context, userID, exerciseID int64, params routines.ExerciseWrite) (routines.Exercise, error) {
	const query = `
		UPDATE exercises
		SET name = $3, description = $4, image_url = $5, video_url = $6
		WHERE user_id = $1 AND id = $2
		RETURNING id, name, description, image_url, video_url`
	var exercise routines.Exercise
	err := r.db.QueryRow(ctx, query, userID, exerciseID, params.Name, params.Description, params.ImageURL, params.VideoURL).
		Scan(&exercise.ID, &exercise.Name, &exercise.Description, &exercise.ImageURL, &exercise.VideoURL)
	if stderrors.Is(err, pgx.ErrNoRows) {
		return routines.Exercise{}, routineserrors.ErrNotFound
	}
	if err != nil {
		return routines.Exercise{}, fmt.Errorf("update exercise: %w", err)
	}
	return exercise, nil
}

func (r *Repository) LockExercises(ctx context.Context, tx pgx.Tx, userID int64, exerciseIDs []int64) ([]int64, error) {
	if len(exerciseIDs) == 0 {
		return []int64{}, nil
	}
	const query = `
		SELECT id
		FROM exercises
		WHERE user_id = $1 AND id = ANY($2::bigint[])
		ORDER BY id
		FOR KEY SHARE`
	rows, err := tx.Query(ctx, query, userID, exerciseIDs)
	if err != nil {
		return nil, fmt.Errorf("lock selected exercises: %w", err)
	}
	defer rows.Close()
	found := make([]int64, 0, len(exerciseIDs))
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan selected exercise: %w", err)
		}
		found = append(found, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate selected exercises: %w", err)
	}
	return found, nil
}

func (r *Repository) CreateSession(ctx context.Context, tx pgx.Tx, params routines.SessionWrite) (int64, error) {
	const query = `
		INSERT INTO workout_sessions (user_id, name, description)
		VALUES ($1, $2, $3)
		RETURNING id`
	var id int64
	if err := tx.QueryRow(ctx, query, params.UserID, params.Name, params.Description).Scan(&id); err != nil {
		return 0, fmt.Errorf("create workout session: %w", err)
	}
	return id, nil
}

func (r *Repository) AddSessionExercises(ctx context.Context, tx pgx.Tx, userID, sessionID int64, selected []routines.SelectedExercise) error {
	const query = `
		INSERT INTO session_exercises
			(user_id, session_id, exercise_id, series_count, repetition_count, execution_order)
		VALUES ($1, $2, $3, $4, $5, $6)`
	for _, item := range selected {
		if _, err := tx.Exec(ctx, query, userID, sessionID, item.ExerciseID, item.Series, item.Repetitions, item.Order); err != nil {
			return fmt.Errorf("add exercise to session: %w", err)
		}
	}
	return nil
}

func (r *Repository) SessionExists(ctx context.Context, tx pgx.Tx, userID, sessionID int64) (bool, error) {
	const query = `SELECT EXISTS (SELECT 1 FROM workout_sessions WHERE user_id = $1 AND id = $2)`
	var exists bool
	if err := tx.QueryRow(ctx, query, userID, sessionID).Scan(&exists); err != nil {
		return false, fmt.Errorf("check workout session existence: %w", err)
	}
	return exists, nil
}

func (r *Repository) LockSession(ctx context.Context, tx pgx.Tx, userID, sessionID int64) error {
	const query = `SELECT id FROM workout_sessions WHERE user_id = $1 AND id = $2 FOR UPDATE`
	var id int64
	err := tx.QueryRow(ctx, query, userID, sessionID).Scan(&id)
	if stderrors.Is(err, pgx.ErrNoRows) {
		return routineserrors.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("lock workout session: %w", err)
	}
	return nil
}

func (r *Repository) UpdateSessionFields(ctx context.Context, tx pgx.Tx, userID, sessionID int64, params routines.SessionWrite) error {
	const query = `UPDATE workout_sessions SET name = $3, description = $4 WHERE user_id = $1 AND id = $2`
	tag, err := tx.Exec(ctx, query, userID, sessionID, params.Name, params.Description)
	if err != nil {
		return fmt.Errorf("update workout session: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return routineserrors.ErrNotFound
	}
	return nil
}

func (r *Repository) ClearSessionExercises(ctx context.Context, tx pgx.Tx, userID, sessionID int64) error {
	if _, err := tx.Exec(ctx, `DELETE FROM session_exercises WHERE user_id = $1 AND session_id = $2`, userID, sessionID); err != nil {
		return fmt.Errorf("clear workout session exercises: %w", err)
	}
	return nil
}

func (r *Repository) ListSessions(ctx context.Context, userID int64) ([]routines.SessionSummary, error) {
	const query = `
		SELECT s.id, s.name, s.description, COUNT(se.exercise_id)
		FROM workout_sessions s
		LEFT JOIN session_exercises se
		  ON se.user_id = s.user_id AND se.session_id = s.id
		WHERE s.user_id = $1
		GROUP BY s.user_id, s.id, s.name, s.description
		ORDER BY s.id`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list workout sessions: %w", err)
	}
	defer rows.Close()
	sessions := make([]routines.SessionSummary, 0)
	for rows.Next() {
		var session routines.SessionSummary
		if err := rows.Scan(&session.ID, &session.Name, &session.Description, &session.ExerciseCount); err != nil {
			return nil, fmt.Errorf("scan workout session: %w", err)
		}
		sessions = append(sessions, session)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate workout sessions: %w", err)
	}
	return sessions, nil
}

func (r *Repository) GetSession(ctx context.Context, userID, sessionID int64) (routines.Session, error) {
	rows, err := r.db.Query(ctx, sessionDetailQuery, userID, sessionID)
	if err != nil {
		return routines.Session{}, fmt.Errorf("get workout session: %w", err)
	}
	return scanSessionDetail(rows)
}

func (r *Repository) GetSessionTx(ctx context.Context, tx pgx.Tx, userID, sessionID int64) (routines.Session, error) {
	rows, err := tx.Query(ctx, sessionDetailQuery, userID, sessionID)
	if err != nil {
		return routines.Session{}, fmt.Errorf("get workout session in transaction: %w", err)
	}
	return scanSessionDetail(rows)
}

func scanSessionDetail(rows pgx.Rows) (routines.Session, error) {
	defer rows.Close()
	var session routines.Session
	seenParent := false
	for rows.Next() {
		var row dao.SessionDetailRow
		if err := rows.Scan(
			&row.SessionID, &row.SessionName, &row.SessionDescription,
			&row.ExerciseID, &row.ExerciseName, &row.ExerciseDescription,
			&row.ImageURL, &row.VideoURL, &row.Series, &row.Repetitions, &row.Order,
		); err != nil {
			return routines.Session{}, fmt.Errorf("scan workout session detail: %w", err)
		}
		if !seenParent {
			session = routines.Session{ID: row.SessionID, Name: row.SessionName, Description: row.SessionDescription, Exercises: make([]routines.SessionExercise, 0)}
			seenParent = true
		}
		if row.ExerciseID != nil {
			session.Exercises = append(session.Exercises, routines.SessionExercise{
				Exercise: routines.Exercise{ID: *row.ExerciseID, Name: *row.ExerciseName, Description: row.ExerciseDescription, ImageURL: row.ImageURL, VideoURL: row.VideoURL},
				Series:   *row.Series, Repetitions: *row.Repetitions, Order: *row.Order,
			})
		}
	}
	if err := rows.Err(); err != nil {
		return routines.Session{}, fmt.Errorf("iterate workout session detail: %w", err)
	}
	if !seenParent {
		return routines.Session{}, routineserrors.ErrNotFound
	}
	return session, nil
}

func (r *Repository) LockSessions(ctx context.Context, tx pgx.Tx, userID int64, sessionIDs []int64) ([]int64, error) {
	if len(sessionIDs) == 0 {
		return []int64{}, nil
	}
	const query = `
		SELECT id
		FROM workout_sessions
		WHERE user_id = $1 AND id = ANY($2::bigint[])
		ORDER BY id
		FOR KEY SHARE`
	rows, err := tx.Query(ctx, query, userID, sessionIDs)
	if err != nil {
		return nil, fmt.Errorf("lock selected sessions: %w", err)
	}
	defer rows.Close()
	found := make([]int64, 0, len(sessionIDs))
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan selected session: %w", err)
		}
		found = append(found, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate selected sessions: %w", err)
	}
	return found, nil
}

func (r *Repository) CreateRoutine(ctx context.Context, tx pgx.Tx, params routines.RoutineWrite) (int64, error) {
	const query = `
		INSERT INTO routines (user_id, name, description)
		VALUES ($1, $2, $3)
		RETURNING id`
	var id int64
	if err := tx.QueryRow(ctx, query, params.UserID, params.Name, params.Description).Scan(&id); err != nil {
		return 0, fmt.Errorf("create routine: %w", err)
	}
	return id, nil
}

func (r *Repository) AddRoutineSessions(ctx context.Context, tx pgx.Tx, userID, routineID int64, selected []routines.SelectedSession) error {
	const query = `
		INSERT INTO routine_sessions (user_id, routine_id, session_id, day_of_week)
		VALUES ($1, $2, $3, $4)`
	for _, item := range selected {
		if _, err := tx.Exec(ctx, query, userID, routineID, item.SessionID, item.Day); err != nil {
			return fmt.Errorf("add session to routine: %w", err)
		}
	}
	return nil
}

func (r *Repository) RoutineExists(ctx context.Context, tx pgx.Tx, userID, routineID int64) (bool, error) {
	const query = `SELECT EXISTS (SELECT 1 FROM routines WHERE user_id = $1 AND id = $2)`
	var exists bool
	if err := tx.QueryRow(ctx, query, userID, routineID).Scan(&exists); err != nil {
		return false, fmt.Errorf("check routine existence: %w", err)
	}
	return exists, nil
}

func (r *Repository) LockRoutine(ctx context.Context, tx pgx.Tx, userID, routineID int64) error {
	const query = `SELECT id FROM routines WHERE user_id = $1 AND id = $2 FOR UPDATE`
	var id int64
	err := tx.QueryRow(ctx, query, userID, routineID).Scan(&id)
	if stderrors.Is(err, pgx.ErrNoRows) {
		return routineserrors.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("lock routine: %w", err)
	}
	return nil
}

func (r *Repository) UpdateRoutineFields(ctx context.Context, tx pgx.Tx, userID, routineID int64, params routines.RoutineWrite) error {
	const query = `UPDATE routines SET name = $3, description = $4 WHERE user_id = $1 AND id = $2`
	tag, err := tx.Exec(ctx, query, userID, routineID, params.Name, params.Description)
	if err != nil {
		return fmt.Errorf("update routine: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return routineserrors.ErrNotFound
	}
	return nil
}

func (r *Repository) ClearRoutineSessions(ctx context.Context, tx pgx.Tx, userID, routineID int64) error {
	if _, err := tx.Exec(ctx, `DELETE FROM routine_sessions WHERE user_id = $1 AND routine_id = $2`, userID, routineID); err != nil {
		return fmt.Errorf("clear routine sessions: %w", err)
	}
	return nil
}

func (r *Repository) ListRoutines(ctx context.Context, userID int64) ([]routines.Routine, error) {
	const query = `
		SELECT id, name, description
		FROM routines
		WHERE user_id = $1
		ORDER BY id`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list routines: %w", err)
	}
	defer rows.Close()
	items := make([]routines.Routine, 0)
	for rows.Next() {
		var routine routines.Routine
		if err := rows.Scan(&routine.ID, &routine.Name, &routine.Description); err != nil {
			return nil, fmt.Errorf("scan routine: %w", err)
		}
		items = append(items, routine)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate routines: %w", err)
	}
	return items, nil
}

func (r *Repository) GetRoutine(ctx context.Context, userID, routineID int64) (routines.Routine, error) {
	rows, err := r.db.Query(ctx, routineDetailQuery, userID, routineID)
	if err != nil {
		return routines.Routine{}, fmt.Errorf("get routine: %w", err)
	}
	return scanRoutineDetail(rows)
}

func (r *Repository) GetRoutineTx(ctx context.Context, tx pgx.Tx, userID, routineID int64) (routines.Routine, error) {
	rows, err := tx.Query(ctx, routineDetailQuery, userID, routineID)
	if err != nil {
		return routines.Routine{}, fmt.Errorf("get routine in transaction: %w", err)
	}
	return scanRoutineDetail(rows)
}

func scanRoutineDetail(rows pgx.Rows) (routines.Routine, error) {
	defer rows.Close()
	var routine routines.Routine
	seenParent := false
	var current *routines.RoutineSession
	for rows.Next() {
		var row dao.RoutineDetailRow
		if err := rows.Scan(
			&row.RoutineID, &row.RoutineName, &row.RoutineDescription,
			&row.Day, &row.SessionID, &row.SessionName, &row.SessionDescription,
			&row.ExerciseID, &row.ExerciseName, &row.ExerciseDescription,
			&row.ImageURL, &row.VideoURL, &row.Series, &row.Repetitions, &row.Order,
		); err != nil {
			return routines.Routine{}, fmt.Errorf("scan routine detail: %w", err)
		}
		if !seenParent {
			routine = routines.Routine{ID: row.RoutineID, Name: row.RoutineName, Description: row.RoutineDescription, Sessions: make([]routines.RoutineSession, 0)}
			seenParent = true
		}
		if row.SessionID == nil {
			continue
		}
		if current == nil || current.Day != *row.Day || current.Session.ID != *row.SessionID {
			routine.Sessions = append(routine.Sessions, routines.RoutineSession{
				Day:     *row.Day,
				Session: routines.Session{ID: *row.SessionID, Name: *row.SessionName, Description: row.SessionDescription, Exercises: make([]routines.SessionExercise, 0)},
			})
			current = &routine.Sessions[len(routine.Sessions)-1]
		}
		if row.ExerciseID != nil {
			current.Session.Exercises = append(current.Session.Exercises, routines.SessionExercise{
				Exercise: routines.Exercise{ID: *row.ExerciseID, Name: *row.ExerciseName, Description: row.ExerciseDescription, ImageURL: row.ImageURL, VideoURL: row.VideoURL},
				Series:   *row.Series, Repetitions: *row.Repetitions, Order: *row.Order,
			})
		}
	}
	if err := rows.Err(); err != nil {
		return routines.Routine{}, fmt.Errorf("iterate routine detail: %w", err)
	}
	if !seenParent {
		return routines.Routine{}, routineserrors.ErrNotFound
	}
	return routine, nil
}

func (r *Repository) DeleteRoutine(ctx context.Context, userID, routineID int64) error {
	tag, err := r.db.Exec(ctx, "DELETE FROM routines WHERE user_id = $1 AND id = $2", userID, routineID)
	if err != nil {
		return fmt.Errorf("delete routine: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return routineserrors.ErrNotFound
	}
	return nil
}

func (r *Repository) DeleteSession(ctx context.Context, userID, sessionID int64) error {
	tag, err := r.db.Exec(ctx, "DELETE FROM workout_sessions WHERE user_id = $1 AND id = $2", userID, sessionID)
	if err != nil {
		return fmt.Errorf("delete workout session: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return routineserrors.ErrNotFound
	}
	return nil
}

func (r *Repository) LockExercise(ctx context.Context, tx pgx.Tx, userID, exerciseID int64) error {
	const query = `SELECT id FROM exercises WHERE user_id = $1 AND id = $2 FOR UPDATE`
	var id int64
	err := tx.QueryRow(ctx, query, userID, exerciseID).Scan(&id)
	if stderrors.Is(err, pgx.ErrNoRows) {
		return routineserrors.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("lock exercise for deletion: %w", err)
	}
	return nil
}

func (r *Repository) LockAffectedSessions(ctx context.Context, tx pgx.Tx, userID, exerciseID int64) ([]int64, error) {
	const query = `
		SELECT s.id
		FROM workout_sessions s
		JOIN session_exercises se ON se.user_id = s.user_id AND se.session_id = s.id
		WHERE se.user_id = $1 AND se.exercise_id = $2
		ORDER BY s.id
		FOR UPDATE OF s`
	rows, err := tx.Query(ctx, query, userID, exerciseID)
	if err != nil {
		return nil, fmt.Errorf("lock affected sessions: %w", err)
	}
	defer rows.Close()
	ids := make([]int64, 0)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan affected session: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate affected sessions: %w", err)
	}
	return ids, nil
}

func (r *Repository) DeleteExerciseInTx(ctx context.Context, tx pgx.Tx, userID, exerciseID int64) error {
	const query = `DELETE FROM exercises WHERE user_id = $1 AND id = $2`
	tag, err := tx.Exec(ctx, query, userID, exerciseID)
	if err != nil {
		return fmt.Errorf("delete exercise: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return routineserrors.ErrNotFound
	}
	return nil
}

func (r *Repository) CompactSessionOrders(ctx context.Context, tx pgx.Tx, userID int64, sessionIDs []int64) error {
	if len(sessionIDs) == 0 {
		return nil
	}
	if _, err := tx.Exec(ctx, "SET CONSTRAINTS session_exercises_order_unique DEFERRED"); err != nil {
		return fmt.Errorf("defer session order constraint: %w", err)
	}
	const query = `
		WITH ranked AS (
			SELECT user_id, session_id, exercise_id,
			       row_number() OVER (
				   PARTITION BY user_id, session_id
				   ORDER BY execution_order, exercise_id
			   )::integer AS new_order
			FROM session_exercises
			WHERE user_id = $1 AND session_id = ANY($2::bigint[])
		)
		UPDATE session_exercises se
		SET execution_order = ranked.new_order
		FROM ranked
		WHERE se.user_id = ranked.user_id
		  AND se.session_id = ranked.session_id
		  AND se.exercise_id = ranked.exercise_id`
	if _, err := tx.Exec(ctx, query, userID, sessionIDs); err != nil {
		return fmt.Errorf("compact session exercise order: %w", err)
	}
	return nil
}
