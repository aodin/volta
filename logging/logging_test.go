package logging

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type testServer struct {
	views []View
}

func (s *testServer) logRequest(w http.ResponseWriter, r *http.Request) {
	s.views = append(s.views, LogRequest(r))
}

func expectString(t *testing.T, a, b string) {
	if a != b {
		t.Errorf("Unexpected string: %s != %s", a, b)
	}
}

func expectView(t *testing.T, a, b View) {
	if a.URI != b.URI {
		t.Errorf("Unexpected URI: %s != %s", a.URI, b.URI)
	}
	if a.IP != b.IP {
		t.Errorf("Unexpected IP: %s != %s", a.IP, b.IP)
	}
	if a.Agent != b.Agent {
		t.Errorf("Unexpected Agent: %s != %s", a.Agent, b.Agent)
	}
	if a.Referer != b.Referer {
		t.Errorf("Unexpected Referer: %s != %s", a.Referer, b.Referer)
	}
}

func TestLogRequest(t *testing.T) {
	ts := &testServer{views: make([]View, 0)}
	server := httptest.NewServer(http.HandlerFunc(ts.logRequest))
	defer server.Close()

	// Hardcode the user agent in case the default GO user agent changes
	client := &http.Client{}
	request, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("User-Agent", "Go")

	// Send the request
	_, err = client.Do(request)
	if err != nil {
		t.Fatal(err)
	}

	if len(ts.views) != 1 {
		t.Fatalf("Unexpected length of test server views:", len(ts.views))
	}

	expected := View{
		URI:     "GET /",
		IP:      "127.0.0.1",
		Agent:   "Go",
		Referer: "",
	}

	expectView(t, ts.views[0], expected)
	expectString(t, ts.views[0].String(), expected.String())

	// Clear the existing views
	ts.views = make([]View, 0)

	// Test the escaping of the URL, Agent, and Referer
	request, err = http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("User-Agent", `"HACKER"`)
	request.Header.Set("Referer", `"ESCAPE"`)

	_, err = client.Do(request)
	if err != nil {
		t.Fatal(err)
	}

	if len(ts.views) != 1 {
		t.Fatalf("Unexpected length of test server views:", len(ts.views))
	}

	expectedString := `"GET /" 127.0.0.1 "\"HACKER\"" "\"ESCAPE\""`
	expectString(t, ts.views[0].String(), expectedString)
}
