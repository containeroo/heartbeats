# envflag

`envflag` is a thin wrapper around [`spf13/pflag`](https://github.com/spf13/pflag) that resolves typed flag values from **CLI flags, environment variables, or defaults**, in that order.

## 📦 Requirements

- Requires [`pflag`](https://github.com/spf13/pflag)
- Intended to be used **with** your existing `pflag.FlagSet`

## ✅ Features

- Resolves typed values with `flag > env > default` priority
- Supports:
  - `String`, `Int`, `Bool`, `Duration`
  - `HostPort` (validates `host:port`)
  - `SchemaHostPort` (validates `schema://host:port`)
  - `URL` (validates scheme + hostname)

## ✨ Example

```go
package main

import (
    "fmt"
    "os"
    "time"

    "github.com/spf13/pflag"
    "github.com/containeroo/heartbeats/pkg/envflag"
)

type Options struct {
    Addr        string
    Debug       bool
    RetryDelay  time.Duration
}

func buildOptions(fs *pflag.FlagSet) (opts Options, err error) {
    defer func() {
        // catch panic from must(...) calls to avoid repetitive `if err != nil` checks
        // and convert them into a single error return instead
        if r := recover(); r != nil {
            err = fmt.Errorf("buildOptions failed: %v", r)
        }
    }()
    return Options{
        Addr:       must(envflag.HostPort(fs, "listen", "APP_LISTEN")),
        Debug:      must(envflag.Bool(fs, "debug", "APP_DEBUG")),
        RetryDelay: must(envflag.Duration(fs, "retry-delay", "APP_RETRY_DELAY")),
    }, nil
}

// must panics on error to simplify error handling in buildOptions.
func must[T any](v T, err error) T {
    if err != nil {
        panic(err)
    }
    return v
}

func main() {
    fs := pflag.NewFlagSet("app", pflag.ExitOnError)
    fs.String("listen", ":8080", "Listen address (env: APP_LISTEN)")
    fs.Bool("debug", false, "Enable debug logging (env: APP_DEBUG)")
    fs.Duration("retry-delay", 2*time.Second, "Retry delay (env: APP_RETRY_DELAY)")

    fs.Parse(os.Args[1:]) // nolint:errcheck

    opts, err := buildOptions(fs)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to load flags: %s\n", err)
        os.Exit(1)
    }

    fmt.Printf("Starting app at %s (debug=%v)\n", opts.Addr, opts.Debug)
}
```

## 🧪 Order of Precedence

1. Flag explicitly set on CLI
2. Environment variable (e.g. `APP_LISTEN`)
3. Flag default value
