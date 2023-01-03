package utils

import (
	"bytes"
	"fmt"
	"text/template"
)

// FormatTemplate format template with intr as input
func FormatTemplate(tmpl string, intr any) (string, error) {
	if tmpl == "" {
		return "", fmt.Errorf("Template is empty")
	}
	t, err := template.New("status").Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("Error executing template: %s", err.Error())
	}
	buf := &bytes.Buffer{}
	err = t.Execute(buf, &intr)
	if err != nil {
		return "", fmt.Errorf("Error executing template: %s", err.Error())
	}

	return buf.String(), nil
}

// CheckDefault checks if value is empty and returns default value
func CheckDefault(value string, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

// IsInListOfStrings checks if value is in list
func IsInListOfStrings(list []string, value string) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}
