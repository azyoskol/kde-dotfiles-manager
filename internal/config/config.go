package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	DotfilesDir         string   `yaml:"dotfiles_dir"`
	GitRepo             string   `yaml:"git_repo,omitempty"`
	Categories          []string `yaml:"categories"`
	BackupBeforeRestore bool     `yaml:"backup_before_restore"`
	Verbose             bool     `yaml:"verbose"`
	Profile             string   `yaml:"profile,omitempty"`
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		DotfilesDir:         "~/kde-dotfiles",
		Profile:             "default",
		Categories:          defaultCategories(),
		BackupBeforeRestore: true,
		Verbose:             false,
	}
}

// defaultCategories returns the list of configuration categories to manage
func defaultCategories() []string {
	return []string{
		"shortcuts",
		"themes",
		"window_management",
		"languages",
		"widgets",
		"panels",
		"system_settings",
	}
}

// ConfigPath returns the path to the configuration file
func ConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "kde-dotfiles-manager", "config.yaml")
}

// Load reads configuration from file, falling back to defaults
func Load() (*Config, error) {
	cfg := DefaultConfig()
	path := ConfigPath()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return cfg, nil
}

// Save writes the configuration to file
func (c *Config) Save() error {
	path := ConfigPath()
	dir := filepath.Dir(path)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// ExpandPath resolves ~ to the user's home directory
func (c *Config) ExpandPath() string {
	dir := c.DotfilesDir
	if len(dir) >= 1 && dir[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return dir
		}
		return filepath.Join(home, dir[1:])
	}
	return dir
}

// GetProfileDotfilesDir returns the dotfiles directory for the current profile
func (c *Config) GetProfileDotfilesDir() string {
	baseDir := c.ExpandPath()
	if c.Profile == "" || c.Profile == "default" {
		return baseDir
	}
	return filepath.Join(baseDir, "profiles", c.Profile)
}

// SetProfile changes the current profile and updates the config file
func (c *Config) SetProfile(profile string) error {
	c.Profile = profile
	return c.Save()
}

// ListProfiles returns a list of available profiles
func ListProfiles() ([]string, error) {
	cfg := DefaultConfig()
	baseDir := cfg.ExpandPath()
	
	profiles := []string{"default"}
	
	// Check if base directory exists
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		return profiles, nil
	}
	
	profilesDir := filepath.Join(baseDir, "profiles")
	if _, err := os.Stat(profilesDir); os.IsNotExist(err) {
		return profiles, nil
	}
	
	entries, err := os.ReadDir(profilesDir)
	if err != nil {
		return profiles, err
	}
	
	for _, entry := range entries {
		if entry.IsDir() {
			profiles = append(profiles, entry.Name())
		}
	}
	
	return profiles, nil
}
