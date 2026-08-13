package service

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	identityerrors "github.com/gmorini/inge-soft-3/backend/internal/identity/errors"
	"github.com/golang-jwt/jwt/v5"
)

func TestNormalizeAndValidateRegistration(t *testing.T) {
	tests := []struct {
		name          string
		apartment     string
		wantApartment *string
	}{
		{name: "complete address", apartment: " 2 B ", wantApartment: stringPointer("2 B")},
		{name: "address without apartment", apartment: "  ", wantApartment: nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := validRegistrationInput()
			input.Apartment = test.apartment

			got, err := normalizeAndValidateRegistration(input)
			if err != nil {
				t.Fatalf("normalizeAndValidateRegistration() error: %v", err)
			}
			if got.FirstName != "Ada" || got.LastName != "Lovelace" {
				t.Errorf("normalized name = %q %q", got.FirstName, got.LastName)
			}
			if got.Username != "ada_01" || got.Email != "ada@example.com" {
				t.Errorf("canonical identity = %q %q", got.Username, got.Email)
			}
			if !equalStringPointers(got.Apartment, test.wantApartment) {
				t.Errorf("apartment = %v, want %v", got.Apartment, test.wantApartment)
			}
		})
	}
}

func TestHashAndVerifyPassword(t *testing.T) {
	password := "Segura!@123"
	hash, err := hashPassword(password)
	if err != nil {
		t.Fatalf("hashPassword() error: %v", err)
	}
	if strings.Contains(hash, password) {
		t.Fatal("password hash contains plaintext")
	}
	if !strings.HasPrefix(hash, "$argon2id$") {
		t.Fatalf("hash prefix = %q", hash)
	}

	valid, err := verifyPassword(password, hash)
	if err != nil {
		t.Fatalf("verifyPassword() error: %v", err)
	}
	if !valid {
		t.Fatal("correct password was rejected")
	}

	valid, err = verifyPassword("Incorrecta!123", hash)
	if err != nil {
		t.Fatalf("verifyPassword() wrong password error: %v", err)
	}
	if valid {
		t.Fatal("wrong password was accepted")
	}
}

func TestTokenManager_Issue(t *testing.T) {
	issuedAt := time.Date(2026, time.August, 13, 12, 0, 0, 0, time.UTC)
	manager, err := newTokenManager(strings.Repeat("s", 32), func() time.Time { return issuedAt })
	if err != nil {
		t.Fatalf("newTokenManager() error: %v", err)
	}

	token, err := manager.Issue(42, "ada_01")
	if err != nil {
		t.Fatalf("Issue() error: %v", err)
	}

	parsed, _, err := new(jwt.Parser).ParseUnverified(token, jwt.MapClaims{})
	if err != nil {
		t.Fatalf("parse issued token: %v", err)
	}
	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatalf("claims type = %T", parsed.Claims)
	}
	encoded, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("marshal claims: %v", err)
	}
	var values map[string]any
	if err := json.Unmarshal(encoded, &values); err != nil {
		t.Fatalf("unmarshal claims: %v", err)
	}
	if len(values) != 4 {
		t.Fatalf("claims = %#v, want exactly four", values)
	}
	if values["sub"] != "42" || values["username"] != "ada_01" {
		t.Errorf("identity claims = %#v", values)
	}
	if values["exp"].(float64)-values["iat"].(float64) != 1800 {
		t.Errorf("token lifetime = %v seconds", values["exp"].(float64)-values["iat"].(float64))
	}
}

func TestTokenManager_Validate(t *testing.T) {
	issuedAt := time.Date(2026, time.August, 13, 12, 0, 0, 0, time.UTC)
	now := issuedAt
	manager, err := newTokenManager(strings.Repeat("s", 32), func() time.Time { return now })
	if err != nil {
		t.Fatalf("newTokenManager() error: %v", err)
	}
	token, err := manager.Issue(42, "ada_01")
	if err != nil {
		t.Fatalf("Issue() error: %v", err)
	}

	tests := []struct {
		name      string
		now       time.Time
		token     func() string
		wantValid bool
	}{
		{name: "immediately before expiration", now: issuedAt.Add(30*time.Minute - time.Second), token: func() string { return token }, wantValid: true},
		{name: "exactly at expiration", now: issuedAt.Add(30 * time.Minute), token: func() string { return token }},
		{name: "altered signature", now: issuedAt, token: func() string { return token[:len(token)-1] + "x" }},
		{name: "malformed token", now: issuedAt, token: func() string { return "not-a-token" }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			now = test.now
			identity, err := manager.Validate(test.token())
			if test.wantValid {
				if err != nil {
					t.Fatalf("Validate() error: %v", err)
				}
				if identity.UserID != 42 || identity.Username != "ada_01" {
					t.Errorf("identity = %+v", identity)
				}
				return
			}
			if err == nil {
				t.Fatalf("Validate() identity = %+v, want error", identity)
			}
		})
	}

	noneToken := jwt.NewWithClaims(jwt.SigningMethodNone, tokenClaims{
		Username: "ada_01",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "42", IssuedAt: jwt.NewNumericDate(issuedAt), ExpiresAt: jwt.NewNumericDate(issuedAt.Add(tokenLifetime)),
		},
	})
	unsigned, err := noneToken.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("create none token: %v", err)
	}
	if _, err := manager.Validate(unsigned); err == nil {
		t.Fatal("Validate() accepted none algorithm")
	}
}

