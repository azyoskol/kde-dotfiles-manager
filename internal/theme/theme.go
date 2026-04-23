package theme

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/user/kde-dotfiles-manager/internal/fileutil"
)

// ThemeConfig holds KDE theme configuration data
type ThemeConfig struct {
	ColorScheme       string `yaml:"color_scheme"`
	WindowDecoration  string `yaml:"window_decoration"`
	IconTheme         string `yaml:"icon_theme"`
	CursorTheme       string `yaml:"cursor_theme"`
	Font              string `yaml:"font"`
	FontSize          int    `yaml:"font_size"`
	Wallpaper         string `yaml:"wallpaper"`
	LookAndFeel       string `yaml:"look_and_feel"`
	GTKTheme          string `yaml:"gtk_theme"`
	ApplicationTheme  string `yaml:"application_theme"`
}

// ExtractFromKdeglobals parses kdeglobals file for theme settings
func ExtractFromKdeglobals(path string) (*ThemeConfig, error) {
	cfg := &ThemeConfig{}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open kdeglobals: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	currentSection := ""

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Section header
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = line[1 : len(line)-1]
			continue
		}

		// Key=Value pairs
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch currentSection {
		case "General":
			switch key {
			case "ColorScheme":
				cfg.ColorScheme = value
			case "widgetStyle":
				cfg.ApplicationTheme = value
			}
		case "WM":
			switch key {
			case "activeFont":
				cfg.Font = value
			}
		case "Icons":
			switch key {
			case "Theme":
				cfg.IconTheme = value
			}
		case "Mouse":
			switch key {
			case "cursorTheme":
				cfg.CursorTheme = value
			}
		}
	}

	return cfg, scanner.Err()
}

// ExtractFromPlasmaRc parses plasmarc for look and feel settings
func ExtractFromPlasmaRc(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open plasmarc: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	currentSection := ""

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = line[1 : len(line)-1]
			continue
		}

		if currentSection == "Theme" {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 && strings.TrimSpace(parts[0]) == "name" {
				return strings.TrimSpace(parts[1]), nil
			}
		}
	}

	return "", scanner.Err()
}

// ExtractWallpaperFromPlasmaDesktop parses the plasma desktop applet source for wallpaper
func ExtractWallpaperFromPlasmaDesktop(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open plasma desktop config: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	currentSection := ""

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Track sections to find wallpaper settings in containment sections
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = line[1 : len(line)-1]
			continue
		}

		// Look for wallpaper settings in various formats
		if strings.Contains(currentSection, "Containments") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				
				// Check for common wallpaper keys
				if key == "Image" || key == "wallpaper" || key == "wallpaperimage" {
					return value, nil
				}
			}
		}
	}

	return "", scanner.Err()
}

// Backup copies theme-related files to the destination directory
func Backup(srcPaths map[string]string, destDir string) error {
	if err := fileutil.EnsureDir(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	for name, srcPath := range srcPaths {
		// Check if source exists
		info, err := os.Stat(srcPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue // Skip non-existent files
			}
			return fmt.Errorf("failed to stat %s: %w", name, err)
		}

		destPath := filepath.Join(destDir, filepath.Base(srcPath))

		if info.IsDir() {
			if err := fileutil.CopyDir(srcPath, destPath); err != nil {
				return fmt.Errorf("failed to copy directory %s: %w", name, err)
			}
		} else {
			if err := fileutil.CopyFile(srcPath, destPath); err != nil {
				return fmt.Errorf("failed to copy file %s: %w", name, err)
			}
		}
	}

	return nil
}

// Restore copies theme files from the backup to their original locations
func Restore(backupDir string, srcPaths map[string]string) error {
	for name, destPath := range srcPaths {
		srcFile := filepath.Join(backupDir, filepath.Base(destPath))

		// Check if backup exists
		if _, err := os.Stat(srcFile); os.IsNotExist(err) {
			continue // Skip if backup doesn't exist
		}

		// Create parent directory if needed
		parent := filepath.Dir(destPath)
		if err := fileutil.EnsureDir(parent, 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", name, err)
		}

		if err := fileutil.CopyFile(srcFile, destPath); err != nil {
			return fmt.Errorf("failed to restore %s: %w", name, err)
		}
	}

	return nil
}
