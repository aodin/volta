package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

type GooseConfig map[string]struct {
	Driver string
	Open   string
}

func ParseTestYAML(path string) (c DatabaseConfig, err error) {
	return parseGooseYAML(path, "test")
}

func parseGooseYAML(path, name string) (c DatabaseConfig, err error) {
	f, err := os.Open(path)
	if err != nil {
		return
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return
	}

	var goose GooseConfig
	if err = yaml.Unmarshal(b, &goose); err != nil {
		return
	}

	db, ok := goose[name]
	if !ok {
		err = fmt.Errorf("config: no database named '%s'", name)
		return
	}
	c.Driver = db.Driver

	// Split the open string
	attrs := strings.Split(db.Open, " ")

	// Where's my dynamic programming?
	m := make(map[string]string)
	for _, attr := range attrs {
		parts := strings.SplitN(attr, "=", 2)
		// Valid attrs will always have 2 parts
		if len(parts) != 2 {
			continue
		}
		m[parts[0]] = parts[1]
	}

	// Yup
	c.Host = m["host"]
	if c.Port, err = strconv.ParseInt(m["port"], 10, 64); err != nil {
		return
	}
	c.Name = m["dbname"]
	c.User = m["user"]
	c.Password = m["password"]
	c.SSLMode = m["sslmode"]
	return
}
