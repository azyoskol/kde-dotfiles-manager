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
	// All profiles including 'default' are stored in the profiles subdirectory
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

// CreateProfile creates a new profile directory
func CreateProfile(name string) error {
	if name == "" {
		return fmt.Errorf("profile name cannot be empty")
	}
	
	if name == "default" {
		return fmt.Errorf("cannot create profile with reserved name 'default'")
	}
	
	cfg := DefaultConfig()
	baseDir := cfg.ExpandPath()
	profileDir := filepath.Join(baseDir, "profiles", name)
	
	// Check if profile already exists
	if _, err := os.Stat(profileDir); err == nil {
		return fmt.Errorf("profile '%s' already exists", name)
	}
	
	if err := os.MkdirAll(profileDir, 0755); err != nil {
		return fmt.Errorf("failed to create profile directory: %w", err)
	}
	
	return nil
}

// DeleteProfile removes a profile and its data
func DeleteProfile(name string) error {
	if name == "" {
		return fmt.Errorf("profile name cannot be empty")
	}
	
	if name == "default" {
		return fmt.Errorf("cannot delete default profile")
	}
	
	cfg := DefaultConfig()
	baseDir := cfg.ExpandPath()
	profileDir := filepath.Join(baseDir, "profiles", name)
	
	// Check if profile exists
	if _, err := os.Stat(profileDir); os.IsNotExist(err) {
		return fmt.Errorf("profile '%s' does not exist", name)
	}
	
	if err := os.RemoveAll(profileDir); err != nil {
		return fmt.Errorf("failed to delete profile directory: %w", err)
	}
	
	return nil
}

// RenameProfile renames an existing profile
func RenameProfile(oldName, newName string) error {
	if oldName == "" || newName == "" {
		return fmt.Errorf("profile names cannot be empty")
	}
	
	if newName == "default" {
		return fmt.Errorf("cannot rename to reserved name 'default'")
	}
	
	cfg := DefaultConfig()
	baseDir := cfg.ExpandPath()
	oldProfileDir := filepath.Join(baseDir, "profiles", oldName)
	newProfileDir := filepath.Join(baseDir, "profiles", newName)
	
	// Check if old profile exists
	if _, err := os.Stat(oldProfileDir); os.IsNotExist(err) {
		return fmt.Errorf("profile '%s' does not exist", oldName)
	}
	
	// Check if new profile already exists
	if _, err := os.Stat(newProfileDir); err == nil {
		return fmt.Errorf("profile '%s' already exists", newName)
	}
	
	if err := os.Rename(oldProfileDir, newProfileDir); err != nil {
		return fmt.Errorf("failed to rename profile directory: %w", err)
	}
	
	// Update config if current profile was renamed
	currentCfg, err := Load()
	if err == nil && currentCfg.Profile == oldName {
		currentCfg.Profile = newName
		currentCfg.Save()
	}
	
	return nil
}

// ProfileExists checks if a profile exists
func ProfileExists(name string) bool {
	if name == "default" {
		return true
	}
	
	cfg := DefaultConfig()
	baseDir := cfg.ExpandPath()
	profileDir := filepath.Join(baseDir, "profiles", name)
	
	_, err := os.Stat(profileDir)
	return err == nil
}
