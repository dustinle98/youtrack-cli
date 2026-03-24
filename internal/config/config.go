package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds the CLI configuration.
type Config struct {
	URL            string `yaml:"url"`
	Token          string `yaml:"token"`
	VerifySSL      bool   `yaml:"verify_ssl"`
	DefaultProject string `yaml:"default_project,omitempty"`
}

var configDir string
var configFile string

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	configDir = filepath.Join(home, ".config", "ytc")
	configFile = filepath.Join(configDir, "config.yml")
}

// Load reads the config from file and env vars. Env vars take precedence.
func Load() (*Config, error) {
	cfg := &Config{
		VerifySSL: true,
	}

	// Try reading config file
	data, err := os.ReadFile(configFile)
	if err == nil {
		_ = yaml.Unmarshal(data, cfg)
	}

	// Env var overrides
	if v := os.Getenv("YOUTRACK_URL"); v != "" {
		cfg.URL = v
	}
	if v := os.Getenv("YOUTRACK_API_TOKEN"); v != "" {
		cfg.Token = v
	}
	if v := os.Getenv("YOUTRACK_VERIFY_SSL"); v == "false" || v == "0" {
		cfg.VerifySSL = false
	}

	return cfg, nil
}

// Save writes the config to disk.
func Save(cfg *Config) error {
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	if err := os.WriteFile(configFile, data, 0600); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}

// Validate checks that required fields are present.
func (c *Config) Validate() error {
	if c.URL == "" {
		return fmt.Errorf("YouTrack URL is required. Run 'ytc auth login' or set YOUTRACK_URL")
	}
	if c.Token == "" {
		return fmt.Errorf("YouTrack API token is required. Run 'ytc auth login' or set YOUTRACK_API_TOKEN")
	}
	return nil
}

// ConfigFilePath returns the path to the config file.
func ConfigFilePath() string {
	return configFile
}
