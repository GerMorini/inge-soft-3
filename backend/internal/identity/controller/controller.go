package controller

import (
	"encoding/json"
	stderrors "errors"
	"io"
	"log/slog"
	"net/http"

	identityerrors "github.com/gmorini/inge-soft-3/backend/internal/identity/errors"
	"github.com/gmorini/inge-soft-3/backend/internal/identity/service"
	"github.com/gmorini/inge-soft-3/backend/internal/platform/requestctx"
)

const maxRequestBody = 16 << 10

type errorResponse struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Code    string              `json:"code"`
	Message string              `json:"message"`
	Fields  map[string][]string `json:"fields,omitempty"`
}

type Controller struct {
	service *service.Service
	tokens  *service.TokenManager
	logger  *slog.Logger
}

type registerRequest struct {
	FirstName string         `json:"firstName"`
	LastName  string         `json:"lastName"`
	Phone     string         `json:"phone"`
	Address   addressRequest `json:"address"`
	Username  string         `json:"username"`
	Email     string         `json:"email"`
	Password  string         `json:"password"`
}

type addressRequest struct {
	Street    string `json:"street"`
	Number    string `json:"number"`
	Apartment string `json:"apartment"`
	City      string `json:"city"`
	Province  string `json:"province"`
}

type registeredUserResponse struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	AccessToken string `json:"accessToken"`
	TokenType   string `json:"tokenType"`
	ExpiresIn   int    `json:"expiresIn"`
}

type currentUserResponse struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
}

func New(
	service *service.Service,
	tokens *service.TokenManager,
	logger *slog.Logger,
) *Controller {
	return &Controller{service: service, tokens: tokens, logger: logger}
}

func (c *Controller) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/auth/register", c.register)
	mux.HandleFunc("POST /api/auth/login", c.login)
	mux.Handle("GET /api/auth/me", authenticate(c.tokens, c.logger, http.HandlerFunc(c.currentUser)))
}

func (c *Controller) currentUser(w http.ResponseWriter, r *http.Request) {
	identity, err := requestctx.IdentityFrom(r.Context())
	if err != nil {
		writeInvalidToken(w)
		return
	}
	writeJSON(w, http.StatusOK, currentUserResponse{ID: identity.UserID, Username: identity.Username})
}

func (c *Controller) login(w http.ResponseWriter, r *http.Request) {
	var request loginRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Solicitud inválida.", nil)
		return
	}

	result, err := c.service.Login(r.Context(), service.LoginInput{
		Username: request.Username,
		Password: request.Password,
	})
	if err != nil {
		var validation *identityerrors.ValidationError
		switch {
		case stderrors.As(err, &validation):
			writeError(w, http.StatusBadRequest, "validation_failed", "Revisá los campos indicados.", validation.Fields)
		case stderrors.Is(err, identityerrors.ErrInvalidCredentials):
			writeError(w, http.StatusUnauthorized, "invalid_credentials", "Username o contraseña inválidos.", nil)
		default:
			c.logger.Error("login failed", "error", err)
			writeError(w, http.StatusInternalServerError, "internal_error", "No se pudo completar la operación.", nil)
		}
		return
	}

	writeJSON(w, http.StatusOK, loginResponse{
		AccessToken: result.AccessToken,
		TokenType:   "Bearer",
		ExpiresIn:   result.ExpiresIn,
	})
}

func (c *Controller) register(w http.ResponseWriter, r *http.Request) {
	var request registerRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Solicitud inválida.", nil)
		return
	}

	created, err := c.service.Register(r.Context(), service.RegisterInput{
		FirstName:    request.FirstName,
		LastName:     request.LastName,
		Phone:        request.Phone,
		Street:       request.Address.Street,
		StreetNumber: request.Address.Number,
		Apartment:    request.Address.Apartment,
		City:         request.Address.City,
		Province:     request.Address.Province,
		Username:     request.Username,
		Email:        request.Email,
		Password:     request.Password,
	})
	if err != nil {
		var validation *identityerrors.ValidationError
		var conflict *identityerrors.ConflictError
		switch {
		case stderrors.As(err, &validation):
			writeError(
				w,
				http.StatusBadRequest,
				"validation_failed",
				"Revisá los campos indicados.",
				validation.Fields,
			)
			return
		case stderrors.As(err, &conflict):
			writeError(
				w,
				http.StatusConflict,
				"registration_conflict",
				"Ya existe una cuenta con ese dato.",
				map[string][]string{conflict.Field: {"Ya está registrado."}},
			)
			return
		}
		c.logger.Error("registration failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "No se pudo completar la operación.", nil)
		return
	}

	writeJSON(w, http.StatusCreated, registeredUserResponse{
		ID:       created.ID,
		Username: created.Username,
		Email:    created.Email,
	})
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
	payload, err := json.Marshal(value)
	if err != nil {
		http.Error(w, "No se pudo completar la operación.", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(append(payload, '\n')); err != nil {
		return
	}
}

func writeError(
	w http.ResponseWriter,
	status int,
	code string,
	message string,
	fields map[string][]string,
) {
	writeJSON(w, status, errorResponse{Error: errorBody{
		Code:    code,
		Message: message,
		Fields:  fields,
	}})
}
