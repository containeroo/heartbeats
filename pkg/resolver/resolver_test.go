package resolver

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveVariable(t *testing.T) {
	t.Run("Resolve environment variable", func(t *testing.T) {
		os.Setenv("TEST_ENV_VAR", "test_value")
		defer os.Unsetenv("TEST_ENV_VAR")

		result, err := ResolveVariable("env:TEST_ENV_VAR")
		assert.NoError(t, err)
		assert.Equal(t, "test_value", result)
	})

	t.Run("Resolve non-existing environment variable", func(t *testing.T) {
		result, err := ResolveVariable("env:NON_EXISTING_ENV_VAR")
		assert.Error(t, err)
		assert.Equal(t, "", result)
		assert.Contains(t, err.Error(), "environment variable 'NON_EXISTING_ENV_VAR' not found")
	})

	t.Run("Resolve file variable", func(t *testing.T) {
		fileContent := "key1=value1\nkey2=value2\n"
		file, err := os.CreateTemp("", "testfile")
		assert.NoError(t, err)
		defer os.Remove(file.Name())

		_, err = file.WriteString(fileContent)
		assert.NoError(t, err)
		file.Close()

		result, err := ResolveVariable("file:" + file.Name())
		assert.NoError(t, err)
		assert.Equal(t, fileContent, result+"\n")
	})

	t.Run("Resolve file with key", func(t *testing.T) {
		fileContent := "key1=value1\nkey2=value2\n"
		file, err := os.CreateTemp("", "testfile")
		assert.NoError(t, err)
		defer os.Remove(file.Name())

		_, err = file.WriteString(fileContent)
		assert.NoError(t, err)
		file.Close()

		result, err := ResolveVariable("file:" + file.Name() + "//key1")
		assert.NoError(t, err)
		assert.Equal(t, "value1", result)
	})

	t.Run("Resolve file with non-existing key", func(t *testing.T) {
		fileContent := "key1=value1\nkey2=value2\n"
		file, err := os.CreateTemp("", "testfile")
		assert.NoError(t, err)
		defer os.Remove(file.Name())

		_, err = file.WriteString(fileContent)
		assert.NoError(t, err)
		file.Close()

		result, err := ResolveVariable("file:" + file.Name() + "//non_existing_key")
		assert.Error(t, err)
		assert.Equal(t, "", result)
		assert.Contains(t, err.Error(), "Key 'non_existing_key' not found in file")
	})

	t.Run("Resolve plain string", func(t *testing.T) {
		result, err := ResolveVariable("plain_string")
		assert.NoError(t, err)
		assert.Equal(t, "plain_string", result)
	})
}
