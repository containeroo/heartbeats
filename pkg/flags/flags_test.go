package flags

import (
	"heartbeats/pkg/config"
	"os"
	"strings"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func resetFlags() {
	pflag.CommandLine = pflag.NewFlagSet("heartbeats", pflag.ExitOnError)
}

func TestParseFlags(t *testing.T) {
	resetFlags()

	output := &strings.Builder{}
	args := []string{"cmd", "-c", "config.yaml", "-l", "127.0.0.1:9090", "-s", "http://example.com", "-m", "200", "-r", "20", "-v"}

	result := ParseFlags(args, output)
	assert.NoError(t, result.Err)
	assert.Equal(t, result.ShowVersion, false)
	assert.Equal(t, result.ShowHelp, false)

	assert.Equal(t, "config.yaml", config.App.Path)
	assert.Equal(t, "127.0.0.1:9090", config.App.Server.ListenAddress)
	assert.Equal(t, "http://example.com", config.App.Server.SiteRoot)
	assert.Equal(t, 200, config.App.Cache.MaxSize)
	assert.Equal(t, 20, config.App.Cache.Reduce)
	assert.True(t, config.App.Verbose)
}

func TestShowVersionFlag(t *testing.T) {
	resetFlags()

	output := &strings.Builder{}
	args := []string{"cmd", "--version"}
	result := ParseFlags(args, output)
	assert.NoError(t, result.Err)
	assert.Equal(t, result.ShowVersion, true)
	assert.Equal(t, result.ShowHelp, false)
}

func TestShowHelpFlag(t *testing.T) {
	resetFlags()

	output := &strings.Builder{}
	args := []string{"cmd", "--help"}

	result := ParseFlags(args, output)
	assert.NoError(t, result.Err)
	assert.Equal(t, result.ShowVersion, false)
	assert.Equal(t, result.ShowHelp, true)
}

func TestProcessEnvVariables(t *testing.T) {
	resetFlags()

	os.Setenv("HEARTBEATS_CONFIG", "env_config.yaml")
	os.Setenv("HEARTBEATS_LISTEN_ADDRESS", "0.0.0.0:8080")
	os.Setenv("HEARTBEATS_SITE_ROOT", "http://env.com")
	os.Setenv("HEARTBEATS_MAX_SIZE", "300")
	os.Setenv("HEARTBEATS_REDUCE", "30")
	os.Setenv("HEARTBEATS_VERBOSE", "true")

	pflag.StringVarP(&config.App.Path, "config", "c", "./deploy/config.yaml", "Path to the configuration file")
	pflag.StringVarP(&config.App.Server.ListenAddress, "listen-address", "l", "localhost:8080", "Address to listen on")
	pflag.StringVarP(&config.App.Server.SiteRoot, "site-root", "s", "", "Site root for the heartbeat service (default \"http://<listenAddress>\")")
	pflag.IntVarP(&config.App.Cache.MaxSize, "max-size", "m", 100, "Maximum size of the cache")
	pflag.IntVarP(&config.App.Cache.Reduce, "reduce", "r", 10, "Amount to reduce when max size is exceeded")
	pflag.BoolVarP(&config.App.Verbose, "verbose", "v", false, "Enable verbose logging")

	pflag.Parse()

	processEnvVariables()

	assert.Equal(t, "env_config.yaml", config.App.Path)
	assert.Equal(t, "0.0.0.0:8080", config.App.Server.ListenAddress)
	assert.Equal(t, "http://env.com", config.App.Server.SiteRoot)
	assert.Equal(t, 300, config.App.Cache.MaxSize)
	assert.Equal(t, 30, config.App.Cache.Reduce)
	assert.True(t, config.App.Verbose)
}
