package resolve

import (
	"fmt"
	"os"
)

// Options controls expansion behavior.
type Options struct {
	Strict bool // Whether unresolved tokens should error.
}

// ExpandEnv replaces ${VAR} with environment values.
func ExpandEnv(input []byte, opts Options) ([]byte, error) {
	return Expand(input, os.LookupEnv, opts)
}

// Expand replaces ${VAR} tokens using a lookup function.
func Expand(input []byte, lookup func(string) (string, bool), opts Options) ([]byte, error) {
	if len(input) == 0 {
		return input, nil
	}

	out := make([]byte, 0, len(input))
	for i := 0; i < len(input); i++ {
		ch := input[i]
		// Only ${VAR} tokens are candidates for expansion.
		if !isTokenStart(input, i, ch) {
			out = append(out, ch)
			continue
		}

		// Find the closing brace for ${...}.
		start := i + 2 // Skip ${
		end, found := findTokenEnd(input, start)
		if !found {
			// Unterminated token; keep the rest unless strict.
			if opts.Strict {
				return nil, fmt.Errorf("unterminated ${...} token")
			}
			out = append(out, input[i:]...)
			break
		}

		name := string(input[start:end])
		if name == "" {
			// Empty token; keep as-is unless strict.
			if opts.Strict {
				return nil, fmt.Errorf("empty ${} token")
			}
			out = append(out, input[i:end+1]...)
			i = end
			continue
		}
		if !isValidName(name) {
			// Invalid name; keep as-is unless strict.
			if opts.Strict {
				return nil, fmt.Errorf("invalid env var name %q", name)
			}
			out = append(out, input[i:end+1]...)
			i = end
			continue
		}
		value, ok := lookup(name)
		if !ok {
			// Unknown env var; keep as-is unless strict.
			if opts.Strict {
				return nil, fmt.Errorf("env var %q is not set", name)
			}
			out = append(out, input[i:end+1]...)
			i = end
			continue
		}
		// Replace ${VAR} with the resolved value.
		out = append(out, value...)
		i = end
	}
	return out, nil
}

// isTokenStart returns whether the input at i is the start of a ${VAR} token.
func isTokenStart(input []byte, i int, ch byte) bool {
	return ch == '$' && i+1 < len(input) && input[i+1] == '{'
}

// findTokenEnd returns the index of the closing brace for a ${...} token.
func findTokenEnd(input []byte, start int) (int, bool) {
	for end := range len(input) {
		if end < start {
			continue
		}
		if input[end] == '}' {
			return end, true
		}
	}
	return 0, false
}

// isValidName returns whether name is a valid env var name.
func isValidName(name string) bool {
	for i := range len(name) {
		ch := name[i]
		if i == 0 { // First char is special.
			if !isAlpha(ch) && ch != '_' {
				return false
			}
			continue
		}
		if !isAlpha(ch) && !isDigit(ch) && ch != '_' {
			return false
		}
	}
	return true
}

// isAlpha returns whether ch is an ASCII letter.
func isAlpha(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

// isDigit returns whether ch is an ASCII digit.
func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}
