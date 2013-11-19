package volta

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"net"
	"time"
)

type ServerConfig struct {
	Port         int
	IP           net.IP
	StaticPath   string
	TemplatePath string
	StaticURL    string
	LogPath      string
	TimeLayout   string
	WriteTimeout time.Duration
	ReadTimeout  time.Duration
}

// TODO Merge multiple configs
// TODO Is there an easy way to do this through reflect?
func (config ServerConfig) Merge(other ServerConfig) ServerConfig {
	if other.Port != 0 {
		config.Port = other.Port
	}
	if other.IP != nil {
		config.IP = other.IP
	}
	if other.StaticPath != "" {
		config.StaticPath = other.StaticPath
	}
	if other.TemplatePath != "" {
		config.TemplatePath = other.TemplatePath
	}
	if other.StaticURL != "" {
		config.StaticURL = other.StaticURL
	}
	if other.LogPath != "" {
		config.LogPath = other.LogPath
	}
	if other.TimeLayout != "" {
		config.TimeLayout = other.TimeLayout
	}
	if other.WriteTimeout != time.Duration(0) {
		config.WriteTimeout = other.WriteTimeout
	}
	if other.ReadTimeout != time.Duration(0) {
		config.ReadTimeout = other.ReadTimeout
	}
	return config
}

// The struct that will be used to parse JSON config files
type JSONConfig struct {
	Port         int    `json:"port"`
	IP           string `json:"ip"`
	StaticPath   string `json:"static-path"`
	TemplatePath string `json:"template-path"`
	StaticURL    string `json:"static-url"`
	LogPath      string `json:"log-path"`
	TimeLayout   string `json:"time-layout"`
	WriteTimeout string `json:"write-timeout"`
	ReadTimeout  string `json:"read-timeout"`
}

var Defaults = ServerConfig{
	Port:         80,
	IP:           net.ParseIP(""),
	StaticPath:   "",
	TemplatePath: "",
	StaticURL:    "/static/",
	LogPath:      "",
	TimeLayout:   time.UnixDate,
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

func ParseFlagConfig(base ServerConfig) (ServerConfig, error) {
	config := ServerConfig{}
	// Declare the flags that can be parsed directly into the configuration
	flag.IntVar(&config.Port, "port", 0, "The port for the server")
	flag.DurationVar(&config.WriteTimeout, "write-timeout", time.Duration(0), "The longest the server should take to write a response")
	flag.DurationVar(&config.ReadTimeout, "read-timeout", time.Duration(0), "The longest the server should take to read a response")
	flag.StringVar(&config.StaticPath, "static-path", "", "The path to the static files you wish to serve")
	flag.StringVar(&config.TemplatePath, "template-path", "", "The path to the template files you wish to parse")
	flag.StringVar(&config.LogPath, "log-path", "", "The location where the log file should be written")
	flag.StringVar(&config.StaticURL, "static-url", "", "The URL that should be used to serve static files")

	// Flags that will require additional processing or error checking
	var ip string
	var timeLayout string

	flag.StringVar(&ip, "ip", "", "IPv4 or IPv6 address where the server should run")
	flag.StringVar(&timeLayout, "time-layout", "", "The time layout that should used throughout the server")

	// Parse!
	flag.Parse()

	// Confirm that a valid IP address was given
	config.IP = net.ParseIP(ip)

	// TODO how to confirm that a time layout is valid?
	config.TimeLayout = timeLayout

	// Merge the flag config with the default base config
	return base.Merge(config), nil
}
