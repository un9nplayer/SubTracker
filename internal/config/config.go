// Package config manages persistent SubTracker configuration stored on disk.
// Config file location:
//   - Linux/macOS: $HOME/.subtracker/config.json
//   - Windows:     %APPDATA%\.subtracker\config.json
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// Config holds all persistent SubTracker settings.
type Config struct {
	APIKey string `json:"api_key"`
}

// configDir returns the OS-appropriate directory for SubTracker config files.
func configDir() (string, error) {
	var base string

	if runtime.GOOS == "windows" {
		base = os.Getenv("APPDATA")
		if base == "" {
			return "", fmt.Errorf("%%APPDATA%% environment variable is not set")
		}
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot determine home directory: %w", err)
		}
		base = home
	}

	return filepath.Join(base, ".subtracker"), nil
}

// Load reads the config from disk. Returns an empty Config (no error) if the
// file does not exist yet.
func Load() (*Config, error) {
	dir, err := configDir()
	if err != nil {
		return &Config{}, err
	}

	path := filepath.Join(dir, "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil // first run — no config yet
		}
		return &Config{}, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return &Config{}, fmt.Errorf("malformed config file: %w", err)
	}

	return &cfg, nil
}

// Save writes cfg to disk (permissions 0600) and returns the path used.
// The directory is created if it does not exist.
func Save(cfg *Config) (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}

	// Create directory with restricted permissions
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("cannot create config directory: %w", err)
	}

	path := filepath.Join(dir, "config.json")

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to encode config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return "", fmt.Errorf("failed to write config: %w", err)
	}

	return path, nil
}
