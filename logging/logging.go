package logging

import (
	"fmt"
	"net/http"
	"strings"
)

type View struct {
	URI     string
	IP      string
	Agent   string
	Referer string
}

func (v View) String() string {
	return fmt.Sprintf(`%q %s %q %q`, v.URI, v.IP, v.Agent, v.Referer)
}

func (v View) Strings() []string {
	return []string{v.URI, v.IP, v.Agent, v.Referer}
}

func LogRequest(request *http.Request) View {
	ip := strings.SplitN(request.Header.Get("X-Real-IP"), ":", 2)[0]
	if ip == "" {
		ip = strings.SplitN(request.RemoteAddr, ":", 2)[0]
	}
	return View{
		URI:     fmt.Sprintf(`%s %s`, request.Method, request.URL),
		IP:      ip,
		Agent:   request.Header.Get("User-Agent"),
		Referer: request.Header.Get("Referer"),
	}
}
