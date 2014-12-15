package email

import (
	"testing"
)

// Create a test sender
type testSender struct{}

func (ts *testSender) Send(to, subject, body string) error {
	return nil
}

func TestSender(t *testing.T) {
	// DefaultSender should implement the Sender interface
	var _ Sender = DefaultSender{}

	// Create an email config
	ts := &testSender{}
	if err := ts.Send("me", "you", "hello"); err != nil {
		t.Fatal(err)
	}
}
