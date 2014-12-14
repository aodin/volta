package router

import (
	"testing"
)

func TestParams(t *testing.T) {
	// Create a param
	id := Param{Key: "id", Value: "1"}

	// Create a slice of params
	params := Params{id}

	// Get the id back out
	if exists := params.ByName("id"); exists != "1" {
		t.Fatalf("unexpected value of id parameter: %s", exists)
	}

	// Get a parameter that doesn't exist
	if doesnt := params.ByName("name"); doesnt != "" {
		t.Fatalf("unexpected value returned by name parameter: %s", doesnt)
	}
}
