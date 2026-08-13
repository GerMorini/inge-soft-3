package service

import (
	"net/mail"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	identityerrors "github.com/gmorini/inge-soft-3/backend/internal/identity/errors"
)

var (
	phonePattern    = regexp.MustCompile(`^\+[1-9][0-9]{1,14}$`)
	usernamePattern = regexp.MustCompile(`^[A-Za-z0-9._-]{3,30}$`)
)

type normalizedRegistration struct {
	FirstName    string
	LastName     string
	Phone        string
	Street       string
	StreetNumber string
	Apartment    *string
	City         string
	Province     string
	Username     string
	Email        string
	Password     string
}

func normalizeAndValidateRegistration(input RegisterInput) (normalizedRegistration, error) {
	normalized := normalizedRegistration{
		FirstName:    strings.TrimSpace(input.FirstName),
		LastName:     strings.TrimSpace(input.LastName),
		Phone:        input.Phone,
		Street:       strings.TrimSpace(input.Street),
		StreetNumber: strings.TrimSpace(input.StreetNumber),
		City:         strings.TrimSpace(input.City),
		Province:     strings.TrimSpace(input.Province),
		Username:     strings.ToLower(input.Username),
		Email:        strings.ToLower(strings.TrimSpace(input.Email)),
		Password:     input.Password,
	}
	if apartment := strings.TrimSpace(input.Apartment); apartment != "" {
		normalized.Apartment = &apartment
	}

	fields := make(map[string][]string)
	validateRequiredText(fields, "firstName", normalized.FirstName, 100)
	validateRequiredText(fields, "lastName", normalized.LastName, 100)
	validateRequiredText(fields, "street", normalized.Street, 120)
	validateRequiredText(fields, "number", normalized.StreetNumber, 20)
	validateRequiredText(fields, "city", normalized.City, 100)
	validateRequiredText(fields, "province", normalized.Province, 100)
	if normalized.Apartment != nil && utf8.RuneCountInString(*normalized.Apartment) > 20 {
		fields["apartment"] = append(fields["apartment"], "Debe tener hasta 20 caracteres.")
	}
	if !phonePattern.MatchString(normalized.Phone) {
		fields["phone"] = append(fields["phone"], "Debe usar formato E.164, por ejemplo +5493515551234.")
	}
	if !usernamePattern.MatchString(input.Username) {
		fields["username"] = append(
			fields["username"],
			"Debe tener entre 3 y 30 caracteres, sin espacios, usando letras, números, punto, guion o guion bajo.",
		)
	}
	if !validEmail(normalized.Email) {
		fields["email"] = append(fields["email"], "Debe ser un email válido.")
	}
	validatePassword(fields, normalized.Password)

	if len(fields) > 0 {
		return normalizedRegistration{}, &identityerrors.ValidationError{Fields: fields}
	}
	return normalized, nil
}

func validateRequiredText(fields map[string][]string, field, value string, maximum int) {
	if value == "" {
		fields[field] = append(fields[field], "Es obligatorio.")
		return
	}
	if utf8.RuneCountInString(value) > maximum {
		fields[field] = append(fields[field], "Supera la longitud máxima permitida.")
	}
}

func validEmail(value string) bool {
	if value == "" || utf8.RuneCountInString(value) > 254 {
		return false
	}
	parsed, err := mail.ParseAddress(value)
	return err == nil && parsed.Address == value
}

func validatePassword(fields map[string][]string, password string) {
	if utf8.RuneCountInString(password) < 8 {
		fields["password"] = append(fields["password"], "Debe tener al menos 8 caracteres.")
	}

	uppercase, digits, symbols := 0, 0, 0
	for _, character := range password {
		switch {
		case unicode.IsUpper(character):
			uppercase++
		case unicode.IsDigit(character):
			digits++
		case !unicode.IsLetter(character) && !unicode.IsSpace(character):
			symbols++
		}
	}
	if uppercase < 1 {
		fields["password"] = append(fields["password"], "Debe incluir una letra mayúscula.")
	}
	if digits < 3 {
		fields["password"] = append(fields["password"], "Debe incluir al menos 3 números.")
	}
	if symbols < 2 {
		fields["password"] = append(fields["password"], "Debe incluir al menos 2 símbolos especiales.")
	}
}
