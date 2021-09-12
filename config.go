package oogway

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/rjeczalik/notify"
)

const (
	configFile = "config.toml"
)

var (
	cfg Config
)

// Config is the Oogway application config.
type Config struct {
	Server  ServerConfig
	Content ContentConfig
}

// ServerConfig is the HTTP server configuration.
type ServerConfig struct {
	Host            string
	Port            int
	ShutdownTimeout int `toml:"shutdown_time"`
	WriteTimeout    int `toml:"write_timeout"`
	ReadTimeout     int `toml:"read_timeout"`
}

// ContentConfig is the content configuration.
type ContentConfig struct {
	NotFound string `toml:"not_found"`
}

func loadConfig(dir string) error {
	content, err := os.ReadFile(filepath.Join(dir, configFile))

	if err != nil {
		return fmt.Errorf("error loading config.toml: %s", err)
	}

	if _, err := toml.Decode(string(content), &cfg); err != nil {
		return fmt.Errorf("error loading config.toml: %s", err)
	}

	setConfigDefaults()
	return nil
}

func setConfigDefaults() {
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}

	if cfg.Server.ShutdownTimeout == 0 {
		cfg.Server.ShutdownTimeout = 30
	}

	if cfg.Server.WriteTimeout == 0 {
		cfg.Server.WriteTimeout = 5
	}

	if cfg.Server.ReadTimeout == 0 {
		cfg.Server.ReadTimeout = 5
	}
}

func watchConfig(ctx context.Context, dir string) error {
	if err := loadConfig(dir); err != nil {
		return err
	}

	change := make(chan notify.EventInfo, 1)

	go func() {
		for {
			select {
			case <-change:
				if err := loadConfig(dir); err != nil {
					log.Printf("Error updating config.toml: %s", err)
				}
			case <-ctx.Done():
				notify.Stop(change)
				return
			}
		}
	}()

	if err := notify.Watch(filepath.Join(dir, configFile), change, notify.Write); err != nil {
		return err
	}

	return nil
}
