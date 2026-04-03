package widgets

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// WidgetInfo represents a single Plasma widget
type WidgetInfo struct {
	Name        string `yaml:"name"`
	Plugin      string `yaml:"plugin"`
	Position    string `yaml:"position"`
	Size        string `yaml:"size"`
	Config      string `yaml:"config"`
	IsCustom    bool   `yaml:"is_custom"`
}

// WidgetConfig holds all widget configurations
type WidgetConfig struct {
	DesktopWidgets []WidgetInfo `yaml:"desktop_widgets"`
	PanelWidgets   []WidgetInfo `yaml:"panel_widgets"`
	LayoutTemplate string       `yaml:"layout_template"`
}

// ParseDesktopAppletSrc parses the plasma desktop applet source file
func ParseDesktopAppletSrc(path string) (*WidgetConfig, error) {
	cfg := &WidgetConfig{
		DesktopWidgets: []WidgetInfo{},
		PanelWidgets:   []WidgetInfo{},
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open plasma desktop config: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	currentWidget := ""
	var widgets []WidgetInfo
	isPanel := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Section headers like [Containments][1][General] or [Containments][2][Applets][3][General]
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			// Save previous widget if any
			if currentWidget != "" {
				widgets = append(widgets, WidgetInfo{Name: currentWidget})
			}

			header := line[1 : len(line)-1]

			// Detect if this is a panel or desktop containment
			if strings.Contains(header, "Containments") {
				parts := strings.Split(header, "][")
				if len(parts) >= 2 {
					// Check containment type
					if strings.Contains(header, "Applets") {
						isPanel = false
						widgetInfo := parseWidgetHeader(header)
						if widgetInfo != "" {
							currentWidget = widgetInfo
						}
					}
				}
			}
			continue
		}

		// Parse widget properties
		if currentWidget != "" {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				// Look for plugin name
				if key == "plugin" {
					widgets = append(widgets, WidgetInfo{
						Plugin:   value,
						IsCustom: !strings.HasPrefix(value, "org.kde."),
					})
					currentWidget = ""
				}
			}
		}
	}

	// Determine widget types
	for _, w := range widgets {
		if isPanel {
			cfg.PanelWidgets = append(cfg.PanelWidgets, w)
		} else {
			cfg.DesktopWidgets = append(cfg.DesktopWidgets, w)
		}
	}

	return cfg, scanner.Err()
}

// parseWidgetHeader extracts widget identifier from section header
func parseWidgetHeader(header string) string {
	parts := strings.Split(header, "][")
	if len(parts) < 4 {
		return ""
	}
	// Format: [Containments][X][Applets][Y][General]
	if len(parts) >= 4 {
		return parts[3] // Widget ID
	}
	return ""
}

// ListInstalledWidgets lists all installed Plasma widgets
func ListInstalledWidgets(dataDir string) ([]WidgetInfo, error) {
	widgetsDir := filepath.Join(dataDir, "plasma", "plasmoids")
	var result []WidgetInfo

	entries, err := os.ReadDir(widgetsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return result, nil
		}
		return nil, fmt.Errorf("failed to read widgets directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			widget := WidgetInfo{
				Name:     entry.Name(),
				Plugin:   entry.Name(),
				IsCustom: !strings.HasPrefix(entry.Name(), "org.kde."),
			}
			result = append(result, widget)
		}
	}

	return result, nil
}

// Backup copies widget configuration files
func Backup(srcPaths map[string]string, destDir string) error {
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	for name, srcPath := range srcPaths {
		info, err := os.Stat(srcPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("failed to stat %s: %w", name, err)
		}

		destPath := filepath.Join(destDir, filepath.Base(srcPath))

		if info.IsDir() {
			if err := copyDir(srcPath, destPath); err != nil {
				return fmt.Errorf("failed to copy directory %s: %w", name, err)
			}
		} else {
			if err := copyFile(srcPath, destPath); err != nil {
				return fmt.Errorf("failed to copy file %s: %w", name, err)
			}
		}
	}

	return nil
}

// Restore copies widget files from backup to original locations
func Restore(backupDir string, srcPaths map[string]string) error {
	for name, destPath := range srcPaths {
		srcFile := filepath.Join(backupDir, filepath.Base(destPath))

		if _, err := os.Stat(srcFile); os.IsNotExist(err) {
			continue
		}

		parent := filepath.Dir(destPath)
		if err := os.MkdirAll(parent, 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", name, err)
		}

		if err := copyFile(srcFile, destPath); err != nil {
			return fmt.Errorf("failed to restore %s: %w", name, err)
		}
	}

	return nil
}

// copyFile copies a single file
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

// copyDir recursively copies a directory
func copyDir(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}
