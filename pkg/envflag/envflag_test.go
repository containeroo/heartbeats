package envflag

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func TestResolve(t *testing.T) {
	parseString := func(s string) (string, error) {
		if s == "bad" {
			return "", errors.New("parse error")
		}
		return s, nil
	}

	t.Run("flag value set", func(t *testing.T) {
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		fs.String("key", "default", "")
		_ = fs.Set("key", "flagval")

		val, err := resolve(fs, "key", "ENV_KEY", parseString, os.Getenv)
		assert.NoError(t, err)
		assert.Equal(t, "flagval", val)
	})

	t.Run("env fallback", func(t *testing.T) {
		t.Setenv("ENV_KEY", "envval")
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		fs.String("key", "default", "")

		val, err := resolve(fs, "key", "ENV_KEY", parseString, os.Getenv)
		assert.NoError(t, err)
		assert.Equal(t, "envval", val)
	})

	t.Run("default fallback", func(t *testing.T) {
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		fs.String("key", "fallback", "")

		val, err := resolve(fs, "key", "MISSING_ENV", parseString, os.Getenv)
		assert.NoError(t, err)
		assert.Equal(t, "fallback", val)
	})

	t.Run("parse error on flag", func(t *testing.T) {
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		fs.String("key", "default", "")
		_ = fs.Set("key", "bad")

		val, err := resolve(fs, "key", "ENV_KEY", parseString, os.Getenv)
		assert.Error(t, err)
		assert.Equal(t, "", val)
		assert.EqualError(t, err, "parse error")
	})

	t.Run("parse error on env", func(t *testing.T) {
		t.Setenv("ENV_KEY", "bad")
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		fs.String("key", "default", "")

		val, err := resolve(fs, "key", "ENV_KEY", parseString, os.Getenv)
		assert.Error(t, err)
		assert.Equal(t, "", val)
		assert.EqualError(t, err, "parse error")
	})

	t.Run("parse error on default", func(t *testing.T) {
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		fs.String("key", "bad", "")

		val, err := resolve(fs, "key", "MISSING_ENV", parseString, os.Getenv)
		assert.Error(t, err)
		assert.Equal(t, "", val)
		assert.EqualError(t, err, "parse error")
	})

	t.Run("flag not found", func(t *testing.T) {
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)

		val, err := resolve(fs, "missing", "ENV_MISSING", parseString, os.Getenv)
		assert.Error(t, err)
		assert.Equal(t, "", val)
		assert.EqualError(t, err, "flag not found: missing")
	})
}

func TestString(t *testing.T) {
	t.Run("from flag", func(t *testing.T) {
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		fs.String("msg", "default", "")
		_ = fs.Set("msg", " hello ")

		v, err := String(fs, "msg", "ENV_MSG")
		assert.NoError(t, err)
		assert.Equal(t, "hello", v)
	})

	t.Run("from env", func(t *testing.T) {
		t.Setenv("ENV_MSG", " fromenv ")
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		fs.String("msg", "default", "")
		v, err := String(fs, "msg", "ENV_MSG")
		assert.NoError(t, err)
		assert.Equal(t, "fromenv", v)
	})

	t.Run("fallback", func(t *testing.T) {
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		fs.String("msg", "fallback", "")
		v, err := String(fs, "msg", "MISSING_ENV")
		assert.NoError(t, err)
		assert.Equal(t, "fallback", v)
	})
}

func TestInt(t *testing.T) {
	t.Run("from flag", func(t *testing.T) {
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		fs.Int("count", 42, "")
		_ = fs.Set("count", "123")

		v, err := Int(fs, "count", "ENV_COUNT")
		assert.NoError(t, err)
		assert.Equal(t, 123, v)
	})

	t.Run("from env", func(t *testing.T) {
		t.Setenv("ENV_COUNT", "777")
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		fs.Int("count", 42, "")
		v, err := Int(fs, "count", "ENV_COUNT")
		assert.NoError(t, err)
		assert.Equal(t, 777, v)
	})

	t.Run("fallback", func(t *testing.T) {
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		fs.Int("count", 99, "")
		v, err := Int(fs, "count", "MISSING")
		assert.NoError(t, err)
		assert.Equal(t, 99, v)
	})
}

