package email

import (
	"fmt"
	"strings"
)

func Normalize(email string) (string, error) {
	parts := strings.Split(strings.TrimSpace(email), "@")
	if len(parts) != 2 {
		return email, fmt.Errorf("There must be one and only one '@'")
	}
	if parts[0] == "" || parts[1] == "" {
		return email, fmt.Errorf("Emails must be of the form 'user@domain'")
	}
	return strings.ToLower(strings.Join(parts, "@")), nil
}
