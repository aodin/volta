package config

import (
	"fmt"
)

// DatabaseConfig contains the fields needed to connect to a database.
type DatabaseConfig struct {
	Driver   string `json:"driver"`
	Host     string `json:"host"`
	Port     int64  `json:"port"`
	Name     string `json:"name"`
	User     string `json:"user"`
	Password string `json:"password"`
}

// Credentials with return a string of credentials appropriate for Go's
// sql.Open function
func (db DatabaseConfig) Credentials() string {
	// TODO Are there different credentials for different drivers?
	// TODO Only add the key=value pair if there is a value
	return fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s",
		db.Host,
		db.Port,
		db.Name,
		db.User,
		db.Password,
	)
}
