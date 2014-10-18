package templates

import (
	"testing"
)

func TestAttrs(t *testing.T) {
	// Create a new attrs map
	a := Attrs{"Greeting": "Hello"}

	// And another
	b := Attrs{"Greeting": "Yo", "ID": 1}

	// Calling the method merge on a should overwrite any duplicate values
	// on a with those in b
	a.Merge(b)

	// The greeting should be overwritten - remember it is an interface{}!
	greetingAttr := a["Greeting"]
	greeting, ok := greetingAttr.(string)
	if !ok {
		t.Fatalf("The Greeting attr could not be cast to a string")
	}
	if greeting != "Yo" {
		t.Fatalf("The Greeting attr was not Yo")
	}

	// And the ID should be added
	IDAttr := a["ID"]
	ID, ok := IDAttr.(int)
	if !ok {
		t.Fatalf("The ID attr could not be cast to an int")
	}
	if ID != 1 {
		t.Fatalf("The ID attr was not 1")
	}
}
