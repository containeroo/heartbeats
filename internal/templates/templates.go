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

	"github.com/containeroo/heartbeats/internal/utils"
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
	return template.FuncMap{
		"toUpper":    strings.ToUpper,
		"toLower":    strings.ToLower,
		"formatTime": func(t time.Time, format string) string { return t.Format(format) },
		"isRecent":   func(t time.Time) bool { return time.Since(t).Truncate(time.Second).Seconds() < 2 },
		"formatDuration": func(d time.Duration) string {
			if d < 0 {
				d = -d
			}
			if d < time.Second {
				return d.Truncate(time.Millisecond).String()
			}
			return d.Truncate(time.Second).String()
		},
		"default": func(fallback any, value any) any {
			if utils.IsZeroValue(value) {
				return fallback
			}
			return value
		},
		"coalesce": func(values ...any) any {
			for _, v := range values {
				if !utils.IsZeroValue(v) {
					return v
				}
			}
			return nil
		},
		"ensurePrefix": func(prefix, value string) string {
			if value == "" {
				return value
			}
			if strings.HasPrefix(value, prefix) {
				return value
			}
			return prefix + value
		},
		"ago": func(t time.Time) string {
			if t.IsZero() {
				return "never"
			}
			return time.Since(t).Truncate(time.Second).String()
		},
		"join": func(elems []string, sep string) string { return strings.Join(elems, sep) },
	}
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
