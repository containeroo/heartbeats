package flag

import (
	"errors"
	"fmt"
	"io"
)

// must panics on err and is used to keep config assembly clean.
func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

// envDesc appends an environment variable hint to a flag description.
func envDesc(desc, envVar string) string {
	return fmt.Sprintf("%s (env: %s)", desc, envVar)
}

// IsHelpRequested checks if the error is a HelpRequested sentinel and prints it.
func IsHelpRequested(err error, w io.Writer) bool {
	var helpErr *HelpRequested
	if errors.As(err, &helpErr) {
		fmt.Fprint(w, helpErr.Error()) // nolint:errcheck
		return true
	}
	return false
}
