package dto

type ErrorBody struct {
	Code    string              `json:"code"`
	Message string              `json:"message"`
	Fields  map[string][]string `json:"fields,omitempty"`
}

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

type CreateExerciseRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ImageURL    string `json:"imageUrl"`
	VideoURL    string `json:"videoUrl"`
}

type Exercise struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	ImageURL    *string `json:"imageUrl,omitempty"`
	VideoURL    *string `json:"videoUrl,omitempty"`
}

type SessionExerciseInput struct {
	ExerciseID  int64 `json:"exerciseId"`
	Series      int32 `json:"series"`
	Repetitions int32 `json:"repetitions"`
	Order       int32 `json:"order"`
}

type CreateSessionRequest struct {
	Name        string                  `json:"name"`
	Description string                  `json:"description"`
	Exercises   *[]SessionExerciseInput `json:"exercises"`
}

type SessionSummary struct {
	ID            int64   `json:"id"`
	Name          string  `json:"name"`
	Description   *string `json:"description,omitempty"`
	ExerciseCount int64   `json:"exerciseCount"`
}

type SessionExercise struct {
	Exercise    Exercise `json:"exercise"`
	Series      int32    `json:"series"`
	Repetitions int32    `json:"repetitions"`
	Order       int32    `json:"order"`
}

type SessionDetail struct {
	ID          int64             `json:"id"`
	Name        string            `json:"name"`
	Description *string           `json:"description,omitempty"`
	Exercises   []SessionExercise `json:"exercises"`
}

type RoutineSessionInput struct {
	SessionID int64 `json:"sessionId"`
	Day       int16 `json:"day"`
}

type CreateRoutineRequest struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Sessions    *[]RoutineSessionInput `json:"sessions"`
}

type RoutineSummary struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

type RoutineSession struct {
	Day     int16         `json:"day"`
	Session SessionDetail `json:"session"`
}

type RoutineDetail struct {
	ID          int64            `json:"id"`
	Name        string           `json:"name"`
	Description *string          `json:"description,omitempty"`
	Sessions    []RoutineSession `json:"sessions"`
}
