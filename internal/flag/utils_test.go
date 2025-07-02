package flag

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	flag "github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func TestMust(t *testing.T) {
	t.Parallel()

	t.Run("returns value if no error", func(t *testing.T) {
		t.Parallel()

		got := must("value", nil)
		assert.Equal(t, "value", got)
	})

	t.Run("panics if error is not nil", func(t *testing.T) {
		t.Parallel()

		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic but got none")
			}
		}()
		_ = must("fail", errors.New("error"))
	})
}

func TestDecorateUsageWithEnv(t *testing.T) {
	t.Run("adds env suffix to flag usage", func(t *testing.T) {
		fs := flag.NewFlagSet("test", flag.ContinueOnError)
		fs.String("my-flag", "", "some description")

		decorateUsageWithEnv(fs, "CASCADER")

		f := fs.Lookup("my-flag")
		if !strings.Contains(f.Usage, "(env: CASCADER_MY_FLAG)") {
			t.Errorf("expected usage to contain env suffix, got: %q", f.Usage)
		}
	})

	t.Run("does not overwrite existing env hint", func(t *testing.T) {
		fs := flag.NewFlagSet("test", flag.ContinueOnError)
		fs.String("already", "", "example (env: CUSTOM)")

		decorateUsageWithEnv(fs, "CASCADER")

		f := fs.Lookup("already")
		if strings.Count(f.Usage, "(env:") > 1 {
			t.Errorf("expected only one env hint, got: %q", f.Usage)
		}
	})
}

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
