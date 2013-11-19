package volta

import (
	"fmt"
)

// TODO Extend an http.Server?
type Server struct {
	address  string
	Settings ServerConfig
}

func (web *Server) Address() string {
	if web.address != "" {
		return web.address
	}
	var ip string
	if web.Settings.IP != nil {
		ip = web.Settings.IP.String()
	}
	web.address = fmt.Sprintf("%s:%d", ip, web.Settings.Port)
	return web.address
}

// Set the server configuration using flags only
func (web *Server) FlagConfig() error {
	config, err := ParseFlagConfig(Defaults)
	if err != nil {
		return err
	}
	web.Settings = config
	return nil
}

// Set the server configuration using the JSON config and flags, with
// preference given to the flags
func (web *Server) Config(path string) error {
	// Get the JSON configuration
	JSONConfig, err := ParseJSONConfig(path, Defaults)
	if err != nil {
		return err
	}
	// Get the flag configuration
	flagConfig, err := ParseFlagConfig(Defaults)
	if err != nil {
		return err
	}
	// Merge to two configurations, defering to the flag-based config
	config := JSONConfig.Merge(flagConfig)
	web.Settings = config
	return nil
}
