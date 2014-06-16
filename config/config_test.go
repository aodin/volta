package config

import (
	"testing"
)

func TestConfig(t *testing.T) {
	// Parse the local settings.json file
	c, err := Parse()
	if err != nil {
		t.Fatalf("Error during Parse(): %s", err)
	}

	// Test the parent config methods
	if c.Address() != "localhost:9001" {
		t.Errorf("Unexpected address: %s", c.Address())
	}

	// Test the SMTP config methods
	if c.SMTP.FromAddress() != `"Example User" <no_reply@example.com>` {
		t.Errorf("Unexpected from address: %s", c.SMTP.FromAddress())
	}

	if c.SMTP.Address() != "example.com:587" {
		t.Errorf("Unexpected SMTP address: %s", c.SMTP.Address())
	}

	// Test the database config methods
	x := "host=localhost port=5432 dbname=db user=pg password=pass"
	if c.Database.Credentials() != x {
		t.Errorf("Unexpected db credentials: %s", c.Database.Credentials())
	}
}
