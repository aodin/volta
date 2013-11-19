package volta

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"time"
)

type ServerConfig struct {
	Port         int
	IP           net.IP
	StaticDir    string
	TemplateDir  string
	StaticURL    string
	LogPath      string
	TimeFormat   string
	WriteTimeout time.Duration
	ReadTimeout  time.Duration
}

// The struct that will be used to parse JSON config files
type JSONConfig struct {
	Port         int    `json:"port"`
	IP           string `json:"ip"`
	StaticDir    string `json:"static directory"`
	WriteTimeout string `json:"write timeout"`
	ReadTimeout  string `json:"read timeout"`
}

var Defaults = ServerConfig{
	Port:         80,
	IP:           net.ParseIP(""),
	StaticDir:    "./static/",
	TemplateDir:  "./templates",
	StaticURL:    "/static/",
	LogPath:      "",
	TimeFormat:   "",
	WriteTimeout: time.Duration(0),
	ReadTimeout:  time.Duration(0),
}

// TODO How to inject defaults into the parser? composition?
func ParseJSONConfig(path string, base ServerConfig) (ServerConfig, error) {
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return base, err
	}
	config := &JSONConfig{}
	err = json.Unmarshal(contents, config)

	// Individually check fields for now, parsing where necessary
	// TODO Strict mode, errors will be returned
	if config.Port != 0 {
		base.Port = config.Port
	}
	ip := net.ParseIP(config.IP)
	if ip != nil {
		base.IP = ip
	}
	if err != nil {
		return base, err
	}
	// If the parsed config did not contain a value, use the default values
	return base, nil
}
