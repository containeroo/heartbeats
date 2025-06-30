package flag

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsHelpRequested(t *testing.T) {
	t.Parallel()

	t.Run("returns true and writes message for HelpRequested error", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		helpMsg := "this is the help message\n"
		err := &HelpRequested{Message: helpMsg}

		ok := IsHelpRequested(err, buf)

		assert.True(t, ok)
		assert.Equal(t, helpMsg, buf.String())
	})

	t.Run("returns false and writes nothing for unrelated error", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		err := errors.New("some other error")

		ok := IsHelpRequested(err, buf)

		assert.False(t, ok)
		assert.Equal(t, "", buf.String())
	})
}
