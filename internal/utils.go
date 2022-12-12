package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"text/template"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// WriteOutput writes the output to the response writer
//
// - w is the response writer
// - statusCode is the HTTP status code
// - outputFormat is the format of the output (json, yaml, yml, txt, text)
// - output is the data to render the template with
// - textTmpl is the template to use for the text output
func WriteOutput(w http.ResponseWriter, statusCode int, outputFormat string, output interface{}, textTemplate string) {
	formatOutput, err := FormatOutput(outputFormat, textTemplate, output)
	if err != nil {
		w.WriteHeader(statusCode)
		_, err = w.Write([]byte(err.Error()))
		if err != nil {
			log.Errorf("Cannot write response: %s", err)
		}
		return
	}
	w.WriteHeader(statusCode)
	_, err = w.Write([]byte(formatOutput))
	if err != nil {
		log.Errorf("Cannot write response: %s", err)
	}
	log.Tracef("Server respond with: %d %s", statusCode, formatOutput)
}

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
			return "", fmt.Errorf("Error formatting output. %s", err.Error())
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

// CheckOutput checks if the output format is supported
func GetOutputFormat(req *http.Request) string {
	outputFormat := req.URL.Query().Get("output")
	if outputFormat == "" {
		outputFormat = "txt"
	}
	return outputFormat
}

// CheckDefault checks if value is empty and returns default value
func CheckDefault(value string, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

func IsInListOfStrings(list []string, value string) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}

func ParseTemplates(templates []string, w http.ResponseWriter) {
	tmpl, err := template.ParseFS(StaticFs, templates...)
	if err != nil {
		log.Errorf("Error parsing template: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("cannot parse template. %s", err.Error())))
		return
	}

	if err := tmpl.ExecuteTemplate(w, "base", &Documentation); err != nil {
		log.Errorf("Error executing template: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("cannot execute template. %s", err.Error())))
		return
	}
}
