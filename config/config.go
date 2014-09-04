package config

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

// Config is the parent configuration struct and includes fields for single
// configurations of a database, cookie, and SMTP connection.
type Config struct {
	HTTPS       bool           `json:"https"`
	Domain      string         `json:"domain"`
	Port        int            `json:"port"`
	TemplateDir string         `json:"templates"`
	StaticDir   string         `json:"static"`
	SecretKey   string         `json:"secret_key"`
	Database    DatabaseConfig `json:"database"`
	Cookie      CookieConfig   `json:"cookie"`
	SMTP        SMTPConfig     `json:"smtp"`
}

// Address returns the domain:port pair.
func (c Config) Address() string {
	return fmt.Sprintf("%s:%d", c.Domain, c.Port)
}

// Parse will create a Config using the file settings.json in the
// current directory.
func Parse() (Config, error) {
	return ParseFile("./settings.json")
}

// ParseFile will create a Config using the file at the given path.
func ParseFile(filename string) (Config, error) {
	f, err := os.Open(filename)
	if err != nil {
		return Config{}, err
	}
	return parse(f)
}

// TODO What about default values? Leave to user?
func parse(f io.Reader) (Config, error) {
	var c Config
	contents, err := ioutil.ReadAll(f)
	if err != nil {
		return c, err
	}
	if err = json.Unmarshal(contents, &c); err != nil {
		return c, err
	}
	return c, nil
}

func DefaultConfig(key string) Config {
	return Config{
		Cookie:    DefaultCookie,
		Port:      8080,
		SecretKey: key,
	}
}
