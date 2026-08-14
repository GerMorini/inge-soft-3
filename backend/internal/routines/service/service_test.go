package service

import (
	stderrors "errors"
	"strings"
	"testing"

	routineserrors "github.com/gmorini/inge-soft-3/backend/internal/routines/errors"
)

func TestValidateExercise(t *testing.T) {
	tests := []struct {
		name      string
		input     ExerciseInput
		wantError bool
	}{
		{name: "minimal", input: ExerciseInput{Name: "Sentadilla"}},
		{name: "complete", input: ExerciseInput{Name: "Peso muerto", Description: "Con barra", ImageURL: "https://example.com/a.png", VideoURL: "http://example.com/a"}},
		{name: "empty optionals become absent", input: ExerciseInput{Name: "Plancha", Description: "", ImageURL: "", VideoURL: ""}},
		{name: "name boundary", input: ExerciseInput{Name: strings.Repeat("a", 100)}},
		{name: "description boundary", input: ExerciseInput{Name: "Plancha", Description: strings.Repeat("a", 500)}},
		{name: "url boundary", input: ExerciseInput{Name: "Plancha", ImageURL: "https://e.co/" + strings.Repeat("a", 2035)}},
		{name: "missing name", input: ExerciseInput{}, wantError: true},
		{name: "leading space", input: ExerciseInput{Name: " Sentadilla"}, wantError: true},
		{name: "trailing space", input: ExerciseInput{Name: "Sentadilla "}, wantError: true},
		{name: "repeated space", input: ExerciseInput{Name: "Peso  muerto"}, wantError: true},
		{name: "tab", input: ExerciseInput{Name: "Peso\tmuerto"}, wantError: true},
		{name: "long name", input: ExerciseInput{Name: strings.Repeat("a", 101)}, wantError: true},
		{name: "relative url", input: ExerciseInput{Name: "Plancha", ImageURL: "/imagen.png"}, wantError: true},
		{name: "ftp url", input: ExerciseInput{Name: "Plancha", VideoURL: "ftp://example.com/a"}, wantError: true},
		{name: "url whitespace", input: ExerciseInput{Name: "Plancha", ImageURL: "https://example.com/a b"}, wantError: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := validateExercise(test.input)
			if (err != nil) != test.wantError {
				t.Fatalf("validateExercise() error = %v", err)
			}
			if err == nil && test.name == "empty optionals become absent" && (got.Description != nil || got.ImageURL != nil || got.VideoURL != nil) {
				t.Fatalf("optionals = %+v", got)
			}
		})
	}
}

func TestValidateSession(t *testing.T) {
	tests := []struct {
		name       string
		input      SessionInput
		wantFields []string
	}{
		{name: "empty session", input: SessionInput{Name: "Descanso"}},
		{name: "zero quantities", input: SessionInput{Name: "Piernas", Exercises: []SelectedExercise{{ExerciseID: 1, Series: 0, Repetitions: 0, Order: 1}}}},
		{name: "maximum quantities", input: SessionInput{Name: "Piernas", Exercises: []SelectedExercise{{ExerciseID: 1, Series: 2147483647, Repetitions: 2147483647, Order: 1}}}},
		{name: "negative values", input: SessionInput{Name: "Piernas", Exercises: []SelectedExercise{{ExerciseID: 1, Series: -1, Repetitions: -2, Order: 1}}}, wantFields: []string{"exercises.0.series", "exercises.0.repetitions"}},
		{name: "duplicate exercise", input: SessionInput{Name: "Piernas", Exercises: []SelectedExercise{{ExerciseID: 1, Order: 1}, {ExerciseID: 1, Order: 2}}}, wantFields: []string{"exercises.1.exerciseId"}},
		{name: "gapped order", input: SessionInput{Name: "Piernas", Exercises: []SelectedExercise{{ExerciseID: 1, Order: 2}}}, wantFields: []string{"exercises.0.order"}},
		{name: "aggregated", input: SessionInput{Name: " Piernas", Exercises: []SelectedExercise{{ExerciseID: 0, Series: -1, Repetitions: -1, Order: 0}}}, wantFields: []string{"name", "exercises.0.exerciseId", "exercises.0.series", "exercises.0.repetitions", "exercises.0.order"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := validateSession(test.input)
			if len(test.wantFields) == 0 {
				if err != nil {
					t.Fatalf("validateSession() error = %v", err)
				}
				return
			}
			var validation *routineserrors.ValidationError
			if !stderrors.As(err, &validation) {
				t.Fatalf("error = %v", err)
			}
			for _, field := range test.wantFields {
				if len(validation.Fields[field]) == 0 {
					t.Errorf("missing field %q in %#v", field, validation.Fields)
				}
			}
		})
	}
}

