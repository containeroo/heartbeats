package envflag

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/pflag"
)

// resolve returns a parsed value from flag, env, or flag default.
func resolve[T any](fs *pflag.FlagSet, name, envKey string, parse func(string) (T, error), getEnv func(string) string) (T, error) {
	f := fs.Lookup(name)
	if f == nil {
		var zero T
		return zero, fmt.Errorf("flag not found: %s", name)
	}
	if f.Changed {
		return parse(f.Value.String())
	}
	if val := getEnv(envKey); val != "" {
		return parse(val)
	}
	return parse(f.DefValue)
}

// String returns a trimmed string from flag/env/default.
func String(fs *pflag.FlagSet, name, envKey string) (string, error) {
	return resolve(fs, name, envKey, func(s string) (string, error) { return strings.TrimSpace(s), nil }, os.Getenv)
}

// Int returns an int from flag/env/default.
func Int(fs *pflag.FlagSet, name, envKey string) (int, error) {
	return resolve(fs, name, envKey, strconv.Atoi, os.Getenv)
}

// Bool returns a bool from flag/env/default.
func Bool(fs *pflag.FlagSet, name, envKey string) (bool, error) {
	return resolve(fs, name, envKey, strconv.ParseBool, os.Getenv)
}

// Duration returns a duration from flag/env/default.
func Duration(fs *pflag.FlagSet, name, envKey string) (time.Duration, error) {
	return resolve(fs, name, envKey, time.ParseDuration, os.Getenv)
}

// HostPort parses host:port from flag/env/default.
func HostPort(fs *pflag.FlagSet, name, envKey string) (string, error) {
	return resolve(fs, name, envKey, func(s string) (string, error) {
		if _, _, err := net.SplitHostPort(s); err != nil {
			return "", fmt.Errorf("invalid format: %w", err)
		}
		return s, nil
	}, os.Getenv)
}

// SchemaHostPort parses schema://host:port from flag/env/default.
func SchemaHostPort(fs *pflag.FlagSet, name, envKey string) (string, error) {
	return resolve(fs, name, envKey, func(s string) (string, error) {
		u, err := url.Parse(s)
		if err != nil {
			return "", fmt.Errorf("invalid format: %w", err)
		}
		if u.Scheme == "" {
			return "", fmt.Errorf("missing protocol scheme")
		}
		host, port, err := net.SplitHostPort(u.Host)
		if err != nil {
			return "", fmt.Errorf("invalid format: %w", err)
		}
		if host == "" {
			return "", fmt.Errorf("missing host in address")
		}
		if port == "" {
			return "", fmt.Errorf("missing port in address")
		}
		return s, nil
	}, os.Getenv)
}

// URL validates a full URL with scheme and hostname.
func URL(fs *pflag.FlagSet, name, envKey string) (string, error) {
	return resolve(fs, name, envKey, func(s string) (string, error) {
		if !strings.Contains(s, "://") {
			return "", fmt.Errorf("missing protocol scheme")
		}
		u, err := url.Parse(s)
		if err != nil {
			return "", fmt.Errorf("invalid URL: %w", err)
		}
		if u.Hostname() == "" {
			return "", fmt.Errorf("missing host in address")
		}
		return s, nil
	}, os.Getenv)
}
