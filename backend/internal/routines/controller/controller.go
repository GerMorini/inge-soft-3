package controller

import (
	"context"
	"encoding/json"
	stderrors "errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gmorini/inge-soft-3/backend/internal/platform/requestctx"
	routines "github.com/gmorini/inge-soft-3/backend/internal/routines"
	"github.com/gmorini/inge-soft-3/backend/internal/routines/dto"
	routineserrors "github.com/gmorini/inge-soft-3/backend/internal/routines/errors"
	"github.com/gmorini/inge-soft-3/backend/internal/routines/service"
)

const maxRequestBody = 16 << 10

type Controller struct {
	service *service.Service
	logger  *slog.Logger
}

func New(service *service.Service, logger *slog.Logger) *Controller {
	return &Controller{service: service, logger: logger}
}

func (c *Controller) RegisterRoutes(mux *http.ServeMux, authenticate func(http.Handler) http.Handler) {
	mux.Handle("POST /api/exercises", authenticate(http.HandlerFunc(c.createExercise)))
	mux.Handle("GET /api/exercises", authenticate(http.HandlerFunc(c.listExercises)))
	mux.Handle("GET /api/exercises/{id}", authenticate(http.HandlerFunc(c.getExercise)))
	mux.Handle("PUT /api/exercises/{id}", authenticate(http.HandlerFunc(c.updateExercise)))
	mux.Handle("DELETE /api/exercises/{id}", authenticate(http.HandlerFunc(c.deleteExercise)))
	mux.Handle("POST /api/sessions", authenticate(http.HandlerFunc(c.createSession)))
	mux.Handle("GET /api/sessions", authenticate(http.HandlerFunc(c.listSessions)))
	mux.Handle("GET /api/sessions/{id}", authenticate(http.HandlerFunc(c.getSession)))
	mux.Handle("PUT /api/sessions/{id}", authenticate(http.HandlerFunc(c.updateSession)))
	mux.Handle("DELETE /api/sessions/{id}", authenticate(http.HandlerFunc(c.deleteSession)))
	mux.Handle("POST /api/routines", authenticate(http.HandlerFunc(c.createRoutine)))
	mux.Handle("GET /api/routines", authenticate(http.HandlerFunc(c.listRoutines)))
	mux.Handle("GET /api/routines/{id}", authenticate(http.HandlerFunc(c.getRoutine)))
	mux.Handle("PUT /api/routines/{id}", authenticate(http.HandlerFunc(c.updateRoutine)))
	mux.Handle("DELETE /api/routines/{id}", authenticate(http.HandlerFunc(c.deleteRoutine)))
}

func (c *Controller) createExercise(w http.ResponseWriter, r *http.Request) {
	var request dto.CreateExerciseRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Solicitud inválida.", nil)
		return
	}
	identity, ok := authenticatedIdentity(w, r)
	if !ok {
		return
	}
	created, err := c.service.CreateExercise(r.Context(), identity.UserID, service.ExerciseInput{
		Name: request.Name, Description: request.Description, ImageURL: request.ImageURL, VideoURL: request.VideoURL,
	})
	if !c.handleError(w, err) {
		return
	}
	writeJSON(w, http.StatusCreated, exerciseResponse(created))
}

func (c *Controller) listExercises(w http.ResponseWriter, r *http.Request) {
	identity, ok := authenticatedIdentity(w, r)
	if !ok {
		return
	}
	exercises, err := c.service.ListExercises(r.Context(), identity.UserID)
	if !c.handleError(w, err) {
		return
	}
	response := make([]dto.Exercise, 0, len(exercises))
	for _, exercise := range exercises {
		response = append(response, exerciseResponse(exercise))
	}
	writeJSON(w, http.StatusOK, response)
}

func (c *Controller) getExercise(w http.ResponseWriter, r *http.Request) {
	identity, ok := authenticatedIdentity(w, r)
	if !ok {
		return
	}
	id, ok := resourceID(w, r)
	if !ok {
		return
	}
	exercise, err := c.service.GetExercise(r.Context(), identity.UserID, id)
	if !c.handleError(w, err) {
		return
	}
	writeJSON(w, http.StatusOK, exerciseResponse(exercise))
}

func (c *Controller) updateExercise(w http.ResponseWriter, r *http.Request) {
	identity, ok := authenticatedIdentity(w, r)
	if !ok {
		return
	}
	id, ok := resourceID(w, r)
	if !ok {
		return
	}
	var request dto.CreateExerciseRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Solicitud inválida.", nil)
		return
	}
	updated, err := c.service.UpdateExercise(r.Context(), identity.UserID, id, exerciseInput(request))
	if !c.handleError(w, err) {
		return
	}
	writeJSON(w, http.StatusOK, exerciseResponse(updated))
}

