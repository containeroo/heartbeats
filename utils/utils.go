package utils

import (
	"bytes"
	"encoding/json"
	"net"
	"strconv"

	"gopkg.in/yaml.v3"
)

// FormatOutput maps each possible output to its corresponding converting function
var FormatOutput = map[string]func(interface{}) (string, error){
	"json": ConvertToJson,
	"yaml": ConvertToYaml,
	"yml":  ConvertToYaml,
	"text": ConvertToYaml,
	"txt":  ConvertToYaml,
}

// convertToYaml converts the output to yaml
func ConvertToYaml(output interface{}) (string, error) {
	var b bytes.Buffer
	yamlEncoder := yaml.NewEncoder(&b)
	yamlEncoder.SetIndent(2)
	if err := yamlEncoder.Encode(output); err != nil {
		return "", err
	}
	return b.String(), nil
}

func ConvertToJson(output interface{}) (string, error) {
	var b bytes.Buffer
	jsonEncoder := json.NewEncoder(&b)
	jsonEncoder.SetIndent("", "  ")
	if err := jsonEncoder.Encode(&output); err != nil {
		return "", err
	}
	return b.String(), nil
}

// ExtractHostAndPort extracts the hostname and port from the listen address
func ExtractHostAndPort(address string) (string, int, error) {
	host, portStr, err := net.SplitHostPort(address)
	if err != nil {
		return "", 0, err
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return "", 0, err
	}

	return host, port, nil
}
