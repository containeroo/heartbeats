package utils

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/Masterminds/sprig"
)

// FormatTemplate formats a template string with the provided data.
//
// Parameters:
//   - name: The name of the template.
//   - tmpl: The template string to format.
//   - intr: The data to be injected into the template.
//
// Returns:
//   - string: The formatted template.
//   - error: An error if the template formatting fails.
func FormatTemplate(name, tmpl string, intr interface{}) (string, error) {
	if tmpl == "" {
		return "", fmt.Errorf("template is empty")
	}

	fmap := sprig.TxtFuncMap()

	t, err := template.New(name).Funcs(fmap).Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("error parsing template. %w", err)
	}

	buf := &bytes.Buffer{}
	if err := t.Execute(buf, &intr); err != nil {
		return "", fmt.Errorf("error executing template. %w", err)
	}

	return buf.String(), nil
}
