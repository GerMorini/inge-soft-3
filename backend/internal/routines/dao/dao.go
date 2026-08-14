package dao

// SessionDetailRow represents the nullable result of the session detail join.
type SessionDetailRow struct {
	SessionID           int64
	SessionName         string
	SessionDescription  *string
	ExerciseID          *int64
	ExerciseName        *string
	ExerciseDescription *string
	ImageURL            *string
	VideoURL            *string
	Series              *int32
	Repetitions         *int32
	Order               *int32
}

// RoutineDetailRow represents one nullable row of the routine detail join.
type RoutineDetailRow struct {
	RoutineID           int64
	RoutineName         string
	RoutineDescription  *string
	Day                 *int16
	SessionID           *int64
	SessionName         *string
	SessionDescription  *string
	ExerciseID          *int64
	ExerciseName        *string
	ExerciseDescription *string
	ImageURL            *string
	VideoURL            *string
	Series              *int32
	Repetitions         *int32
	Order               *int32
}
