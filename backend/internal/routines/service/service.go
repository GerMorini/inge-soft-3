package service

import (
	"context"
	stderrors "errors"
	"fmt"

	routines "github.com/gmorini/inge-soft-3/backend/internal/routines"
	routineserrors "github.com/gmorini/inge-soft-3/backend/internal/routines/errors"
	"github.com/gmorini/inge-soft-3/backend/internal/routines/repository"
	"github.com/jackc/pgx/v5/pgconn"
)

type Service struct {
	repository *repository.Repository
}

func New(repository *repository.Repository) *Service {
	return &Service{repository: repository}
}

type ExerciseInput struct {
	Name        string
	Description string
	ImageURL    string
	VideoURL    string
}

type SelectedExercise struct {
	ExerciseID  int64
	Series      int32
	Repetitions int32
	Order       int32
}

type SessionInput struct {
	Name        string
	Description string
	Exercises   []SelectedExercise
}

type SelectedSession struct {
	SessionID int64
	Day       int16
}

type RoutineInput struct {
	Name        string
	Description string
	Sessions    []SelectedSession
}

type normalizedExercise struct {
	Name        string
	Description *string
	ImageURL    *string
	VideoURL    *string
}

type normalizedSession struct {
	Name        string
	Description *string
	Exercises   []SelectedExercise
}

type normalizedRoutine struct {
	Name        string
	Description *string
	Sessions    []SelectedSession
}

func (s *Service) CreateExercise(ctx context.Context, userID int64, input ExerciseInput) (routines.Exercise, error) {
	normalized, err := validateExercise(input)
	if err != nil {
		return routines.Exercise{}, err
	}
	return s.repository.CreateExercise(ctx, routines.ExerciseWrite{
		UserID: userID, Name: normalized.Name, Description: normalized.Description,
		ImageURL: normalized.ImageURL, VideoURL: normalized.VideoURL,
	})
}

func (s *Service) ListExercises(ctx context.Context, userID int64) ([]routines.Exercise, error) {
	return s.repository.ListExercises(ctx, userID)
}

func (s *Service) GetExercise(ctx context.Context, userID, exerciseID int64) (routines.Exercise, error) {
	return s.repository.GetExercise(ctx, userID, exerciseID)
}

func (s *Service) UpdateExercise(ctx context.Context, userID, exerciseID int64, input ExerciseInput) (routines.Exercise, error) {
	normalized, err := validateExercise(input)
	if err != nil {
		return routines.Exercise{}, err
	}
	return s.repository.UpdateExercise(ctx, userID, exerciseID, routines.ExerciseWrite{
		Name: normalized.Name, Description: normalized.Description,
		ImageURL: normalized.ImageURL, VideoURL: normalized.VideoURL,
	})
}

func (s *Service) CreateSession(ctx context.Context, userID int64, input SessionInput) (routines.Session, error) {
	normalized, err := validateSession(input)
	if err != nil {
		return routines.Session{}, err
	}
	tx, err := s.repository.Begin(ctx)
	if err != nil {
		return routines.Session{}, err
	}
	defer tx.Rollback(ctx)

	ids := uniqueExerciseIDs(normalized.Exercises)
	found, err := s.repository.LockExercises(ctx, tx, userID, ids)
	if err != nil {
		return routines.Session{}, err
	}
	if len(found) != len(ids) {
		return routines.Session{}, unavailableExercises(normalized.Exercises, found)
	}
	id, err := s.repository.CreateSession(ctx, tx, routines.SessionWrite{
		UserID: userID, Name: normalized.Name, Description: normalized.Description,
	})
	if err != nil {
		return routines.Session{}, err
	}
	selected := make([]routines.SelectedExercise, 0, len(normalized.Exercises))
	for _, item := range normalized.Exercises {
		selected = append(selected, routines.SelectedExercise(item))
	}
	if err := s.repository.AddSessionExercises(ctx, tx, userID, id, selected); err != nil {
		return routines.Session{}, classifyUnavailable(err, "exercises")
	}
	if err := tx.Commit(ctx); err != nil {
		return routines.Session{}, fmt.Errorf("commit session creation: %w", err)
	}
	return s.repository.GetSession(ctx, userID, id)
}

func (s *Service) ListSessions(ctx context.Context, userID int64) ([]routines.SessionSummary, error) {
	return s.repository.ListSessions(ctx, userID)
}

func (s *Service) GetSession(ctx context.Context, userID, sessionID int64) (routines.Session, error) {
	return s.repository.GetSession(ctx, userID, sessionID)
}

