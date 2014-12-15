package email

import (
	"bytes"
	"text/template"
)

var msg = `From: {{ .From }}
To: {{ .To }}
Subject: {{ .Subject }}
{{ range $key, $value := .Header }}{{ $key }}: {{ $value }}
{{ end }}
{{ .Body }}
`
var emailTemplate = template.Must(template.New("email").Parse(msg))

// Email is the structure of an email
type Email struct {
	From    string
	To      string
	Subject string
	Header  map[string]string
	Body    string
}

// String returns the string ready to be passed to a Sender.Send()
func (email Email) String() string {
	// TODO Prevent email lines from being over 78 characters?
	b := new(bytes.Buffer)

	// TODO The following error is ignored
	emailTemplate.Execute(b, email)
	return b.String()
}
