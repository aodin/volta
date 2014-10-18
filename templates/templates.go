package templates

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Templates holds the parsed template files and any local attributes.
type Templates struct {
	parsed *template.Template
	locals Attrs
}

// SetAttr sets a local templates variable. Overwritten variables return an
// error, which can be ignored as needed.
func (t *Templates) SetAttr(key string, value interface{}) (err error) {
	if _, exists := t.locals[key]; exists {
		err = fmt.Errorf("templates: overwriting key %s in attrs", key)
	}
	t.locals[key] = value
	return
}

// Execute will render the given template define name with the given attrs.
// Panics on error, because that's how we roll.
func (t *Templates) Execute(w http.ResponseWriter, n string, attrs ...Attrs) {
	data := Attrs{}
	data.Merge(t.locals)
	for _, attr := range attrs {
		data.Merge(attr)
	}
	if err := t.parsed.ExecuteTemplate(w, n, data); err != nil {
		panic(fmt.Sprintf("templates: could execute template %s: %s", n, err))
	}
}

// New creates a new Templates instance by compiling all files recursively in
// the given directory. Panics on error.
func New(p string, attrs ...Attrs) *Templates {
	data := Attrs{}
	for _, attr := range attrs {
		data.Merge(attr)
	}
	t := &Templates{
		parsed: template.New("template"),
		locals: data,
	}

	// Recursively walk the given directory and build templates, returning
	// immediately on error.
	err := filepath.Walk(p, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(strings.ToLower(path), ".html") {
			if _, err := t.parsed.ParseFiles(path); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		panic(fmt.Sprintf("templates: error while parsing templates: %s", err))
	}
	return t
}
