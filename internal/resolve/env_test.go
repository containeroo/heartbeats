package resolve

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExpand(t *testing.T) {
	t.Parallel()

	lookup := func(key string) (string, bool) {
		switch key {
		case "HOST":
			return "example.com", true
		case "PORT":
			return "8080", true
		default:
			return "", false
		}
	}

	out, err := Expand([]byte("url: http://${HOST}:${PORT}\n"), lookup, Options{})
	assert.NoError(t, err)
	assert.Equal(t, "url: http://example.com:8080\n", string(out))
}

func TestExpandErrors(t *testing.T) {
	t.Parallel()

	lookup := func(key string) (string, bool) {
		return "", false
	}

	out, err := Expand([]byte("missing: ${MISSING}\n"), lookup, Options{})
	assert.NoError(t, err)
	assert.Equal(t, "missing: ${MISSING}\n", string(out))
}

func TestExpandStrictErrors(t *testing.T) {
	t.Parallel()

	lookup := func(key string) (string, bool) {
		return "", false
	}

	_, err := Expand([]byte("missing: ${MISSING}\n"), lookup, Options{Strict: true})
	assert.Error(t, err)
}
