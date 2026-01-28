package config

import "gopkg.in/yaml.v3"

// yamlUnmarshal unmarshals YAML data into a value.
func yamlUnmarshal(data []byte, v any) error {
	return yaml.Unmarshal(data, v)
}
