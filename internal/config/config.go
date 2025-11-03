package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	User     UserConfig     `yaml:"user"`
	Network  NetworkConfig  `yaml:"network"`
	UI       UIConfig       `yaml:"ui"`
	Security SecurityConfig `yaml:"security"`
	Debug    bool           `yaml:"debug"`
}

// UserConfig contains user-specific settings
type UserConfig struct {
	ID            string `yaml:"id"`
	DisplayName   string `yaml:"display_name"`
	StatusMessage string `yaml:"status_message"`
}

// NetworkConfig contains network-related settings
type NetworkConfig struct {
	RelayServers      []string      `yaml:"relay_servers"`
	P2PEnabled        bool          `yaml:"p2p_enabled"`
	ConnectionTimeout time.Duration `yaml:"connection_timeout"`
	Port              int           `yaml:"port"`
	BindAddress       string        `yaml:"bind_address"`
}

// UIConfig contains user interface settings
type UIConfig struct {
	Theme           string `yaml:"theme"`
	Notifications   bool   `yaml:"notifications"`
	SoundEnabled    bool   `yaml:"sound_enabled"`
	TimestampFormat string `yaml:"timestamp_format"`
	ShowTyping      bool   `yaml:"show_typing"`
	CompactMode     bool   `yaml:"compact_mode"`
}

// SecurityConfig contains security-related settings
type SecurityConfig struct {
	AutoAcceptKeys       bool   `yaml:"auto_accept_keys"`
	MessageRetentionDays int    `yaml:"message_retention_days"`
	ExportKeysPath       string `yaml:"export_keys_path"`
	RequireVerification  bool   `yaml:"require_verification"`
}

// Default returns a configuration with sensible defaults
func Default() *Config {
	homeDir, _ := os.UserHomeDir()
	
	return &Config{
		User: UserConfig{
			DisplayName:   "SecureChat User",
			StatusMessage: "Available",
		},
		Network: NetworkConfig{
			RelayServers: []string{
				"relay1.securechat.dev:8080",
				"relay2.securechat.dev:8080",
			},
			P2PEnabled:        true,
			ConnectionTimeout: 30 * time.Second,
			Port:              8080,
			BindAddress:       "0.0.0.0",
		},
		UI: UIConfig{
			Theme:           "dark",
			Notifications:   true,
			SoundEnabled:    false,
			TimestampFormat: "15:04",
			ShowTyping:      true,
			CompactMode:     false,
		},
		Security: SecurityConfig{
			AutoAcceptKeys:       false,
			MessageRetentionDays: 30,
			ExportKeysPath:       filepath.Join(homeDir, ".config", "securechat", "keys"),
			RequireVerification:  true,
		},
		Debug: false,
	}
}

// LoadFromFile loads configuration from a YAML file
func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := Default()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return cfg, nil
}

// SaveToFile saves the configuration to a YAML file
func (c *Config) SaveToFile(path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.User.DisplayName == "" {
		return fmt.Errorf("user display name cannot be empty")
	}

	if len(c.Network.RelayServers) == 0 && !c.Network.P2PEnabled {
		return fmt.Errorf("at least one relay server must be configured or P2P must be enabled")
	}

	if c.Network.ConnectionTimeout <= 0 {
		return fmt.Errorf("connection timeout must be positive")
	}

	if c.Security.MessageRetentionDays < 0 {
		return fmt.Errorf("message retention days cannot be negative")
	}

	validThemes := map[string]bool{
		"dark":  true,
		"light": true,
		"auto":  true,
	}
	if !validThemes[c.UI.Theme] {
		return fmt.Errorf("invalid theme: %s", c.UI.Theme)
	}

	return nil
}

// GetDataDir returns the data directory for the application
func (c *Config) GetDataDir() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".local", "share", "securechat")
}

// GetCacheDir returns the cache directory for the application
func (c *Config) GetCacheDir() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".cache", "securechat")
}

// GetConfigDir returns the configuration directory for the application
func (c *Config) GetConfigDir() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".config", "securechat")
}