func (c *Controller) createSession(w http.ResponseWriter, r *http.Request) {
	var request dto.CreateSessionRequest
	if err := decodeJSON(w, r, &request); err != nil || request.Exercises == nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Solicitud inválida.", nil)
		return
	}
	identity, ok := authenticatedIdentity(w, r)
	if !ok {
		return
	}
	selected := make([]service.SelectedExercise, 0, len(*request.Exercises))
	for _, item := range *request.Exercises {
		selected = append(selected, service.SelectedExercise{
			ExerciseID: item.ExerciseID, Series: item.Series, Repetitions: item.Repetitions, Order: item.Order,
		})
	}
	created, err := c.service.CreateSession(r.Context(), identity.UserID, service.SessionInput{
		Name: request.Name, Description: request.Description, Exercises: selected,
	})
	if !c.handleError(w, err) {
		return
	}
	writeJSON(w, http.StatusCreated, sessionDetailResponse(created))
}

func (c *Controller) listSessions(w http.ResponseWriter, r *http.Request) {
	identity, ok := authenticatedIdentity(w, r)
	if !ok {
		return
	}
	sessions, err := c.service.ListSessions(r.Context(), identity.UserID)
	if !c.handleError(w, err) {
		return
	}
	response := make([]dto.SessionSummary, 0, len(sessions))
	for _, session := range sessions {
		response = append(response, dto.SessionSummary{ID: session.ID, Name: session.Name, Description: session.Description, ExerciseCount: session.ExerciseCount})
	}
	writeJSON(w, http.StatusOK, response)
}

func (c *Controller) getSession(w http.ResponseWriter, r *http.Request) {
	identity, ok := authenticatedIdentity(w, r)
	if !ok {
		return
	}
	id, ok := resourceID(w, r)
	if !ok {
		return
	}
	session, err := c.service.GetSession(r.Context(), identity.UserID, id)
	if !c.handleError(w, err) {
		return
	}
	writeJSON(w, http.StatusOK, sessionDetailResponse(session))
}

func (c *Controller) updateSession(w http.ResponseWriter, r *http.Request) {
	identity, ok := authenticatedIdentity(w, r)
	if !ok {
		return
	}
	id, ok := resourceID(w, r)
	if !ok {
		return
	}
	var request dto.CreateSessionRequest
	if err := decodeJSON(w, r, &request); err != nil || request.Exercises == nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Solicitud inválida.", nil)
		return
	}
	updated, err := c.service.UpdateSession(r.Context(), identity.UserID, id, sessionInput(request))
	if !c.handleError(w, err) {
		return
	}
	writeJSON(w, http.StatusOK, sessionDetailResponse(updated))
}

func (c *Controller) createRoutine(w http.ResponseWriter, r *http.Request) {
	var request dto.CreateRoutineRequest
	if err := decodeJSON(w, r, &request); err != nil || request.Sessions == nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Solicitud inválida.", nil)
		return
	}
	identity, ok := authenticatedIdentity(w, r)
	if !ok {
		return
	}
	selected := make([]service.SelectedSession, 0, len(*request.Sessions))
	for _, item := range *request.Sessions {
		selected = append(selected, service.SelectedSession{SessionID: item.SessionID, Day: item.Day})
	}
	created, err := c.service.CreateRoutine(r.Context(), identity.UserID, service.RoutineInput{
		Name: request.Name, Description: request.Description, Sessions: selected,
	})
	if !c.handleError(w, err) {
		return
	}
	writeJSON(w, http.StatusCreated, routineDetailResponse(created))
}

func (c *Controller) listRoutines(w http.ResponseWriter, r *http.Request) {
	identity, ok := authenticatedIdentity(w, r)
	if !ok {
		return
	}
	routines, err := c.service.ListRoutines(r.Context(), identity.UserID)
	if !c.handleError(w, err) {
		return
	}
	response := make([]dto.RoutineSummary, 0, len(routines))
	for _, routine := range routines {
		response = append(response, dto.RoutineSummary{ID: routine.ID, Name: routine.Name, Description: routine.Description})
	}
	writeJSON(w, http.StatusOK, response)
}

func (c *Controller) getRoutine(w http.ResponseWriter, r *http.Request) {
	identity, ok := authenticatedIdentity(w, r)
	if !ok {
		return
	}
	id, ok := resourceID(w, r)
	if !ok {
		return
	}
	routine, err := c.service.GetRoutine(r.Context(), identity.UserID, id)
	if !c.handleError(w, err) {
		return
	}
	writeJSON(w, http.StatusOK, routineDetailResponse(routine))
}

func (c *Controller) updateRoutine(w http.ResponseWriter, r *http.Request) {
	identity, ok := authenticatedIdentity(w, r)
	if !ok {
		return
	}
	id, ok := resourceID(w, r)
	if !ok {
		return
	}
	var request dto.CreateRoutineRequest
	if err := decodeJSON(w, r, &request); err != nil || request.Sessions == nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Solicitud inválida.", nil)
		return
	}
	updated, err := c.service.UpdateRoutine(r.Context(), identity.UserID, id, routineInput(request))
	if !c.handleError(w, err) {
		return
	}
	writeJSON(w, http.StatusOK, routineDetailResponse(updated))
}

