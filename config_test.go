package volta

import (
	"testing"
)

// TODO An assert function

// Parse a config that only contains a port and ip address
func TestSimpleJSONConfig(t *testing.T) {
	config, err := ParseJSONConfig("./_testfiles/simple_config.json", Defaults)
	if err != nil {
		t.Fatalf("Unexpected error thrown by JSON Config Parser: %s", err)
	}
	// These values should have replaced the defaults
	if config.Port != 8080 {
		t.Errorf("Unexpected port in config: %d != 8080", config.Port)
	}
	if config.IP.String() != `192.168.1.1` {
		t.Errorf("Unexpected ip in config: %s != 192.168.1.1", config.IP.String())
	}
	// These defaults should remain
	if config.StaticURL != `/static/` {
		t.Errorf("Unexpected static URL: %s != /static/", config.StaticURL)
	}
}
