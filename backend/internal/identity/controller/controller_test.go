package controller

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gmorini/inge-soft-3/backend/internal/identity/service"
	"github.com/gmorini/inge-soft-3/backend/internal/platform/requestctx"
)

func TestAuthenticate(t *testing.T) {
	manager, err := service.NewTokenManager(strings.Repeat("s", 32))
	if err != nil {
		t.Fatalf("NewTokenManager() error: %v", err)
	}
	validToken, err := manager.Issue(42, "ada_01")
	if err != nil {
		t.Fatalf("Issue() error: %v", err)
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	tests := []struct {
		name       string
		header     string
		wantStatus int
		wantCalled bool
	}{
		{name: "valid bearer", header: "Bearer " + validToken, wantStatus: http.StatusOK, wantCalled: true},
		{name: "missing header", wantStatus: http.StatusUnauthorized},
		{name: "wrong scheme", header: "Basic " + validToken, wantStatus: http.StatusUnauthorized},
		{name: "malformed bearer", header: "Bearer invalid", wantStatus: http.StatusUnauthorized},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			called := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				identity, err := requestctx.IdentityFrom(r.Context())
				if err != nil {
					t.Fatalf("IdentityFrom() error: %v", err)
				}
				if identity.UserID != 42 || identity.Username != "ada_01" {
					t.Errorf("identity = %+v", identity)
				}
				w.WriteHeader(http.StatusOK)
			})
			request := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
			request.Header.Set("Authorization", test.header)
			response := httptest.NewRecorder()

			authenticate(manager, logger, next).ServeHTTP(response, request)

			if response.Code != test.wantStatus {
				t.Errorf("status = %d, want %d", response.Code, test.wantStatus)
			}
			if called != test.wantCalled {
				t.Errorf("handler called = %t, want %t", called, test.wantCalled)
			}
		})
	}
}

func TestCurrentUserRequiresIdentity(t *testing.T) {
	controller := &Controller{logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	request := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	response := httptest.NewRecorder()
	controller.currentUser(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", response.Code)
	}

	request = request.WithContext(requestctx.WithIdentity(context.Background(), requestctx.Identity{
		UserID: 42, Username: "ada_01",
	}))
	response = httptest.NewRecorder()
	controller.currentUser(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d", response.Code)
	}
}
