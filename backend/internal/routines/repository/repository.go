package repository

import (
	"context"
	stderrors "errors"
	"fmt"

	"github.com/gmorini/inge-soft-3/backend/internal/routines/dao"
	routineserrors "github.com/gmorini/inge-soft-3/backend/internal/routines/errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

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

func (r *Repository) CreateExercise(ctx context.Context, params dao.CreateExerciseParams) (dao.Exercise, error) {
	const query = `
		INSERT INTO exercises (user_id, name, description, image_url, video_url)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, description, image_url, video_url`
	var exercise dao.Exercise
	err := r.db.QueryRow(ctx, query, params.UserID, params.Name, params.Description, params.ImageURL, params.VideoURL).
		Scan(&exercise.ID, &exercise.Name, &exercise.Description, &exercise.ImageURL, &exercise.VideoURL)
	if err != nil {
		return dao.Exercise{}, fmt.Errorf("create exercise: %w", err)
	}
	return exercise, nil
}

func (r *Repository) ListExercises(ctx context.Context, userID int64) ([]dao.Exercise, error) {
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

	exercises := make([]dao.Exercise, 0)
	for rows.Next() {
		var exercise dao.Exercise
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

func (r *Repository) GetExercise(ctx context.Context, userID, exerciseID int64) (dao.Exercise, error) {
	const query = `
		SELECT id, name, description, image_url, video_url
		FROM exercises
		WHERE user_id = $1 AND id = $2`
	var exercise dao.Exercise
	err := r.db.QueryRow(ctx, query, userID, exerciseID).
		Scan(&exercise.ID, &exercise.Name, &exercise.Description, &exercise.ImageURL, &exercise.VideoURL)
	if stderrors.Is(err, pgx.ErrNoRows) {
		return dao.Exercise{}, routineserrors.ErrNotFound
	}
	if err != nil {
		return dao.Exercise{}, fmt.Errorf("get exercise: %w", err)
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

func (r *Repository) CreateSession(ctx context.Context, tx pgx.Tx, params dao.CreateSessionParams) (int64, error) {
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

func (r *Repository) AddSessionExercises(ctx context.Context, tx pgx.Tx, userID, sessionID int64, selected []dao.SelectedExercise) error {
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

func (r *Repository) ListSessions(ctx context.Context, userID int64) ([]dao.Session, error) {
	const query = `
		SELECT id, name, description
		FROM workout_sessions
		WHERE user_id = $1
		ORDER BY id`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list workout sessions: %w", err)
	}
	defer rows.Close()
	sessions := make([]dao.Session, 0)
	for rows.Next() {
		var session dao.Session
		if err := rows.Scan(&session.ID, &session.Name, &session.Description); err != nil {
			return nil, fmt.Errorf("scan workout session: %w", err)
		}
		sessions = append(sessions, session)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate workout sessions: %w", err)
	}
	return sessions, nil
}

func (r *Repository) GetSession(ctx context.Context, userID, sessionID int64) (dao.Session, error) {
	const parentQuery = `
		SELECT id, name, description
		FROM workout_sessions
		WHERE user_id = $1 AND id = $2`
	var session dao.Session
	err := r.db.QueryRow(ctx, parentQuery, userID, sessionID).
		Scan(&session.ID, &session.Name, &session.Description)
	if stderrors.Is(err, pgx.ErrNoRows) {
		return dao.Session{}, routineserrors.ErrNotFound
	}
	if err != nil {
		return dao.Session{}, fmt.Errorf("get workout session: %w", err)
	}

	const detailQuery = `
		SELECT e.id, e.name, e.description, e.image_url, e.video_url,
		       se.series_count, se.repetition_count, se.execution_order
		FROM session_exercises se
		JOIN exercises e ON e.user_id = se.user_id AND e.id = se.exercise_id
		WHERE se.user_id = $1 AND se.session_id = $2
		ORDER BY se.execution_order`
	rows, err := r.db.Query(ctx, detailQuery, userID, sessionID)
	if err != nil {
		return dao.Session{}, fmt.Errorf("list session exercises: %w", err)
	}
	defer rows.Close()
	session.Exercises = make([]dao.SessionExercise, 0)
	for rows.Next() {
		var item dao.SessionExercise
		if err := rows.Scan(
			&item.Exercise.ID, &item.Exercise.Name, &item.Exercise.Description,
			&item.Exercise.ImageURL, &item.Exercise.VideoURL, &item.Series,
			&item.Repetitions, &item.Order,
		); err != nil {
			return dao.Session{}, fmt.Errorf("scan session exercise: %w", err)
		}
		session.Exercises = append(session.Exercises, item)
	}
	if err := rows.Err(); err != nil {
		return dao.Session{}, fmt.Errorf("iterate session exercises: %w", err)
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

func (r *Repository) CreateRoutine(ctx context.Context, tx pgx.Tx, params dao.CreateRoutineParams) (int64, error) {
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

func (r *Repository) AddRoutineSessions(ctx context.Context, tx pgx.Tx, userID, routineID int64, selected []dao.SelectedSession) error {
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

func (r *Repository) ListRoutines(ctx context.Context, userID int64) ([]dao.Routine, error) {
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
	routines := make([]dao.Routine, 0)
	for rows.Next() {
		var routine dao.Routine
		if err := rows.Scan(&routine.ID, &routine.Name, &routine.Description); err != nil {
			return nil, fmt.Errorf("scan routine: %w", err)
		}
		routines = append(routines, routine)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate routines: %w", err)
	}
	return routines, nil
}

func (r *Repository) GetRoutine(ctx context.Context, userID, routineID int64) (dao.Routine, error) {
	const parentQuery = `
		SELECT id, name, description
		FROM routines
		WHERE user_id = $1 AND id = $2`
	var routine dao.Routine
	err := r.db.QueryRow(ctx, parentQuery, userID, routineID).
		Scan(&routine.ID, &routine.Name, &routine.Description)
	if stderrors.Is(err, pgx.ErrNoRows) {
		return dao.Routine{}, routineserrors.ErrNotFound
	}
	if err != nil {
		return dao.Routine{}, fmt.Errorf("get routine: %w", err)
	}

	const detailQuery = `
		SELECT rs.day_of_week, s.id, s.name, s.description,
		       e.id, e.name, e.description, e.image_url, e.video_url,
		       se.series_count, se.repetition_count, se.execution_order
		FROM routine_sessions rs
		JOIN workout_sessions s ON s.user_id = rs.user_id AND s.id = rs.session_id
		LEFT JOIN session_exercises se ON se.user_id = s.user_id AND se.session_id = s.id
		LEFT JOIN exercises e ON e.user_id = se.user_id AND e.id = se.exercise_id
		WHERE rs.user_id = $1 AND rs.routine_id = $2
		ORDER BY rs.day_of_week, s.id, se.execution_order, e.id`
	rows, err := r.db.Query(ctx, detailQuery, userID, routineID)
	if err != nil {
		return dao.Routine{}, fmt.Errorf("load routine detail: %w", err)
	}
	defer rows.Close()
	routine.Sessions = make([]dao.RoutineSession, 0)
	var current *dao.RoutineSession
	for rows.Next() {
		var day int16
		var sessionID int64
		var sessionName string
		var sessionDescription *string
		var exerciseID *int64
		var exerciseName, exerciseDescription, imageURL, videoURL *string
		var series, repetitions, order *int32
		if err := rows.Scan(
			&day, &sessionID, &sessionName, &sessionDescription,
			&exerciseID, &exerciseName, &exerciseDescription, &imageURL, &videoURL,
			&series, &repetitions, &order,
		); err != nil {
			return dao.Routine{}, fmt.Errorf("scan routine detail: %w", err)
		}
		if current == nil || current.Day != day || current.Session.ID != sessionID {
			routine.Sessions = append(routine.Sessions, dao.RoutineSession{
				Day:     day,
				Session: dao.Session{ID: sessionID, Name: sessionName, Description: sessionDescription, Exercises: make([]dao.SessionExercise, 0)},
			})
			current = &routine.Sessions[len(routine.Sessions)-1]
		}
		if exerciseID != nil {
			current.Session.Exercises = append(current.Session.Exercises, dao.SessionExercise{
				Exercise: dao.Exercise{ID: *exerciseID, Name: *exerciseName, Description: exerciseDescription, ImageURL: imageURL, VideoURL: videoURL},
				Series:   *series, Repetitions: *repetitions, Order: *order,
			})
		}
	}
	if err := rows.Err(); err != nil {
		return dao.Routine{}, fmt.Errorf("iterate routine detail: %w", err)
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
