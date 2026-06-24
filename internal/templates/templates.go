package templates

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/containeroo/tmplfuncs"
)

// Template wraps a parsed text template loaded from disk.
type Template struct {
	tmpl *template.Template
}

// StringTemplate wraps a parsed text template for string output.
type StringTemplate struct {
	tmpl *template.Template
}

// Load reads and parses a template file.
func Load(path string) (*Template, error) {
	if path == "" {
		return nil, errors.New("template path is required")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read template: %w", err)
	}
	parsed, err := parseTemplate(string(data))
	if err != nil {
		return nil, fmt.Errorf("parse template: %w", err)
	}
	return &Template{tmpl: parsed}, nil
}

// LoadDefault parses the embedded heartbeat template.
func LoadDefault(tmplFS fs.FS) (*Template, error) {
	return LoadFromFS(tmplFS, "templates/default.tmpl")
}

// LoadFromFS reads and parses a template from an fs.FS.
func LoadFromFS(tmplFS fs.FS, path string) (*Template, error) {
	if path == "" {
		return nil, errors.New("template path is required")
	}
	file, err := tmplFS.Open(path)
	if err != nil {
		return nil, fmt.Errorf("read template: %w", err)
	}
	defer file.Close() // nolint:errcheck

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("read template: %w", err)
	}
	parsed, err := parseTemplate(string(data))
	if err != nil {
		return nil, fmt.Errorf("parse embedded template: %w", err)
	}
	return &Template{tmpl: parsed}, nil
}

// FuncMap returns a set of custom template functions for use in notifications.
func FuncMap() template.FuncMap {
	funcs := tmplfuncs.FuncMap()

	// Backward-compatible aliases used by existing heartbeat templates.
	funcs["toUpper"] = tmplfuncs.UpperValue
	funcs["toLower"] = tmplfuncs.LowerValue
	funcs["ensurePrefix"] = tmplfuncs.WithPrefixValue

	// Keep helpers with project-specific behavior local.
	funcs["formatTime"] = formatTime
	funcs["formatDuration"] = formatDuration
	funcs["isRecent"] = isRecent
	funcs["ago"] = ago
	funcs["join"] = strings.Join

	return funcs
}

// parseTemplate parses a notification template with the shared FuncMap.
func parseTemplate(input string) (*template.Template, error) {
	return template.New("heartbeat").Option("missingkey=error").Funcs(FuncMap()).Parse(input)
}

// ParseStringTemplate parses a string template with missingkey=error.
func ParseStringTemplate(name, input string) (*StringTemplate, error) {
	if input == "" {
		return nil, errors.New("template input is required")
	}
	parsed, err := template.New(name).Option("missingkey=error").Funcs(FuncMap()).Parse(input)
	if err != nil {
		return nil, fmt.Errorf("parse template: %w", err)
	}
	return &StringTemplate{tmpl: parsed}, nil
}

// LoadString reads a file and parses it into a StringTemplate.
func LoadString(path string) (*StringTemplate, error) {
	if path == "" {
		return nil, errors.New("template path is required")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read template: %w", err)
	}
	return ParseStringTemplate("string", string(data))
}

// LoadStringFromFS reads a file from an fs.FS into a StringTemplate.
func LoadStringFromFS(tmplFS fs.FS, path string) (*StringTemplate, error) {
	if path == "" {
		return nil, errors.New("template path is required")
	}
	file, err := tmplFS.Open(path)
	if err != nil {
		return nil, fmt.Errorf("read template: %w", err)
	}
	defer file.Close() // nolint:errcheck

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("read template: %w", err)
	}
	return ParseStringTemplate("string", string(data))
}

// Render executes the template with the provided data and returns a string.
func (t *StringTemplate) Render(data any) (string, error) {
	var buf bytes.Buffer
	if err := t.tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}
	return buf.String(), nil
}

// Render executes the template with the provided data and returns the bytes.
func (t *Template) Render(data any) ([]byte, error) {
	var buf bytes.Buffer
	if err := t.tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("execute template: %w", err)
	}
	return buf.Bytes(), nil
}

// formatTime formats a time value with a Go layout.
//
// The argument order preserves the existing heartbeat template API:
//
//	{{ formatTime .Time "2006-01-02 15:04:05 MST" }}
func formatTime(value any, layout string) (string, error) {
	return tmplfuncs.FormatTimeValue(layout, value)
}

// isRecent reports whether t happened less than two seconds ago.
func isRecent(t time.Time) bool {
	return time.Since(t).Truncate(time.Second).Seconds() < 2
}

// formatDuration renders a duration with heartbeat-specific rounding behavior.
func formatDuration(d time.Duration) string {
	if d < 0 {
		d = -d
	}
	if d < time.Second {
		return d.Truncate(time.Millisecond).String()
	}
	return d.Truncate(time.Second).String()
}

// ago renders the duration since t.
//
// A zero time is rendered as never.
func ago(t time.Time) string {
	if t.IsZero() {
		return "never"
	}
	return time.Since(t).Truncate(time.Second).String()
}
