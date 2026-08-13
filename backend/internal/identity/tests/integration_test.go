//go:build integration

package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gmorini/inge-soft-3/backend/internal/identity/controller"
	"github.com/gmorini/inge-soft-3/backend/internal/identity/repository"
	"github.com/gmorini/inge-soft-3/backend/internal/identity/service"
)

func TestRegisterValidUser(t *testing.T) {
	handler := testHandler(t)

	tests := []struct {
		name      string
		apartment any
	}{
		{name: "with apartment", apartment: "2 B"},
		{name: "without apartment", apartment: nil},
	}
	for index, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body := validRegistrationBody(index, test.apartment)
			response := performJSONRequest(t, handler, http.MethodPost, "/api/auth/register", body, "")
			if response.Code != http.StatusCreated {
				t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
			}
			var payload map[string]any
			if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if payload["username"] != body["username"] || payload["email"] != body["email"] {
				t.Errorf("response identity = %#v", payload)
			}
			if _, exists := payload["password"]; exists {
				t.Fatal("response exposes password")
			}
			if _, exists := payload["passwordHash"]; exists {
				t.Fatal("response exposes password hash")
			}
		})
	}
}

func TestLogin(t *testing.T) {
	handler := testHandler(t)
	registration := validRegistrationBody(0, nil)
	response := performJSONRequest(t, handler, http.MethodPost, "/api/auth/register", registration, "")
	if response.Code != http.StatusCreated {
		t.Fatalf("register status = %d, body = %s", response.Code, response.Body.String())
	}

	success := performJSONRequest(t, handler, http.MethodPost, "/api/auth/login", map[string]any{
		"username": "ADA_01",
		"password": registration["password"],
	}, "")
	if success.Code != http.StatusOK {
		t.Fatalf("login status = %d, body = %s", success.Code, success.Body.String())
	}
	var loginPayload map[string]any
	if err := json.Unmarshal(success.Body.Bytes(), &loginPayload); err != nil {
		t.Fatalf("decode login: %v", err)
	}
	if loginPayload["tokenType"] != "Bearer" || loginPayload["expiresIn"] != float64(1800) {
		t.Errorf("login payload = %#v", loginPayload)
	}
	if loginPayload["accessToken"] == "" {
		t.Fatal("login returned empty access token")
	}

	wrongPassword := performJSONRequest(t, handler, http.MethodPost, "/api/auth/login", map[string]any{
		"username": "ada_01",
		"password": "Incorrecta!123",
	}, "")
	unknownUser := performJSONRequest(t, handler, http.MethodPost, "/api/auth/login", map[string]any{
		"username": "nadie",
		"password": "Incorrecta!123",
	}, "")
	if wrongPassword.Code != http.StatusUnauthorized || unknownUser.Code != http.StatusUnauthorized {
		t.Fatalf("credential statuses = %d and %d", wrongPassword.Code, unknownUser.Code)
	}
	if wrongPassword.Body.String() != unknownUser.Body.String() {
		t.Errorf("credential responses differ: %q versus %q", wrongPassword.Body.String(), unknownUser.Body.String())
	}

	for range 3 {
		performJSONRequest(t, handler, http.MethodPost, "/api/auth/login", map[string]any{
			"username": "ada_01",
			"password": "Incorrecta!123",
		}, "")
	}
	retry := performJSONRequest(t, handler, http.MethodPost, "/api/auth/login", map[string]any{
		"username": "ada_01",
		"password": registration["password"],
	}, "")
	if retry.Code != http.StatusOK {
		t.Fatalf("login after failures status = %d", retry.Code)
	}
}

func TestAuthenticatedIdentity(t *testing.T) {
	handler := testHandler(t)
	registration := validRegistrationBody(0, nil)
	registerResponse := performJSONRequest(t, handler, http.MethodPost, "/api/auth/register", registration, "")
	if registerResponse.Code != http.StatusCreated {
		t.Fatalf("register status = %d", registerResponse.Code)
	}
	var registered map[string]any
	if err := json.Unmarshal(registerResponse.Body.Bytes(), &registered); err != nil {
		t.Fatalf("decode registration: %v", err)
	}

	loginResponse := performJSONRequest(t, handler, http.MethodPost, "/api/auth/login", map[string]any{
		"username": registration["username"], "password": registration["password"],
	}, "")
	var loggedIn map[string]any
	if err := json.Unmarshal(loginResponse.Body.Bytes(), &loggedIn); err != nil {
		t.Fatalf("decode login: %v", err)
	}
	token, ok := loggedIn["accessToken"].(string)
	if !ok || token == "" {
		t.Fatalf("access token = %#v", loggedIn["accessToken"])
	}

	meResponse := performJSONRequest(t, handler, http.MethodGet, "/api/auth/me", nil, token)
	if meResponse.Code != http.StatusOK {
		t.Fatalf("me status = %d, body = %s", meResponse.Code, meResponse.Body.String())
	}
	var current map[string]any
	if err := json.Unmarshal(meResponse.Body.Bytes(), &current); err != nil {
		t.Fatalf("decode current identity: %v", err)
	}
	if current["id"] != registered["id"] || current["username"] != registered["username"] {
		t.Errorf("current identity = %#v, registered = %#v", current, registered)
	}

	for _, invalid := range []string{"", "malformed", alterToken(token)} {
		response := performJSONRequest(t, handler, http.MethodGet, "/api/auth/me", nil, invalid)
		if response.Code != http.StatusUnauthorized {
			t.Errorf("invalid token status = %d", response.Code)
		}
	}
}

