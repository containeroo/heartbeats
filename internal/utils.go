package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"text/template"

	"gopkg.in/yaml.v3"
)

// FormatOutput formats the output according to outputFormat
func FormatOutput(outputFormat string, textTemplate string, output interface{}) (string, error) {
	switch outputFormat {
	case "json":
		var b bytes.Buffer
		jsonEncoder := json.NewEncoder(&b)
		jsonEncoder.SetIndent("", "  ")
		if err := jsonEncoder.Encode(&output); err != nil {
			return "", err
		}
		return b.String(), nil

	case "yaml", "yml":
		var b bytes.Buffer
		yamlEncoder := yaml.NewEncoder(&b)
		yamlEncoder.SetIndent(2)
		if err := yamlEncoder.Encode(&output); err != nil {
			return "", err
		}
		return b.String(), nil

	case "txt", "text":
		txt, err := FormatTemplate(textTemplate, &output)
		if err != nil {
			return "", fmt.Errorf("Error formatting output")
		}
		return fmt.Sprintf("%+v", txt), nil

	default:
		return "", fmt.Errorf("Output format %s not supported", outputFormat)
	}
}

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
