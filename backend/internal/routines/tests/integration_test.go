//go:build integration

package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"

	identitycontroller "github.com/gmorini/inge-soft-3/backend/internal/identity/controller"
	identityservice "github.com/gmorini/inge-soft-3/backend/internal/identity/service"
	routinescontroller "github.com/gmorini/inge-soft-3/backend/internal/routines/controller"
	routinesrepository "github.com/gmorini/inge-soft-3/backend/internal/routines/repository"
	routinesservice "github.com/gmorini/inge-soft-3/backend/internal/routines/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestAuthenticatedRoutineContracts(t *testing.T) {
	pool := openTestPool(t)
	applyMigrations(t, pool)
	mux, tokenManager := testMux(t, pool)
	firstUser := seedHTTPUser(t, pool, "first")
	secondUser := seedHTTPUser(t, pool, "second")
	firstToken, _ := tokenManager.Issue(firstUser, "first_user")
	secondToken, _ := tokenManager.Issue(secondUser, "second_user")

	response := perform(t, mux, http.MethodPost, "/api/exercises", firstToken, `{"name":"Sentadilla","description":"Con barra","imageUrl":"https://example.com/a.png"}`)
	if response.Code != http.StatusCreated {
		t.Fatalf("create exercise status = %d, body = %s", response.Code, response.Body.String())
	}
	firstExercise := decode[struct {
		ID int64 `json:"id"`
	}](t, response)
	response = perform(t, mux, http.MethodPost, "/api/exercises", firstToken, `{"name":"Plancha"}`)
	secondExercise := decode[struct {
		ID int64 `json:"id"`
	}](t, response)
	response = perform(t, mux, http.MethodPost, "/api/exercises", secondToken, `{"name":"Ajeno"}`)
	foreignExercise := decode[struct {
		ID int64 `json:"id"`
	}](t, response)
	response = perform(t, mux, http.MethodPost, "/api/sessions", secondToken, `{"name":"Sesión ajena","exercises":[]}`)
	foreignSession := decode[struct {
		ID int64 `json:"id"`
	}](t, response)

	response = perform(t, mux, http.MethodGet, "/api/exercises", firstToken, "")
	items := decode[[]map[string]any](t, response)
	if response.Code != http.StatusOK || len(items) != 2 {
		t.Fatalf("exercise list = %d, %#v", response.Code, items)
	}
	response = perform(t, mux, http.MethodGet, "/api/exercises/"+itoa(firstExercise.ID), secondToken, "")
	if response.Code != http.StatusNotFound {
		t.Fatalf("foreign detail status = %d", response.Code)
	}
	missing := perform(t, mux, http.MethodGet, "/api/exercises/999999", secondToken, "")
	if missing.Code != response.Code || missing.Body.String() != response.Body.String() {
		t.Fatalf("foreign and missing responses differ")
	}

	response = perform(t, mux, http.MethodPost, "/api/sessions", firstToken, `{"name":"Piernas","exercises":[{"exerciseId":`+itoa(firstExercise.ID)+`,"series":4,"repetitions":8,"order":1},{"exerciseId":`+itoa(secondExercise.ID)+`,"series":3,"repetitions":10,"order":2}]}`)
	if response.Code != http.StatusCreated {
		t.Fatalf("create session status = %d, body = %s", response.Code, response.Body.String())
	}
	session := decode[struct {
		ID        int64 `json:"id"`
		Exercises []any `json:"exercises"`
	}](t, response)
	if len(session.Exercises) != 2 {
		t.Fatalf("session exercises = %#v", session.Exercises)
	}
	sessionList := decode[[]struct {
		ID            int64 `json:"id"`
		ExerciseCount int64 `json:"exerciseCount"`
	}](t, perform(t, mux, http.MethodGet, "/api/sessions", firstToken, ""))
	if len(sessionList) != 1 || sessionList[0].ID != session.ID || sessionList[0].ExerciseCount != 2 {
		t.Fatalf("session summaries = %#v", sessionList)
	}

	response = perform(t, mux, http.MethodPost, "/api/sessions", firstToken, `{"name":"Fallida","exercises":[{"exerciseId":`+itoa(foreignExercise.ID)+`,"series":1,"repetitions":1,"order":1}]}`)
	if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), "exercises.0.exerciseId") {
		t.Fatalf("foreign selection = %d, %s", response.Code, response.Body.String())
	}
	var partial int
	if err := pool.QueryRow(t.Context(), "SELECT count(*) FROM workout_sessions WHERE user_id = $1 AND name = 'Fallida'", firstUser).Scan(&partial); err != nil || partial != 0 {
		t.Fatalf("partial session = %d, %v", partial, err)
	}

	response = perform(t, mux, http.MethodPost, "/api/routines", firstToken, `{"name":"Semana A","sessions":[{"sessionId":`+itoa(session.ID)+`,"day":1},{"sessionId":`+itoa(session.ID)+`,"day":7}]}`)
	if response.Code != http.StatusCreated {
		t.Fatalf("create routine status = %d, body = %s", response.Code, response.Body.String())
	}
	routine := decode[struct {
		ID       int64 `json:"id"`
		Sessions []struct {
			Day     int `json:"day"`
			Session struct {
				Exercises []any `json:"exercises"`
			} `json:"session"`
		} `json:"sessions"`
	}](t, response)
	if len(routine.Sessions) != 2 || routine.Sessions[0].Day != 1 || len(routine.Sessions[0].Session.Exercises) != 2 {
		t.Fatalf("routine detail = %#v", routine)
	}
	response = perform(t, mux, http.MethodPost, "/api/routines", firstToken, `{"name":"Rutina fallida","sessions":[{"sessionId":`+itoa(foreignSession.ID)+`,"day":1}]}`)
	if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), "sessions.0.sessionId") {
		t.Fatalf("foreign routine selection = %d, %s", response.Code, response.Body.String())
	}
	if err := pool.QueryRow(t.Context(), "SELECT count(*) FROM routines WHERE user_id = $1 AND name = 'Rutina fallida'", firstUser).Scan(&partial); err != nil || partial != 0 {
		t.Fatalf("partial routine = %d, %v", partial, err)
	}
	if response := perform(t, mux, http.MethodDelete, "/api/exercises/"+itoa(firstExercise.ID), secondToken, ""); response.Code != http.StatusNotFound {
		t.Fatalf("foreign exercise delete = %d", response.Code)
	}
	if response := perform(t, mux, http.MethodDelete, "/api/sessions/"+itoa(session.ID), secondToken, ""); response.Code != http.StatusNotFound {
		t.Fatalf("foreign session delete = %d", response.Code)
	}

	response = perform(t, mux, http.MethodDelete, "/api/exercises/"+itoa(firstExercise.ID), firstToken, "")
	if response.Code != http.StatusNoContent || response.Body.Len() != 0 {
		t.Fatalf("delete exercise = %d, %q", response.Code, response.Body.String())
	}
	response = perform(t, mux, http.MethodGet, "/api/sessions/"+itoa(session.ID), firstToken, "")
	compacted := decode[struct {
		Exercises []struct {
			Order int `json:"order"`
		} `json:"exercises"`
	}](t, response)
	if len(compacted.Exercises) != 1 || compacted.Exercises[0].Order != 1 {
		t.Fatalf("compacted detail = %#v", compacted)
	}

	if response := perform(t, mux, http.MethodDelete, "/api/routines/"+itoa(routine.ID), secondToken, ""); response.Code != http.StatusNotFound {
		t.Fatalf("foreign delete = %d", response.Code)
	}
	if response := perform(t, mux, http.MethodDelete, "/api/routines/"+itoa(routine.ID), firstToken, ""); response.Code != http.StatusNoContent {
		t.Fatalf("routine delete = %d", response.Code)
	}
	if response := perform(t, mux, http.MethodDelete, "/api/sessions/"+itoa(session.ID), firstToken, ""); response.Code != http.StatusNoContent {
		t.Fatalf("session delete = %d", response.Code)
	}
	for _, path := range []string{"/api/exercises/999999", "/api/sessions/999999", "/api/routines/999999"} {
		if response := perform(t, mux, http.MethodDelete, path, firstToken, ""); response.Code != http.StatusNotFound {
			t.Fatalf("missing delete %s = %d", path, response.Code)
		}
		if response := perform(t, mux, http.MethodDelete, path, "", ""); response.Code != http.StatusUnauthorized {
			t.Fatalf("unauthenticated delete %s = %d", path, response.Code)
		}
	}
}

