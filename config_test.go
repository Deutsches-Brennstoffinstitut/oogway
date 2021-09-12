package oogway

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	sampleConfig = `[server]
host = "localhost"
port = 8080
shutdown_time = 9
write_timeout = 10
read_timeout = 11

[content]
not_found = "not-found"`
)

func TestLoadConfig(t *testing.T) {
	assert.NoError(t, os.RemoveAll("config.toml"))
	assert.NoError(t, os.WriteFile("config.toml", []byte(sampleConfig), 0644))
	assert.NoError(t, loadConfig("."))
	assert.Equal(t, "localhost", cfg.Server.Host)
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, 9, cfg.Server.ShutdownTimeout)
	assert.Equal(t, 10, cfg.Server.WriteTimeout)
	assert.Equal(t, 11, cfg.Server.ReadTimeout)
	assert.Equal(t, "not-found", cfg.Content.NotFound)
}

func TestLoadConfigNotExists(t *testing.T) {
	assert.NoError(t, os.RemoveAll("config.toml"))
	err := loadConfig(".")
	assert.NotNil(t, err)
	assert.Equal(t, "error loading config.toml: open config.toml: no such file or directory", err.Error())
}

func TestWatchConfig(t *testing.T) {
	assert.NoError(t, os.RemoveAll("config.toml"))
	assert.NoError(t, os.WriteFile("config.toml", []byte(sampleConfig), 0644))
	ctx, cancel := context.WithCancel(context.Background())
	assert.NoError(t, watchConfig(ctx, "."))
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.NoError(t, os.WriteFile("config.toml", []byte(strings.Replace(sampleConfig, "8080", "8888", 1)), 0644))
	time.Sleep(time.Millisecond * 10)
	cancel()
	assert.Equal(t, 8888, cfg.Server.Port)
}

func TestWatchConfigNotExists(t *testing.T) {
	assert.NoError(t, os.RemoveAll("config.toml"))
	ctx, cancel := context.WithCancel(context.Background())
	err := watchConfig(ctx, ".")
	assert.Equal(t, "error loading config.toml: open config.toml: no such file or directory", err.Error())
	cancel()
}
