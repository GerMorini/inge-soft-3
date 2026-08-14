package controller

import (
	"testing"

	routines "github.com/gmorini/inge-soft-3/backend/internal/routines"
)

func TestExerciseResponseMapsModuleType(t *testing.T) {
	description := "Con barra"
	response := exerciseResponse(routines.Exercise{ID: 7, Name: "Sentadilla", Description: &description})
	if response.ID != 7 || response.Name != "Sentadilla" || response.Description == nil || *response.Description != description {
		t.Fatalf("response = %+v", response)
	}
}
