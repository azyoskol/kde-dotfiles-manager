package shortcuts

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/user/kde-dotfiles-manager/internal/fileutil"
)

// ShortcutEntry represents a single keyboard shortcut
type ShortcutEntry struct {
	Application string
	Action      string
	Shortcuts   []string
	Default     string
}

// ShortcutConfig holds all keyboard shortcut configurations
type ShortcutConfig struct {
	GlobalShortcuts []ShortcutEntry
	KWinShortcuts   []ShortcutEntry
	AppShortcuts    []ShortcutEntry
}

// ParseKGlobalShortcuts parses the kglobalshortcutsrc file
func ParseKGlobalShortcuts(path string) (*ShortcutConfig, error) {
	cfg := &ShortcutConfig{}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open kglobalshortcutsrc: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	currentApp := ""
	var currentShortcuts []ShortcutEntry

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Section header (application name)
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			// Save previous section
			if currentApp != "" {
				cfg.addShortcuts(currentApp, currentShortcuts)
			}

			currentApp = line[1 : len(line)-1]
			currentShortcuts = nil
			continue
		}

		// Parse shortcut entry: ActionName=shortcut1,shortcut2\tdefault
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		action := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Split by tab to get shortcuts and default
		valueParts := strings.SplitN(value, "\t", 2)
		shortcutsStr := valueParts[0]
		defaultStr := ""
		if len(valueParts) > 1 {
			defaultStr = valueParts[1]
		}

		// Parse shortcuts (comma-separated)
		var shortcuts []string
		if shortcutsStr != "" {
			shortcuts = strings.Split(shortcutsStr, ",")
		}

		entry := ShortcutEntry{
			Action:    action,
			Shortcuts: shortcuts,
			Default:   defaultStr,
		}
		currentShortcuts = append(currentShortcuts, entry)
	}

	// Save last section
	if currentApp != "" {
		cfg.addShortcuts(currentApp, currentShortcuts)
	}

	return cfg, scanner.Err()
}

// addShortcuts categorizes shortcuts by application type
func (cfg *ShortcutConfig) addShortcuts(app string, entries []ShortcutEntry) {
	for i := range entries {
		entries[i].Application = app
	}

	// Categorize based on application name
	switch {
	case app == "kwin":
		cfg.KWinShortcuts = append(cfg.KWinShortcuts, entries...)
	case strings.Contains(app, "kwin"):
		cfg.KWinShortcuts = append(cfg.KWinShortcuts, entries...)
	default:
		cfg.GlobalShortcuts = append(cfg.GlobalShortcuts, entries...)
	}
}

// ParseKHotkeys parses the khotkeysrc file for custom hotkeys
func ParseKHotkeys(path string) ([]ShortcutEntry, error) {
	var entries []ShortcutEntry

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open khotkeysrc: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	currentAction := ""

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentAction = line[1 : len(line)-1]
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if key == "Shortcut" && currentAction != "" {
			entries = append(entries, ShortcutEntry{
				Application: "khotkeys",
				Action:      currentAction,
				Shortcuts:   strings.Split(value, ","),
			})
		}
	}

	return entries, scanner.Err()
}

// Backup copies shortcut configuration files
func Backup(srcPath, destPath string) error {
	if !fileutil.FileExists(srcPath) {
		return nil // File doesn't exist, skip
	}

	if err := fileutil.EnsureDir(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	return fileutil.CopyFile(srcPath, destPath)
}

// Restore copies shortcut configuration files back
func Restore(srcPath, destPath string) error {
	if !fileutil.FileExists(srcPath) {
		return nil // Backup doesn't exist, skip
	}

	// Create parent directory if needed
	parent := filepath.Dir(destPath)
	if err := fileutil.EnsureDir(parent, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	return fileutil.CopyFile(srcPath, destPath)
}