func (s *Service) UpdateSession(ctx context.Context, userID, sessionID int64, input SessionInput) (routines.Session, error) {
	normalized, err := validateSession(input)
	if err != nil {
		return routines.Session{}, err
	}
	tx, err := s.repository.Begin(ctx)
	if err != nil {
		return routines.Session{}, err
	}
	defer tx.Rollback(ctx)

	exists, err := s.repository.SessionExists(ctx, tx, userID, sessionID)
	if err != nil {
		return routines.Session{}, err
	}
	if !exists {
		return routines.Session{}, routineserrors.ErrNotFound
	}
	ids := uniqueExerciseIDs(normalized.Exercises)
	found, err := s.repository.LockExercises(ctx, tx, userID, ids)
	if err != nil {
		return routines.Session{}, err
	}
	if err := s.repository.LockSession(ctx, tx, userID, sessionID); err != nil {
		return routines.Session{}, err
	}
	if len(found) != len(ids) {
		return routines.Session{}, unavailableExercises(normalized.Exercises, found)
	}
	if err := s.repository.UpdateSessionFields(ctx, tx, userID, sessionID, routines.SessionWrite{
		Name: normalized.Name, Description: normalized.Description,
	}); err != nil {
		return routines.Session{}, err
	}
	if err := s.repository.ClearSessionExercises(ctx, tx, userID, sessionID); err != nil {
		return routines.Session{}, err
	}
	selected := make([]routines.SelectedExercise, 0, len(normalized.Exercises))
	for _, item := range normalized.Exercises {
		selected = append(selected, routines.SelectedExercise(item))
	}
	if err := s.repository.AddSessionExercises(ctx, tx, userID, sessionID, selected); err != nil {
		return routines.Session{}, classifyUnavailable(err, "exercises")
	}
	updated, err := s.repository.GetSessionTx(ctx, tx, userID, sessionID)
	if err != nil {
		return routines.Session{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return routines.Session{}, fmt.Errorf("commit session update: %w", err)
	}
	return updated, nil
}

func (s *Service) CreateRoutine(ctx context.Context, userID int64, input RoutineInput) (routines.Routine, error) {
	normalized, err := validateRoutine(input)
	if err != nil {
		return routines.Routine{}, err
	}
	tx, err := s.repository.Begin(ctx)
	if err != nil {
		return routines.Routine{}, err
	}
	defer tx.Rollback(ctx)

	ids := uniqueSessionIDs(normalized.Sessions)
	found, err := s.repository.LockSessions(ctx, tx, userID, ids)
	if err != nil {
		return routines.Routine{}, err
	}
	if len(found) != len(ids) {
		return routines.Routine{}, unavailableSessions(normalized.Sessions, found)
	}
	id, err := s.repository.CreateRoutine(ctx, tx, routines.RoutineWrite{
		UserID: userID, Name: normalized.Name, Description: normalized.Description,
	})
	if err != nil {
		return routines.Routine{}, err
	}
	selected := make([]routines.SelectedSession, 0, len(normalized.Sessions))
	for _, item := range normalized.Sessions {
		selected = append(selected, routines.SelectedSession(item))
	}
	if err := s.repository.AddRoutineSessions(ctx, tx, userID, id, selected); err != nil {
		return routines.Routine{}, classifyUnavailable(err, "sessions")
	}
	if err := tx.Commit(ctx); err != nil {
		return routines.Routine{}, fmt.Errorf("commit routine creation: %w", err)
	}
	return s.repository.GetRoutine(ctx, userID, id)
}

func (s *Service) ListRoutines(ctx context.Context, userID int64) ([]routines.Routine, error) {
	return s.repository.ListRoutines(ctx, userID)
}

func (s *Service) GetRoutine(ctx context.Context, userID, routineID int64) (routines.Routine, error) {
	return s.repository.GetRoutine(ctx, userID, routineID)
}

func (s *Service) UpdateRoutine(ctx context.Context, userID, routineID int64, input RoutineInput) (routines.Routine, error) {
	normalized, err := validateRoutine(input)
	if err != nil {
		return routines.Routine{}, err
	}
	tx, err := s.repository.Begin(ctx)
	if err != nil {
		return routines.Routine{}, err
	}
	defer tx.Rollback(ctx)

	exists, err := s.repository.RoutineExists(ctx, tx, userID, routineID)
	if err != nil {
		return routines.Routine{}, err
	}
	if !exists {
		return routines.Routine{}, routineserrors.ErrNotFound
	}
	ids := uniqueSessionIDs(normalized.Sessions)
	found, err := s.repository.LockSessions(ctx, tx, userID, ids)
	if err != nil {
		return routines.Routine{}, err
	}
	if err := s.repository.LockRoutine(ctx, tx, userID, routineID); err != nil {
		return routines.Routine{}, err
	}
	if len(found) != len(ids) {
		return routines.Routine{}, unavailableSessions(normalized.Sessions, found)
	}
	if err := s.repository.UpdateRoutineFields(ctx, tx, userID, routineID, routines.RoutineWrite{
		Name: normalized.Name, Description: normalized.Description,
	}); err != nil {
		return routines.Routine{}, err
	}
	if err := s.repository.ClearRoutineSessions(ctx, tx, userID, routineID); err != nil {
		return routines.Routine{}, err
	}
	selected := make([]routines.SelectedSession, 0, len(normalized.Sessions))
	for _, item := range normalized.Sessions {
		selected = append(selected, routines.SelectedSession(item))
	}
	if err := s.repository.AddRoutineSessions(ctx, tx, userID, routineID, selected); err != nil {
		return routines.Routine{}, classifyUnavailable(err, "sessions")
	}
	updated, err := s.repository.GetRoutineTx(ctx, tx, userID, routineID)
	if err != nil {
		return routines.Routine{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return routines.Routine{}, fmt.Errorf("commit routine update: %w", err)
	}
	return updated, nil
}

func (s *Service) DeleteRoutine(ctx context.Context, userID, routineID int64) error {
	return s.repository.DeleteRoutine(ctx, userID, routineID)
}

func (s *Service) DeleteSession(ctx context.Context, userID, sessionID int64) error {
	return s.repository.DeleteSession(ctx, userID, sessionID)
}

func (s *Service) DeleteExercise(ctx context.Context, userID, exerciseID int64) error {
	tx, err := s.repository.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err := s.repository.LockExercise(ctx, tx, userID, exerciseID); err != nil {
		return err
	}
	sessionIDs, err := s.repository.LockAffectedSessions(ctx, tx, userID, exerciseID)
	if err != nil {
		return err
	}
	if err := s.repository.DeleteExerciseInTx(ctx, tx, userID, exerciseID); err != nil {
		return err
	}
	if err := s.repository.CompactSessionOrders(ctx, tx, userID, sessionIDs); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit exercise deletion: %w", err)
	}
	return nil
}

func uniqueExerciseIDs(items []SelectedExercise) []int64 {
	ids := make([]int64, 0, len(items))
	seen := make(map[int64]struct{}, len(items))
	for _, item := range items {
		if _, exists := seen[item.ExerciseID]; exists {
			continue
		}
		seen[item.ExerciseID] = struct{}{}
		ids = append(ids, item.ExerciseID)
	}
	return ids
}

func uniqueSessionIDs(items []SelectedSession) []int64 {
	ids := make([]int64, 0, len(items))
	seen := make(map[int64]struct{}, len(items))
	for _, item := range items {
		if _, exists := seen[item.SessionID]; exists {
			continue
		}
		seen[item.SessionID] = struct{}{}
		ids = append(ids, item.SessionID)
	}
	return ids
}

func unavailable(field string) error {
	return &routineserrors.ValidationError{Fields: map[string][]string{field: {"Uno o más elementos no están disponibles."}}}
}

func unavailableExercises(items []SelectedExercise, found []int64) error {
	available := make(map[int64]struct{}, len(found))
	for _, id := range found {
		available[id] = struct{}{}
	}
	fields := make(map[string][]string)
	for index, item := range items {
		if _, ok := available[item.ExerciseID]; !ok {
			fields[fmt.Sprintf("exercises.%d.exerciseId", index)] = []string{"No está disponible."}
		}
	}
	return &routineserrors.ValidationError{Fields: fields}
}

func unavailableSessions(items []SelectedSession, found []int64) error {
	available := make(map[int64]struct{}, len(found))
	for _, id := range found {
		available[id] = struct{}{}
	}
	fields := make(map[string][]string)
	for index, item := range items {
		if _, ok := available[item.SessionID]; !ok {
			fields[fmt.Sprintf("sessions.%d.sessionId", index)] = []string{"No está disponible."}
		}
	}
	return &routineserrors.ValidationError{Fields: fields}
}

func classifyUnavailable(err error, field string) error {
	var postgresError *pgconn.PgError
	if stderrors.As(err, &postgresError) && postgresError.Code == "23503" {
		return unavailable(field)
	}
	return err
}
