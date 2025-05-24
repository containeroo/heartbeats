package notifier

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
	"time"
)

// FuncMap returns a set of custom template functions for use in notifications.
func FuncMap() template.FuncMap {
	return template.FuncMap{
		"upper":      strings.ToUpper,
		"lower":      strings.ToLower,
		"formatTime": func(t time.Time, format string) string { return t.Format(format) },
		"ago": func(t time.Time) string {
			if t.IsZero() {
				return "never"
			}
			return time.Since(t).Truncate(time.Second).String()
		},
		"join": func(elems []string, sep string) string { return strings.Join(elems, sep) },
	}
}

// applyTemplate renders a template string using the provided data and FuncMap.
func applyTemplate(tmplStr string, data any) (string, error) {
	tmpl, err := template.New("notification").Funcs(FuncMap()).Parse(tmplStr)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// formatNotification renders title and message templates with fallbacks.
func formatNotification(data NotificationData, titleTmpl, textTmpl, defaultTitle, defaultText string) (NotificationData, error) {
	if titleTmpl == "" {
		titleTmpl = defaultTitle
	}
	if textTmpl == "" {
		textTmpl = defaultText
	}

	title, err := applyTemplate(titleTmpl, data)
	if err != nil {
		return data, fmt.Errorf("format title: %w", err)
	}
	text, err := applyTemplate(textTmpl, data)
	if err != nil {
		return data, fmt.Errorf("format message: %w", err)
	}

	data.Title = title
	data.Message = text
	return data, nil
}

// resolveSkipTLS returns the effective TLS setting, prioritizing an explicit value.
func resolveSkipTLS(explicit *bool, fallback bool) bool {
	if explicit != nil {
		return *explicit
	}
	return fallback
}
