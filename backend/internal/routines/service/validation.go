package service

import (
	"fmt"
	"net/url"
	"strings"
	"unicode"
	"unicode/utf8"

	routineserrors "github.com/gmorini/inge-soft-3/backend/internal/routines/errors"
)

const (
	maxNameLength        = 100
	maxDescriptionLength = 500
	maxURLLength         = 2048
)

func validateExercise(input ExerciseInput) (normalizedExercise, error) {
	fields := make(map[string][]string)
	validateRequiredText(fields, "name", input.Name, maxNameLength)
	description := validateOptionalText(fields, "description", input.Description, maxDescriptionLength)
	imageURL := validateOptionalURL(fields, "imageUrl", input.ImageURL)
	videoURL := validateOptionalURL(fields, "videoUrl", input.VideoURL)
	if len(fields) > 0 {
		return normalizedExercise{}, &routineserrors.ValidationError{Fields: fields}
	}
	return normalizedExercise{Name: input.Name, Description: description, ImageURL: imageURL, VideoURL: videoURL}, nil
}

func validateSession(input SessionInput) (normalizedSession, error) {
	fields := make(map[string][]string)
	validateRequiredText(fields, "name", input.Name, maxNameLength)
	description := validateOptionalText(fields, "description", input.Description, maxDescriptionLength)
	seenIDs := make(map[int64]struct{}, len(input.Exercises))
	seenOrders := make(map[int32]struct{}, len(input.Exercises))
	selected := make([]SelectedExercise, 0, len(input.Exercises))
	for index, item := range input.Exercises {
		prefix := fmt.Sprintf("exercises.%d.", index)
		if item.ExerciseID < 1 {
			addFieldError(fields, prefix+"exerciseId", "Debe ser un identificador positivo.")
		} else if _, exists := seenIDs[item.ExerciseID]; exists {
			addFieldError(fields, prefix+"exerciseId", "El ejercicio no puede repetirse.")
		}
		seenIDs[item.ExerciseID] = struct{}{}
		if item.Series < 0 {
			addFieldError(fields, prefix+"series", "Debe ser un número entero no negativo.")
		}
		if item.Repetitions < 0 {
			addFieldError(fields, prefix+"repetitions", "Debe ser un número entero no negativo.")
		}
		expected := int32(index + 1)
		if item.Order != expected {
			addFieldError(fields, prefix+"order", fmt.Sprintf("Debe ser %d.", expected))
		}
		if _, exists := seenOrders[item.Order]; exists {
			addFieldError(fields, prefix+"order", "El orden no puede repetirse.")
		}
		seenOrders[item.Order] = struct{}{}
		selected = append(selected, item)
	}
	if len(fields) > 0 {
		return normalizedSession{}, &routineserrors.ValidationError{Fields: fields}
	}
	return normalizedSession{Name: input.Name, Description: description, Exercises: selected}, nil
}

func validateRoutine(input RoutineInput) (normalizedRoutine, error) {
	fields := make(map[string][]string)
	validateRequiredText(fields, "name", input.Name, maxNameLength)
	description := validateOptionalText(fields, "description", input.Description, maxDescriptionLength)
	seen := make(map[[2]int64]struct{}, len(input.Sessions))
	selected := make([]SelectedSession, 0, len(input.Sessions))
	for index, item := range input.Sessions {
		prefix := fmt.Sprintf("sessions.%d.", index)
		if item.SessionID < 1 {
			addFieldError(fields, prefix+"sessionId", "Debe ser un identificador positivo.")
		}
		if item.Day < 1 || item.Day > 7 {
			addFieldError(fields, prefix+"day", "Debe estar entre 1 y 7.")
		}
		key := [2]int64{item.SessionID, int64(item.Day)}
		if _, exists := seen[key]; exists {
			addFieldError(fields, prefix+"sessionId", "La sesión ya está asignada a ese día.")
		}
		seen[key] = struct{}{}
		selected = append(selected, item)
	}
	if len(fields) > 0 {
		return normalizedRoutine{}, &routineserrors.ValidationError{Fields: fields}
	}
	return normalizedRoutine{Name: input.Name, Description: description, Sessions: selected}, nil
}

func validateRequiredText(fields map[string][]string, field, value string, maximum int) {
	if value == "" {
		addFieldError(fields, field, "Es obligatorio.")
		return
	}
	validateText(fields, field, value, maximum)
}

func validateOptionalText(fields map[string][]string, field, value string, maximum int) *string {
	if value == "" {
		return nil
	}
	validateText(fields, field, value, maximum)
	return &value
}

func validateText(fields map[string][]string, field, value string, maximum int) {
	if utf8.RuneCountInString(value) > maximum {
		addFieldError(fields, field, fmt.Sprintf("No puede superar %d caracteres.", maximum))
	}
	if strings.HasPrefix(value, " ") || strings.HasSuffix(value, " ") || strings.Contains(value, "  ") {
		addFieldError(fields, field, "No puede tener espacios al inicio, al final ni repetidos.")
	}
	for _, character := range value {
		if unicode.IsSpace(character) && character != ' ' {
			addFieldError(fields, field, "Solo puede usar un espacio entre palabras.")
			break
		}
	}
}

func validateOptionalURL(fields map[string][]string, field, value string) *string {
	if value == "" {
		return nil
	}
	if utf8.RuneCountInString(value) > maxURLLength {
		addFieldError(fields, field, "No puede superar 2048 caracteres.")
		return &value
	}
	for _, character := range value {
		if unicode.IsSpace(character) {
			addFieldError(fields, field, "Debe ser una URL HTTP o HTTPS válida.")
			return &value
		}
	}
	parsed, err := url.Parse(value)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		addFieldError(fields, field, "Debe ser una URL HTTP o HTTPS válida.")
	}
	return &value
}

func addFieldError(fields map[string][]string, field, message string) {
	fields[field] = append(fields[field], message)
}