func TestBool(t *testing.T) {
	t.Run("from flag", func(t *testing.T) {
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		fs.Bool("enabled", false, "")
		_ = fs.Set("enabled", "true")

		v, err := Bool(fs, "enabled", "ENV_ENABLED")
		assert.NoError(t, err)
		assert.True(t, v)
	})

	t.Run("from env", func(t *testing.T) {
		t.Setenv("ENV_ENABLED", "true")
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		fs.Bool("enabled", false, "")
		v, err := Bool(fs, "enabled", "ENV_ENABLED")
		assert.NoError(t, err)
		assert.True(t, v)
	})

	t.Run("fallback", func(t *testing.T) {
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		fs.Bool("enabled", true, "")
		v, err := Bool(fs, "enabled", "MISSING")
		assert.NoError(t, err)
		assert.True(t, v)
	})
}

func TestDuration(t *testing.T) {
	t.Run("from flag", func(t *testing.T) {
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		fs.Duration("wait", 1*time.Second, "")
		_ = fs.Set("wait", "3s")

		v, err := Duration(fs, "wait", "ENV_WAIT")
		assert.NoError(t, err)
		assert.Equal(t, 3*time.Second, v)
	})

	t.Run("from env", func(t *testing.T) {
		t.Setenv("ENV_WAIT", "5s")
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		fs.Duration("wait", 2*time.Second, "")
		v, err := Duration(fs, "wait", "ENV_WAIT")
		assert.NoError(t, err)
		assert.Equal(t, 5*time.Second, v)
	})

	t.Run("fallback", func(t *testing.T) {
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		fs.Duration("wait", 7*time.Second, "")
		v, err := Duration(fs, "wait", "MISSING")
		assert.NoError(t, err)
		assert.Equal(t, 7*time.Second, v)
	})
}

func TestHostPort(t *testing.T) {
	t.Run("from flag", func(t *testing.T) {
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		fs.String("addr", "localhost:80", "")
		_ = fs.Set("addr", "127.0.0.1:3000")

		v, err := HostPort(fs, "addr", "ENV_ADDR")
		assert.NoError(t, err)
		assert.Equal(t, "127.0.0.1:3000", v)
	})

	t.Run("from env", func(t *testing.T) {
		t.Setenv("ENV_ADDR", "192.168.1.1:1234")
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		fs.String("addr", "default:9", "")
		v, err := HostPort(fs, "addr", "ENV_ADDR")
		assert.NoError(t, err)
		assert.Equal(t, "192.168.1.1:1234", v)
	})

	t.Run("fallback", func(t *testing.T) {
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		fs.String("addr", "fallback:42", "")
		v, err := HostPort(fs, "addr", "MISSING")
		assert.NoError(t, err)
		assert.Equal(t, "fallback:42", v)
	})

	t.Run("invalid", func(t *testing.T) {
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		fs.String("addr", "invalid", "")
		_, err := HostPort(fs, "addr", "MISSING")
		assert.Error(t, err)
		assert.EqualError(t, err, "invalid format: address invalid: missing port in address")
	})
}

