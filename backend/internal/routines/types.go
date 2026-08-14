package routines

type Exercise struct {
	ID          int64
	Name        string
	Description *string
	ImageURL    *string
	VideoURL    *string
}

type SessionSummary struct {
	ID            int64
	Name          string
	Description   *string
	ExerciseCount int64
}

type Session struct {
	ID          int64
	Name        string
	Description *string
	Exercises   []SessionExercise
}

type SessionExercise struct {
	Exercise    Exercise
	Series      int32
	Repetitions int32
	Order       int32
}

type Routine struct {
	ID          int64
	Name        string
	Description *string
	Sessions    []RoutineSession
}

type RoutineSession struct {
	Day     int16
	Session Session
}

type ExerciseWrite struct {
	UserID      int64
	Name        string
	Description *string
	ImageURL    *string
	VideoURL    *string
}

type SessionWrite struct {
	UserID      int64
	Name        string
	Description *string
}

type SelectedExercise struct {
	ExerciseID  int64
	Series      int32
	Repetitions int32
	Order       int32
}

type RoutineWrite struct {
	UserID      int64
	Name        string
	Description *string
}

type SelectedSession struct {
	SessionID int64
	Day       int16
}
