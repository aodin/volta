package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// DatabaseConfig contains the fields needed to connect to a database.
type DatabaseConfig struct {
	Driver   string `json:"driver"`
	Host     string `json:"host"`
	Port     int64  `json:"port"`
	Name     string `json:"name"`
	User     string `json:"user"`
	Password string `json:"password"`
	SSLMode  string `json:"ssl_mode"`
}

// Credentials with return a string of credentials appropriate for Go's
// sql.Open function
func (db DatabaseConfig) Credentials() string {
	// Only add the key if there is a value
	var values []string
	if db.Host != "" {
		values = append(values, fmt.Sprintf("host=%s", db.Host))
	}
	if db.Port != 0 {
		values = append(values, fmt.Sprintf("port=%d", db.Port))
	}
	if db.Name != "" {
		values = append(values, fmt.Sprintf("dbname=%s", db.Name))
	}
	if db.User != "" {
		values = append(values, fmt.Sprintf("user=%s", db.User))
	}
	if db.Password != "" {
		values = append(values, fmt.Sprintf("password=%s", db.Password))
	}
	if db.SSLMode != "" {
		values = append(values, fmt.Sprintf("sslmode=%s", db.SSLMode))
	}
	return strings.Join(values, " ")
}

// ParseDatabaseFile will create a DatabaseConfig using the given filepath.
func ParseDatabaseFile(filepath string) (c DatabaseConfig, err error) {
	f, err := os.Open(filepath)
	if err != nil {
		return
	}
	err = json.NewDecoder(f).Decode(&c)
	return
}