func TestSchemaHostPort(t *testing.T) {
	t.Run("from flag", func(t *testing.T) {
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		fs.String("target", "http://localhost:80", "")
		_ = fs.Set("target", "https://127.0.0.1:443")

		v, err := SchemaHostPort(fs, "target", "ENV_TARGET")
		assert.NoError(t, err)
		assert.Equal(t, "https://127.0.0.1:443", v)
	})

	t.Run("from env", func(t *testing.T) {
		t.Setenv("ENV_TARGET", "http://env-host:9999")
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		fs.String("target", "default://foo", "")
		v, err := SchemaHostPort(fs, "target", "ENV_TARGET")
		assert.NoError(t, err)
		assert.Equal(t, "http://env-host:9999", v)
	})

	t.Run("fallback", func(t *testing.T) {
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		fs.String("target", "http://fallback:8888", "")
		v, err := SchemaHostPort(fs, "target", "MISSING")
		assert.NoError(t, err)
		assert.Equal(t, "http://fallback:8888", v)
	})

	t.Run("missing scheme", func(t *testing.T) {
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		fs.String("target", "invalid", "")
		_, err := SchemaHostPort(fs, "target", "MISSING")
		assert.Error(t, err)
		assert.EqualError(t, err, "missing protocol scheme")
	})

	t.Run("missing host", func(t *testing.T) {
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		fs.String("target", "http://:8888", "")
		res, err := SchemaHostPort(fs, "target", "MISSING")
		assert.Equal(t, "", res)
		assert.Error(t, err)
		assert.EqualError(t, err, "missing host in address")
	})

	t.Run("missing port", func(t *testing.T) {
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		fs.String("target", "http://localhost", "")
		res, err := SchemaHostPort(fs, "target", "MISSING")
		assert.Equal(t, "", res)
		assert.Error(t, err)
		assert.EqualError(t, err, "invalid format: address localhost: missing port in address")
	})

	t.Run("missing host and port", func(t *testing.T) {
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		fs.String("target", "https://", "")
		res, err := SchemaHostPort(fs, "target", "MISSING")
		assert.Equal(t, "", res)
		assert.Error(t, err)
		assert.EqualError(t, err, "invalid format: missing port in address")
	})

	t.Run("invalid", func(t *testing.T) {
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		fs.String("target", "http://::::", "")
		res, err := SchemaHostPort(fs, "target", "MISSING")
		assert.Equal(t, "", res)
		assert.Error(t, err)
		assert.EqualError(t, err, "invalid format: address ::::: too many colons in address")
	})

	t.Run("invalid url", func(t *testing.T) {
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		fs.String("site", "http://example.com/%zz", "") // invalid escape
		val, err := SchemaHostPort(fs, "site", "MISSING")
		assert.Empty(t, val)
		assert.EqualError(t, err, "invalid format: parse \"http://example.com/%zz\": invalid URL escape \"%zz\"")
	})
}

func TestURL(t *testing.T) {
	t.Run("flag wins over env", func(t *testing.T) {
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		fs.String("site", "http://default", "")
		_ = fs.Set("site", "https://flag.local")

		t.Setenv("ENV_SITE", "https://env.local")
		val, err := URL(fs, "site", "ENV_SITE")
		assert.NoError(t, err)
		assert.Equal(t, "https://flag.local", val)
	})

	t.Run("env used if flag not set", func(t *testing.T) {
		t.Setenv("ENV_SITE", "https://env.local")
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		fs.String("site", "http://fallback", "")

		val, err := URL(fs, "site", "ENV_SITE")
		assert.NoError(t, err)
		assert.Equal(t, "https://env.local", val)
	})

	t.Run("default fallback", func(t *testing.T) {
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		fs.String("site", "https://default.local", "")

		val, err := URL(fs, "site", "MISSING_ENV")
		assert.NoError(t, err)
		assert.Equal(t, "https://default.local", val)
	})

	t.Run("missing scheme", func(t *testing.T) {
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		fs.String("site", "localhost:8080", "") // no scheme
		val, err := URL(fs, "site", "MISSING")
		assert.Empty(t, val)
		assert.EqualError(t, err, "missing protocol scheme")
	})

	t.Run("missing host", func(t *testing.T) {
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		fs.String("site", "http://:8080", "") // scheme ok, host missing
		val, err := URL(fs, "site", "MISSING")
		assert.Empty(t, val)
		assert.EqualError(t, err, "missing host in address")
	})

	t.Run("invalid URL", func(t *testing.T) {
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		fs.String("site", "http://example.com/%zz", "") // invalid escape
		val, err := URL(fs, "site", "MISSING")
		assert.Empty(t, val)
		assert.EqualError(t, err, "invalid URL: parse \"http://example.com/%zz\": invalid URL escape \"%zz\"")
	})
}