func (c *Controller) deleteExercise(w http.ResponseWriter, r *http.Request) {
	c.deleteResource(w, r, c.service.DeleteExercise)
}

func (c *Controller) deleteSession(w http.ResponseWriter, r *http.Request) {
	c.deleteResource(w, r, c.service.DeleteSession)
}

func (c *Controller) deleteRoutine(w http.ResponseWriter, r *http.Request) {
	c.deleteResource(w, r, c.service.DeleteRoutine)
}

func (c *Controller) deleteResource(w http.ResponseWriter, r *http.Request, remove func(context.Context, int64, int64) error) {
	identity, ok := authenticatedIdentity(w, r)
	if !ok {
		return
	}
	id, ok := resourceID(w, r)
	if !ok {
		return
	}
	if !c.handleError(w, remove(r.Context(), identity.UserID, id)) {
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (c *Controller) handleError(w http.ResponseWriter, err error) bool {
	if err == nil {
		return true
	}
	var validation *routineserrors.ValidationError
	switch {
	case stderrors.As(err, &validation):
		writeError(w, http.StatusBadRequest, "validation_failed", "Revisá los campos indicados.", validation.Fields)
	case stderrors.Is(err, routineserrors.ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", "El recurso no está disponible.", nil)
	default:
		c.logger.Error("routines request failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "No se pudo completar la operación.", nil)
	}
	return false
}

func authenticatedIdentity(w http.ResponseWriter, r *http.Request) (requestctx.Identity, bool) {
	identity, err := requestctx.IdentityFrom(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid_token", "Token de acceso inválido o vencido.", nil)
		return requestctx.Identity{}, false
	}
	return identity, true
}

func resourceID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		writeError(w, http.StatusBadRequest, "invalid_request", "Identificador inválido.", nil)
		return 0, false
	}
	return id, true
}

func decodeJSON(w http.ResponseWriter, r *http.Request, destination any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBody)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !stderrors.Is(err, io.EOF) {
		if err == nil {
			return stderrors.New("request body must contain one JSON value")
		}
		return err
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		return
	}
}

func writeError(w http.ResponseWriter, status int, code, message string, fields map[string][]string) {
	writeJSON(w, status, dto.ErrorResponse{Error: dto.ErrorBody{Code: code, Message: message, Fields: fields}})
}

func exerciseResponse(exercise routines.Exercise) dto.Exercise {
	return dto.Exercise{
		ID: exercise.ID, Name: exercise.Name, Description: exercise.Description,
		ImageURL: exercise.ImageURL, VideoURL: exercise.VideoURL,
	}
}

func exerciseInput(request dto.CreateExerciseRequest) service.ExerciseInput {
	return service.ExerciseInput{
		Name: request.Name, Description: request.Description,
		ImageURL: request.ImageURL, VideoURL: request.VideoURL,
	}
}

func sessionInput(request dto.CreateSessionRequest) service.SessionInput {
	selected := make([]service.SelectedExercise, 0, len(*request.Exercises))
	for _, item := range *request.Exercises {
		selected = append(selected, service.SelectedExercise{
			ExerciseID: item.ExerciseID, Series: item.Series,
			Repetitions: item.Repetitions, Order: item.Order,
		})
	}
	return service.SessionInput{Name: request.Name, Description: request.Description, Exercises: selected}
}

func routineInput(request dto.CreateRoutineRequest) service.RoutineInput {
	selected := make([]service.SelectedSession, 0, len(*request.Sessions))
	for _, item := range *request.Sessions {
		selected = append(selected, service.SelectedSession{SessionID: item.SessionID, Day: item.Day})
	}
	return service.RoutineInput{Name: request.Name, Description: request.Description, Sessions: selected}
}

func sessionDetailResponse(session routines.Session) dto.SessionDetail {
	response := dto.SessionDetail{
		ID: session.ID, Name: session.Name, Description: session.Description,
		Exercises: make([]dto.SessionExercise, 0, len(session.Exercises)),
	}
	for _, item := range session.Exercises {
		response.Exercises = append(response.Exercises, dto.SessionExercise{
			Exercise: exerciseResponse(item.Exercise), Series: item.Series,
			Repetitions: item.Repetitions, Order: item.Order,
		})
	}
	return response
}

func routineDetailResponse(routine routines.Routine) dto.RoutineDetail {
	response := dto.RoutineDetail{
		ID: routine.ID, Name: routine.Name, Description: routine.Description,
		Sessions: make([]dto.RoutineSession, 0, len(routine.Sessions)),
	}
	for _, item := range routine.Sessions {
		response.Sessions = append(response.Sessions, dto.RoutineSession{
			Day: item.Day, Session: sessionDetailResponse(item.Session),
		})
	}
	return response
}