func TestValidateRoutine(t *testing.T) {
	tests := []struct {
		name      string
		input     RoutineInput
		wantField string
	}{
		{name: "empty routine", input: RoutineInput{Name: "Semana A"}},
		{name: "weekday boundaries", input: RoutineInput{Name: "Semana A", Sessions: []SelectedSession{{SessionID: 1, Day: 1}, {SessionID: 1, Day: 7}}}},
		{name: "same day different sessions", input: RoutineInput{Name: "Semana A", Sessions: []SelectedSession{{SessionID: 1, Day: 2}, {SessionID: 2, Day: 2}}}},
		{name: "day zero", input: RoutineInput{Name: "Semana A", Sessions: []SelectedSession{{SessionID: 1, Day: 0}}}, wantField: "sessions.0.day"},
		{name: "day eight", input: RoutineInput{Name: "Semana A", Sessions: []SelectedSession{{SessionID: 1, Day: 8}}}, wantField: "sessions.0.day"},
		{name: "duplicate pair", input: RoutineInput{Name: "Semana A", Sessions: []SelectedSession{{SessionID: 1, Day: 2}, {SessionID: 1, Day: 2}}}, wantField: "sessions.1.sessionId"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := validateRoutine(test.input)
			if test.wantField == "" {
				if err != nil {
					t.Fatalf("validateRoutine() error = %v", err)
				}
				return
			}
			var validation *routineserrors.ValidationError
			if !stderrors.As(err, &validation) || len(validation.Fields[test.wantField]) == 0 {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

func TestReplacementInputsReuseValidationAndIndexUnavailableChildren(t *testing.T) {
	exercise, err := validateExercise(ExerciseInput{Name: "Remo", Description: "", ImageURL: "", VideoURL: ""})
	if err != nil || exercise.Description != nil || exercise.ImageURL != nil || exercise.VideoURL != nil {
		t.Fatalf("normalized exercise = %+v, %v", exercise, err)
	}
	session, err := validateSession(SessionInput{Name: "Sesión vacía", Exercises: []SelectedExercise{}})
	if err != nil || len(session.Exercises) != 0 {
		t.Fatalf("normalized session = %+v, %v", session, err)
	}
	routine, err := validateRoutine(RoutineInput{Name: "Rutina vacía", Sessions: []SelectedSession{}})
	if err != nil || len(routine.Sessions) != 0 {
		t.Fatalf("normalized routine = %+v, %v", routine, err)
	}

	var validation *routineserrors.ValidationError
	err = unavailableExercises([]SelectedExercise{{ExerciseID: 1}, {ExerciseID: 2}}, []int64{1})
	if !stderrors.As(err, &validation) || len(validation.Fields["exercises.1.exerciseId"]) == 0 {
		t.Fatalf("exercise availability error = %v", err)
	}
	err = unavailableSessions([]SelectedSession{{SessionID: 3, Day: 1}, {SessionID: 4, Day: 2}}, []int64{4})
	if !stderrors.As(err, &validation) || len(validation.Fields["sessions.0.sessionId"]) == 0 {
		t.Fatalf("session availability error = %v", err)
	}
}
