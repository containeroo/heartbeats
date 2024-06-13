package resolver

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// Constants for variable resolution.
const (
	envPrefix  = "env:"  // Prefix to identify environment variable references
	filePrefix = "file:" // Prefix to identify file references
	keyDelim   = "//"    // Delimiter to identify a key in a file
)

// ResolveVariable resolves a string value based on its prefix.
//
// - "env:": Treated as an environment variable and resolved accordingly.
// - "file:": Treated as a file path, optionally followed by a key to retrieve a specific line in "key = value" format.
// - No prefix: The string is returned as is.
//
// Parameters:
//   - value: The string to resolve.
//
// Returns:
//   - The resolved value of the input string.
//   - An error if the resolution fails.
func ResolveVariable(value string) (string, error) {
	if strings.HasPrefix(value, envPrefix) {
		return resolveEnvVariable(value[len(envPrefix):])
	}

	if strings.HasPrefix(value, filePrefix) {
		return resolveFileVariable(value[len(filePrefix):])
	}

	return value, nil
}

// resolveEnvVariable resolves an environment variable.
//
// Parameters:
//   - envVar: The name of the environment variable to resolve.
//
// Returns:
//   - The value of the environment variable.
//   - An error if the environment variable is not found.
func resolveEnvVariable(envVar string) (string, error) {
	resolvedVariable, found := os.LookupEnv(envVar)
	if !found {
		return "", fmt.Errorf("environment variable '%s' not found.", envVar)
	}
	return resolvedVariable, nil
}

// resolveFileVariable resolves a file path with an optional key.
// The key should be in the format "key = value".
//
// Parameters:
//   - filePathWithKey: The string containing the file path and optional key.
//
// Returns:
//   - The resolved value based on the file and optional key.
//   - An error if resolving the file or key fails.
func resolveFileVariable(filePathWithKey string) (string, error) {
	lastSeparatorIndex := strings.LastIndex(filePathWithKey, keyDelim)
	filePath := filePathWithKey // default filePath (whole value)
	key := ""                   // default key (no key)

	// Check for key specification
	if lastSeparatorIndex != -1 {
		filePath = filePathWithKey[:lastSeparatorIndex]
		key = filePathWithKey[lastSeparatorIndex+len(keyDelim):]
	}

	filePath = os.ExpandEnv(filePath)
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("Failed to open file '%s'. %v", filePath, err)
	}
	defer file.Close()

	if key != "" {
		return searchKeyInFile(file, key)
	}

	// No key specified, read the whole file
	data, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("Failed to read file '%s'. %v", filePath, err)
	}
	return strings.TrimSpace(string(data)), nil
}

// searchKeyInFile searches for a specified key in a file and returns its associated value.
// The key should be in the format "key = value".
//
// Parameters:
//   - file: The opened file to search for the key.
//   - key: The key to search for in the file.
//
// Returns:
//   - The value associated with the specified key.
//   - An error if the key is not found in the file.
func searchKeyInFile(file *os.File, key string) (string, error) {
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		pair := strings.SplitN(line, "=", 2)
		if len(pair) == 2 && strings.TrimSpace(pair[0]) == key {
			return strings.TrimSpace(pair[1]), nil
		}
	}
	return "", fmt.Errorf("Key '%s' not found in file '%s'.", key, file.Name())
}
