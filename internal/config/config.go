package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds the CLI configuration
type Config struct {
	APIToken     string `yaml:"api_token,omitempty"`
	APIKey       string `yaml:"api_key,omitempty"`
	APIEmail     string `yaml:"api_email,omitempty"`
	OutputFormat string `yaml:"output_format,omitempty"`
}

// DefaultConfigPath returns the default config file path
func DefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".cloudflare", "config.yaml")
}

// Load loads configuration from file and environment variables.
// Environment variables take precedence over config file values.
func Load(configPath string) (*Config, error) {
	cfg := &Config{}

	// Try to load from config file
	if configPath == "" {
		configPath = DefaultConfigPath()
	}

	if configPath != "" {
		if data, err := os.ReadFile(configPath); err == nil {
			_ = yaml.Unmarshal(data, cfg)
		}
		// Ignore file read errors - config file is optional
	}

	// Environment variables override config file (check multiple env var names)
	if token := getEnv("CLOUDFLARE_API_TOKEN", "CF_API_TOKEN"); token != "" {
		cfg.APIToken = token
	}
	if key := getEnv("CLOUDFLARE_API_KEY", "CF_API_KEY"); key != "" {
		cfg.APIKey = key
	}
	if email := getEnv("CLOUDFLARE_API_EMAIL", "CF_API_EMAIL"); email != "" {
		cfg.APIEmail = email
	}

	return cfg, nil
}

// getEnv returns the first non-empty environment variable from the given names
func getEnv(names ...string) string {
	for _, name := range names {
		if val := os.Getenv(name); val != "" {
			return val
		}
	}
	return ""
}

// HasCredentials returns true if valid credentials are configured
func (c *Config) HasCredentials() bool {
	return c.APIToken != "" || (c.APIKey != "" && c.APIEmail != "")
}

// AuthMethod returns a description of the configured auth method
func (c *Config) AuthMethod() string {
	if c.APIToken != "" {
		return "API Token"
	}
	if c.APIKey != "" && c.APIEmail != "" {
		return "API Key + Email"
	}
	return "None"
}

// Save saves the configuration to a file
func (c *Config) Save(configPath string) error {
	if configPath == "" {
		configPath = DefaultConfigPath()
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0600)
}