func TestInvalidRoutineRequests(t *testing.T) {
	pool := openTestPool(t)
	applyMigrations(t, pool)
	mux, tokenManager := testMux(t, pool)
	userID := seedHTTPUser(t, pool, "invalid")
	token, _ := tokenManager.Issue(userID, "invalid_user")
	tests := []struct {
		name, path, body string
		want             int
	}{
		{name: "missing token", path: "/api/exercises", body: `{"name":"Plancha"}`, want: http.StatusUnauthorized},
		{name: "unknown field", path: "/api/exercises", body: `{"name":"Plancha","extra":true}`, want: http.StatusBadRequest},
		{name: "aggregated fields", path: "/api/exercises", body: `{"name":" Plancha ","imageUrl":"relative"}`, want: http.StatusBadRequest},
		{name: "decimal quantity", path: "/api/sessions", body: `{"name":"A","exercises":[{"exerciseId":1,"series":1.5,"repetitions":1,"order":1}]}`, want: http.StatusBadRequest},
		{name: "overflow quantity", path: "/api/sessions", body: `{"name":"A","exercises":[{"exerciseId":1,"series":2147483648,"repetitions":1,"order":1}]}`, want: http.StatusBadRequest},
		{name: "missing composition", path: "/api/sessions", body: `{"name":"A"}`, want: http.StatusBadRequest},
		{name: "duplicate exercise", path: "/api/sessions", body: `{"name":"A","exercises":[{"exerciseId":1,"series":1,"repetitions":1,"order":1},{"exerciseId":1,"series":1,"repetitions":1,"order":2}]}`, want: http.StatusBadRequest},
		{name: "gapped order", path: "/api/sessions", body: `{"name":"A","exercises":[{"exerciseId":1,"series":1,"repetitions":1,"order":2}]}`, want: http.StatusBadRequest},
		{name: "invalid path", path: "/api/exercises/0", want: http.StatusBadRequest},
		{name: "request limit", path: "/api/exercises", body: `{"name":"` + strings.Repeat("a", 17000) + `"}`, want: http.StatusBadRequest},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			requestToken := token
			if test.name == "missing token" {
				requestToken = ""
			}
			method := http.MethodPost
			if test.name == "invalid path" {
				method = http.MethodGet
			}
			response := perform(t, mux, method, test.path, requestToken, test.body)
			if response.Code != test.want {
				t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
			}
		})
	}
	response := perform(t, mux, http.MethodPost, "/api/exercises", token, `{"name":" Plancha ","imageUrl":"relative"}`)
	if !strings.Contains(response.Body.String(), `"name"`) || !strings.Contains(response.Body.String(), `"imageUrl"`) {
		t.Fatalf("aggregated errors = %s", response.Body.String())
	}
}

