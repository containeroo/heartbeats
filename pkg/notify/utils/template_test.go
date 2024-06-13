package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatTemplate(t *testing.T) {
	t.Run("Empty template", func(t *testing.T) {
		result, err := FormatTemplate("empty", "", nil)
		assert.Error(t, err)
		assert.Equal(t, "", result)
		assert.Equal(t, "template is empty", err.Error())
	})

	t.Run("Simple template", func(t *testing.T) {
		tmpl := "Hello, {{ .Name }}!"
		data := map[string]string{"Name": "World"}

		result, err := FormatTemplate("simple", tmpl, data)
		assert.NoError(t, err)
		assert.Equal(t, "Hello, World!", result)
	})

	t.Run("Template with sprig function", func(t *testing.T) {
		tmpl := "The date is {{ now | date \"2006-01-02\" }}."
		data := map[string]string{}

		result, err := FormatTemplate("sprig", tmpl, data)
		assert.NoError(t, err)
		assert.Contains(t, result, "The date is ")
	})

	t.Run("Template with missing field", func(t *testing.T) {
		tmpl := "Hello, {{ .Name | default \"\" }}!"
		data := struct{ Name *string }{nil}

		result, err := FormatTemplate("missing", tmpl, data)
		assert.NoError(t, err)
		assert.Equal(t, "Hello, !", result)
	})

	t.Run("Complex template with multiple fields", func(t *testing.T) {
		tmpl := "Hello, {{ .Name }}! Today is {{ .Day }}."
		data := map[string]string{"Name": "Alice", "Day": "Monday"}

		result, err := FormatTemplate("complex", tmpl, data)
		assert.NoError(t, err)
		assert.Equal(t, "Hello, Alice! Today is Monday.", result)
	})
}