func alterToken(token string) string {
	separator := strings.LastIndexByte(token, '.')
	if separator < 0 || separator == len(token)-1 {
		return token + "x"
	}
	replacement := byte('a')
	if token[separator+1] == replacement {
		replacement = 'b'
	}
	return token[:separator+1] + string(replacement) + token[separator+2:]
}

func TestRegistrationErrors(t *testing.T) {
	handler := testHandler(t)
	invalid := validRegistrationBody(0, nil)
	invalid["phone"] = "351-555"
	invalid["username"] = " bad user "
	invalid["password"] = "weak"
	response := performJSONRequest(t, handler, http.MethodPost, "/api/auth/register", invalid, "")
	if response.Code != http.StatusBadRequest {
		t.Fatalf("invalid registration status = %d", response.Code)
	}
	var payload struct {
		Error struct {
			Fields map[string][]string `json:"fields"`
		} `json:"error"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode validation: %v", err)
	}
	for _, field := range []string{"phone", "username", "password"} {
		if len(payload.Error.Fields[field]) == 0 {
			t.Errorf("missing %s error in %#v", field, payload.Error.Fields)
		}
	}

	valid := validRegistrationBody(1, nil)
	first := performJSONRequest(t, handler, http.MethodPost, "/api/auth/register", valid, "")
	if first.Code != http.StatusCreated {
		t.Fatalf("seed status = %d", first.Code)
	}
	duplicate := validRegistrationBody(2, nil)
	duplicate["email"] = strings.ToUpper(valid["email"].(string))
	conflict := performJSONRequest(t, handler, http.MethodPost, "/api/auth/register", duplicate, "")
	if conflict.Code != http.StatusConflict {
		t.Fatalf("conflict status = %d, body = %s", conflict.Code, conflict.Body.String())
	}
}

func TestRegistrationRejectsInvalidJSON(t *testing.T) {
	handler := testHandler(t)
	tests := []struct {
		name string
		body string
	}{
		{name: "malformed", body: "{"},
		{name: "unknown field", body: `{"unknown":true}`},
		{name: "oversized", body: `{"padding":"` + strings.Repeat("x", 17<<10) + `"}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(test.body))
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != http.StatusBadRequest {
				t.Errorf("status = %d", response.Code)
			}
		})
	}
}

func validRegistrationBody(index int, apartment any) map[string]any {
	return map[string]any{
		"firstName": "Ada",
		"lastName":  "Lovelace",
		"phone":     "+5493515551234",
		"address": map[string]any{
			"street":    "San Martín",
			"number":    "123 Bis",
			"apartment": apartment,
			"city":      "Córdoba",
			"province":  "Córdoba",
		},
		"username": "ada_0" + string(rune('1'+index)),
		"email":    "ada" + string(rune('1'+index)) + "@example.com",
		"password": "Segura!@123",
	}
}

func performJSONRequest(
	t *testing.T,
	handler http.Handler,
	method string,
	path string,
	body any,
	token string,
) *httptest.ResponseRecorder {
	t.Helper()
	encoded, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("encode request: %v", err)
	}
	request := httptest.NewRequest(method, path, bytes.NewReader(encoded))
	request.Header.Set("Content-Type", "application/json")
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func testHandler(t *testing.T) http.Handler {
	t.Helper()
	pool := openTestPool(t)
	applyMigration(t, pool)
	repository := repository.New(pool)
	tokenManager, err := service.NewTokenManager("test-secret-with-at-least-32-bytes")
	if err != nil {
		t.Fatalf("create token manager: %v", err)
	}
	identityService, err := service.New(repository, tokenManager)
	if err != nil {
		t.Fatalf("create identity service: %v", err)
	}
	identityController := controller.New(
		identityService,
		tokenManager,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
	mux := http.NewServeMux()
	identityController.RegisterRoutes(mux)
	return mux
}
