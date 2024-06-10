package utils

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/Masterminds/sprig"
)

// FormatTemplate format template with intr as input
func FormatTemplate(name, tmpl string, intr any) (string, error) {
	if tmpl == "" {
		return "", fmt.Errorf("Template is empty")
	}

	fmap := sprig.TxtFuncMap()

	t, err := template.New(name).Funcs(fmap).Parse(tmpl)
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