func TestAuthenticatedUpdateContracts(t *testing.T) {
	pool := openTestPool(t)
	applyMigrations(t, pool)
	mux, tokenManager := testMux(t, pool)
	ownerID := seedHTTPUser(t, pool, "update_owner")
	foreignID := seedHTTPUser(t, pool, "update_foreign")
	ownerToken, _ := tokenManager.Issue(ownerID, "update_owner")
	foreignToken, _ := tokenManager.Issue(foreignID, "update_foreign")

	first := decode[struct {
		ID int64 `json:"id"`
	}](t, perform(t, mux, http.MethodPost, "/api/exercises", ownerToken, `{"name":"Remo","description":"Vieja","imageUrl":"https://example.com/old"}`))
	second := decode[struct {
		ID int64 `json:"id"`
	}](t, perform(t, mux, http.MethodPost, "/api/exercises", ownerToken, `{"name":"Plancha"}`))
	session := decode[struct {
		ID int64 `json:"id"`
	}](t, perform(t, mux, http.MethodPost, "/api/sessions", ownerToken,
		`{"name":"Fuerza","exercises":[{"exerciseId":`+itoa(first.ID)+`,"series":3,"repetitions":8,"order":1}]}`))
	routine := decode[struct {
		ID int64 `json:"id"`
	}](t, perform(t, mux, http.MethodPost, "/api/routines", ownerToken,
		`{"name":"Semana","sessions":[{"sessionId":`+itoa(session.ID)+`,"day":1}]}`))

	response := perform(t, mux, http.MethodPut, "/api/exercises/"+itoa(first.ID), ownerToken, `{"name":"Remo sentado","description":"","imageUrl":"","videoUrl":""}`)
	updatedExercise := decode[map[string]any](t, response)
	if response.Code != http.StatusOK || int64(updatedExercise["id"].(float64)) != first.ID || updatedExercise["name"] != "Remo sentado" {
		t.Fatalf("exercise update = %d, %#v", response.Code, updatedExercise)
	}
	if _, exists := updatedExercise["description"]; exists {
		t.Fatalf("cleared description still present: %#v", updatedExercise)
	}

	response = perform(t, mux, http.MethodPut, "/api/sessions/"+itoa(session.ID), ownerToken,
		`{"name":"Fuerza B","description":"Nueva","exercises":[{"exerciseId":`+itoa(second.ID)+`,"series":4,"repetitions":20,"order":1}]}`)
	updatedSession := decode[struct {
		ID        int64 `json:"id"`
		Exercises []struct {
			Exercise struct {
				ID int64 `json:"id"`
			} `json:"exercise"`
		} `json:"exercises"`
	}](t, response)
	if response.Code != http.StatusOK || updatedSession.ID != session.ID || len(updatedSession.Exercises) != 1 || updatedSession.Exercises[0].Exercise.ID != second.ID {
		t.Fatalf("session update = %d, %#v", response.Code, updatedSession)
	}

	response = perform(t, mux, http.MethodGet, "/api/routines/"+itoa(routine.ID), ownerToken, "")
	propagated := decode[struct {
		Sessions []struct {
			Session struct {
				Name string `json:"name"`
			} `json:"session"`
		} `json:"sessions"`
	}](t, response)
	if len(propagated.Sessions) != 1 || propagated.Sessions[0].Session.Name != "Fuerza B" {
		t.Fatalf("updated session did not propagate: %#v", propagated)
	}

	response = perform(t, mux, http.MethodPut, "/api/routines/"+itoa(routine.ID), ownerToken, `{"name":"Semana vacía","sessions":[]}`)
	emptyRoutine := decode[struct {
		Sessions []any `json:"sessions"`
	}](t, response)
	if response.Code != http.StatusOK || len(emptyRoutine.Sessions) != 0 {
		t.Fatalf("empty routine replacement = %d, %#v", response.Code, emptyRoutine)
	}

	missingTarget := perform(t, mux, http.MethodPut, "/api/sessions/999999", ownerToken,
		`{"name":"Ausente","exercises":[{"exerciseId":999999,"series":1,"repetitions":1,"order":1}]}`)
	if missingTarget.Code != http.StatusNotFound || !strings.Contains(missingTarget.Body.String(), `"not_found"`) {
		t.Fatalf("target priority = %d, %s", missingTarget.Code, missingTarget.Body.String())
	}
	missingChild := perform(t, mux, http.MethodPut, "/api/sessions/"+itoa(session.ID), ownerToken,
		`{"name":"No debe persistir","exercises":[{"exerciseId":999999,"series":1,"repetitions":1,"order":1}]}`)
	if missingChild.Code != http.StatusBadRequest || !strings.Contains(missingChild.Body.String(), "exercises.0.exerciseId") {
		t.Fatalf("child validation = %d, %s", missingChild.Code, missingChild.Body.String())
	}
	unchanged := perform(t, mux, http.MethodGet, "/api/sessions/"+itoa(session.ID), ownerToken, "")
	if !strings.Contains(unchanged.Body.String(), `"name":"Fuerza B"`) || strings.Contains(unchanged.Body.String(), "No debe persistir") {
		t.Fatalf("failed update changed session: %s", unchanged.Body.String())
	}

	if response := perform(t, mux, http.MethodPut, "/api/exercises/"+itoa(first.ID), foreignToken, `{"name":"Ajeno"}`); response.Code != http.StatusNotFound {
		t.Fatalf("foreign update = %d, %s", response.Code, response.Body.String())
	}
	if response := perform(t, mux, http.MethodPut, "/api/exercises/"+itoa(first.ID), "", `{"name":"Sin token"}`); response.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated update = %d", response.Code)
	}
	for _, id := range []string{"abc", "0", "-1", "9223372036854775808"} {
		if response := perform(t, mux, http.MethodPut, "/api/exercises/"+id, ownerToken, `{"name":"Válido"}`); response.Code != http.StatusBadRequest {
			t.Fatalf("invalid update path %q = %d", id, response.Code)
		}
	}
	if response := perform(t, mux, http.MethodPut, "/api/exercises/999999", ownerToken, `{"name":"","extra":true}`); response.Code != http.StatusBadRequest {
		t.Fatalf("body validation priority = %d, %s", response.Code, response.Body.String())
	}

	payloads := []string{
		`{"name":"Concurrente A","exercises":[]}`,
		`{"name":"Concurrente B","exercises":[{"exerciseId":` + itoa(second.ID) + `,"series":5,"repetitions":6,"order":1}]}`,
	}
	responses := make([]*httptest.ResponseRecorder, len(payloads))
	var wait sync.WaitGroup
	for index, payload := range payloads {
		wait.Add(1)
		go func() {
			defer wait.Done()
			responses[index] = perform(t, mux, http.MethodPut, "/api/sessions/"+itoa(session.ID), ownerToken, payload)
		}()
	}
	wait.Wait()
	for index, concurrentResponse := range responses {
		if concurrentResponse.Code != http.StatusOK {
			t.Fatalf("concurrent update %d = %d, %s", index, concurrentResponse.Code, concurrentResponse.Body.String())
		}
	}
	finalResponse := perform(t, mux, http.MethodGet, "/api/sessions/"+itoa(session.ID), ownerToken, "")
	finalSession := decode[struct {
		Name      string `json:"name"`
		Exercises []any  `json:"exercises"`
	}](t, finalResponse)
	if (finalSession.Name == "Concurrente A" && len(finalSession.Exercises) != 0) ||
		(finalSession.Name == "Concurrente B" && len(finalSession.Exercises) != 1) ||
		(finalSession.Name != "Concurrente A" && finalSession.Name != "Concurrente B") {
		t.Fatalf("mixed concurrent state = %#v", finalSession)
	}
}

