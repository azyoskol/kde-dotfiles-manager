package widgets

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
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
	isPanel := false
	currentWidget := WidgetInfo{}
	inAppletSection := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Section headers like [Containments][1][General] or [Containments][2][Applets][3][General]
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			// Save previous widget if any
			if currentWidget.Plugin != "" {
				if isPanel {
					cfg.PanelWidgets = append(cfg.PanelWidgets, currentWidget)
				} else {
					cfg.DesktopWidgets = append(cfg.DesktopWidgets, currentWidget)
				}
				currentWidget = WidgetInfo{}
				inAppletSection = false
			}

			header := line[1 : len(line)-1]

			// Detect if this is a panel or desktop containment
			if strings.Contains(header, "Containments") {
				parts := strings.Split(header, "][")
				if len(parts) >= 2 {
					// Check containment type - panels typically have specific containment plugins
					if strings.Contains(header, "org.kde.panel") || 
					   strings.Contains(header, "org.kde.plasma.panel") {
						isPanel = true
					} else {
						isPanel = false
					}
				}
			}

			// Check if we're in an Applets section (widget configuration)
			if strings.Contains(header, "Applets") {
				inAppletSection = true
				parts := strings.Split(header, "][")
				if len(parts) >= 4 {
					currentWidget.Name = parts[3] // Widget ID
				}
			}
			continue
		}

		// Parse widget properties only in Applets sections
		if inAppletSection {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				switch key {
				case "plugin":
					currentWidget.Plugin = value
					currentWidget.IsCustom = !strings.HasPrefix(value, "org.kde.")
				case "position":
					currentWidget.Position = value
				case "size":
					currentWidget.Size = value
				case "config":
					currentWidget.Config = value
				}
			}
		}
	}

	// Save last widget if any
	if currentWidget.Plugin != "" {
		if isPanel {
			cfg.PanelWidgets = append(cfg.PanelWidgets, currentWidget)
		} else {
			cfg.DesktopWidgets = append(cfg.DesktopWidgets, currentWidget)
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

// GetCustomWidgets returns only custom (non-system) widgets from a widget config
func GetCustomWidgets(cfg *WidgetConfig) []WidgetInfo {
	var custom []WidgetInfo
	
	for _, w := range cfg.DesktopWidgets {
		if w.IsCustom {
			custom = append(custom, w)
		}
	}
	
	for _, w := range cfg.PanelWidgets {
		if w.IsCustom && !containsWidget(custom, w.Plugin) {
			custom = append(custom, w)
		}
	}
	
	return custom
}

// containsWidget checks if a widget with given plugin name exists in the list
func containsWidget(widgets []WidgetInfo, plugin string) bool {
	for _, w := range widgets {
		if w.Plugin == plugin {
			return true
		}
	}
	return false
}

// InstallWidget installs a widget from a .plasmoid package file using kpackagetool6
func InstallWidget(packagePath, dataDir string) error {
	// Use kpackagetool6 for KDE Plasma 6+
	cmd := exec.Command("kpackagetool6", "--install", packagePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		// Fallback to kpackagetool5 for older versions
		cmd = exec.Command("kpackagetool5", "--install", packagePath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install widget %s: %w", packagePath, err)
		}
	}
	
	return nil
}

// InstallWidgetsFromBackup installs all custom widgets from backup
func InstallWidgetsFromBackup(backupDir, dataDir string, dryRun bool) ([]string, error) {
	var installed []string
	var errors []string
	
	widgetsDir := filepath.Join(backupDir, "widgets", "plasma", "plasmoids")
	
	// Check if widgets directory exists
	if _, err := os.Stat(widgetsDir); os.IsNotExist(err) {
		return installed, nil // No widgets to install
	}
	
	// Read widget directories
	entries, err := os.ReadDir(widgetsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup widgets directory: %w", err)
	}
	
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		
		widgetName := entry.Name()
		widgetPath := filepath.Join(widgetsDir, widgetName)
		
		// Check if already installed
		installedWidgets, err := ListInstalledWidgets(dataDir)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Failed to check installed widgets: %v", err))
			continue
		}
		
		isInstalled := false
		for _, w := range installedWidgets {
			if w.Plugin == widgetName {
				isInstalled = true
				break
			}
		}
		
		if isInstalled {
			fmt.Printf("Widget %s is already installed, skipping...\n", widgetName)
			continue
		}
		
		if dryRun {
			fmt.Printf("[DRY RUN] Would install widget: %s\n", widgetName)
			installed = append(installed, widgetName)
			continue
		}
		
		// Try to install using kpackagetool6 if there's a .plasmoid package
		packageFile := filepath.Join(widgetPath, "contents", "code", "main.qml")
		if _, err := os.Stat(packageFile); err == nil {
			// This is an unpacked widget, use kpackagetool6 to install directly from directory
			if err := InstallWidget(widgetPath, dataDir); err != nil {
				errors = append(errors, fmt.Sprintf("Failed to install widget %s: %v", widgetName, err))
				continue
			}
			installed = append(installed, widgetName)
			fmt.Printf("Installed widget: %s\n", widgetName)
		} else {
			// Look for .plasmoid package file
			plasmoidFile := filepath.Join(widgetPath, widgetName+".plasmoid")
			if _, err := os.Stat(plasmoidFile); err == nil {
				if err := InstallWidget(plasmoidFile, dataDir); err != nil {
					errors = append(errors, fmt.Sprintf("Failed to install widget %s: %v", widgetName, err))
					continue
				}
				installed = append(installed, widgetName)
			} else {
				// Direct copy as fallback (kpackagetool6 can also work with unpacked directories)
				if err := InstallWidget(widgetPath, dataDir); err != nil {
					errors = append(errors, fmt.Sprintf("Failed to install widget %s: %v", widgetName, err))
					continue
				}
				installed = append(installed, widgetName)
				fmt.Printf("Installed widget (from directory): %s\n", widgetName)
			}
		}
	}
	
	if len(errors) > 0 {
		return installed, fmt.Errorf("errors during installation: %v", errors)
	}
	
	return installed, nil
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