func TestNormalizeAndValidateRegistration_InvalidFields(t *testing.T) {
	tests := []struct {
		name      string
		modify    func(*RegisterInput)
		wantField string
	}{
		{name: "short username", modify: func(input *RegisterInput) { input.Username = "ab" }, wantField: "username"},
		{name: "long username", modify: func(input *RegisterInput) { input.Username = strings.Repeat("a", 31) }, wantField: "username"},
		{name: "leading username space", modify: func(input *RegisterInput) { input.Username = " ada_01" }, wantField: "username"},
		{name: "middle username space", modify: func(input *RegisterInput) { input.Username = "ada 01" }, wantField: "username"},
		{name: "invalid phone separators", modify: func(input *RegisterInput) { input.Phone = "+54 9 351" }, wantField: "phone"},
		{name: "invalid phone international prefix", modify: func(input *RegisterInput) { input.Phone = "005493515551234" }, wantField: "phone"},
		{name: "short password", modify: func(input *RegisterInput) { input.Password = "A!1" }, wantField: "password"},
		{name: "missing uppercase", modify: func(input *RegisterInput) { input.Password = "segura!!123" }, wantField: "password"},
		{name: "missing digits", modify: func(input *RegisterInput) { input.Password = "Segura!!12" }, wantField: "password"},
		{name: "missing symbols", modify: func(input *RegisterInput) { input.Password = "Segura123" }, wantField: "password"},
		{name: "blank trimmed name", modify: func(input *RegisterInput) { input.FirstName = "  " }, wantField: "firstName"},
		{name: "street number too long", modify: func(input *RegisterInput) { input.StreetNumber = strings.Repeat("1", 21) }, wantField: "number"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := validRegistrationInput()
			test.modify(&input)
			_, err := normalizeAndValidateRegistration(input)
			var validation *identityerrors.ValidationError
			if !errors.As(err, &validation) {
				t.Fatalf("error = %v, want ValidationError", err)
			}
			if len(validation.Fields[test.wantField]) == 0 {
				t.Errorf("fields = %#v, want %q", validation.Fields, test.wantField)
			}
		})
	}
}

func TestNormalizeAndValidateRegistration_PasswordThresholds(t *testing.T) {
	input := validRegistrationInput()
	input.Password = "A!!111aa"
	if _, err := normalizeAndValidateRegistration(input); err != nil {
		t.Fatalf("exact password thresholds rejected: %v", err)
	}
	input.Password = "A@@111aa"
	if _, err := normalizeAndValidateRegistration(input); err != nil {
		t.Fatalf("repeated symbol and digit counts rejected: %v", err)
	}
	input.StreetNumber = strings.Repeat("1", 20)
	if _, err := normalizeAndValidateRegistration(input); err != nil {
		t.Fatalf("20-character street number rejected: %v", err)
	}
}

func validRegistrationInput() RegisterInput {
	return RegisterInput{
		FirstName:    " Ada ",
		LastName:     " Lovelace ",
		Phone:        "+5493515551234",
		Street:       " San Martín ",
		StreetNumber: " 123 Bis ",
		Apartment:    " 2 B ",
		City:         " Córdoba ",
		Province:     " Córdoba ",
		Username:     "Ada_01",
		Email:        " ADA@example.com ",
		Password:     "Segura!@123",
	}
}

func stringPointer(value string) *string {
	return &value
}

func equalStringPointers(left, right *string) bool {
	if left == nil || right == nil {
		return left == right
	}
	return *left == *right
}