func testMux(t *testing.T, pool *pgxpool.Pool) (*http.ServeMux, *identityservice.TokenManager) {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	tokens, err := identityservice.NewTokenManager(strings.Repeat("s", 32))
	if err != nil {
		t.Fatal(err)
	}
	identityController := identitycontroller.New(nil, tokens, logger)
	routinesController := routinescontroller.New(routinesservice.New(routinesrepository.New(pool)), logger)
	mux := http.NewServeMux()
	routinesController.RegisterRoutes(mux, identityController.Authenticate)
	return mux, tokens
}

func perform(t *testing.T, handler http.Handler, method, path, token, body string) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func decode[T any](t *testing.T, response *httptest.ResponseRecorder) T {
	t.Helper()
	var value T
	if err := json.Unmarshal(response.Body.Bytes(), &value); err != nil {
		t.Fatalf("decode response %q: %v", response.Body.String(), err)
	}
	return value
}

func seedHTTPUser(t *testing.T, pool *pgxpool.Pool, suffix string) int64 {
	t.Helper()
	const query = `
		INSERT INTO users (first_name, last_name, phone, street, street_number, city, province, username, email, password_hash)
		VALUES ('Ada', 'Lovelace', '+5493515551234', 'San Martín', '123', 'Córdoba', 'Córdoba', $1, $2, '$argon2id$test')
		RETURNING id`
	var id int64
	if err := pool.QueryRow(t.Context(), query, suffix+"_user", suffix+"@example.com").Scan(&id); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return id
}

func itoa(value int64) string {
	return strconv.FormatInt(value, 10)
}
