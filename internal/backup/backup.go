package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/user/kde-dotfiles-manager/internal/config"
	"github.com/user/kde-dotfiles-manager/internal/kde"
	"github.com/user/kde-dotfiles-manager/internal/widgets"
)

// Manager handles backup and restore operations for KDE configurations
type Manager struct {
	cfg      *config.Config
	kdePaths *kde.Paths
}

// NewManager creates a new backup manager
func NewManager(cfg *config.Config) (*Manager, error) {
	paths, err := kde.NewPaths()
	if err != nil {
		paths = &kde.Paths{}
	}
	return &Manager{
		cfg:      cfg,
		kdePaths: paths,
	}, nil
}

// Backup executes backup for specified categories
func (m *Manager) Backup(categories []string) error {
	dotfilesDir := m.cfg.GetProfileDotfilesDir()

	// Create base directory structure
	if err := os.MkdirAll(dotfilesDir, 0755); err != nil {
		return fmt.Errorf("failed to create dotfiles directory: %w", err)
	}

	for _, category := range categories {
		var srcPaths map[string]string
		var destDir string

		switch category {
		case "shortcuts":
			srcPaths = m.kdePaths.ShortcutPaths()
			destDir = filepath.Join(dotfilesDir, "shortcuts")
		case "themes":
			srcPaths = m.kdePaths.ThemePaths()
			destDir = filepath.Join(dotfilesDir, "themes")
		case "window_management":
			srcPaths = m.kdePaths.KWinPaths()
			destDir = filepath.Join(dotfilesDir, "window_management")
		case "languages":
			srcPaths = m.kdePaths.LocalePaths()
			destDir = filepath.Join(dotfilesDir, "languages")
		case "widgets":
			srcPaths = m.kdePaths.WidgetPaths()
			destDir = filepath.Join(dotfilesDir, "widgets")
		case "panels":
			srcPaths = m.kdePaths.PanelPaths()
			destDir = filepath.Join(dotfilesDir, "panels")
		case "system_settings":
			srcPaths = m.kdePaths.SystemSettingsPaths()
			destDir = filepath.Join(dotfilesDir, "system_settings")
		default:
			continue
		}

		if err := m.backupCategory(category, srcPaths, destDir); err != nil {
			return fmt.Errorf("failed to backup %s: %w", category, err)
		}
	}

	return nil
}

// backupCategory backs up files for a single category
func (m *Manager) backupCategory(name string, srcPaths map[string]string, destDir string) error {
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

		// Determine destination path
		var destPath string
		if info.IsDir() {
			destPath = filepath.Join(destDir, filepath.Base(srcPath))
		} else {
			// For files, preserve directory structure relative to config/data dir
			relPath := m.getRelativePath(srcPath)
			destPath = filepath.Join(destDir, relPath)
		}

		// Create parent directories
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", name, err)
		}

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

// getRelativePath returns the path relative to config or data directory
func (m *Manager) getRelativePath(path string) string {
	configDir := m.kdePaths.ConfigDir
	dataDir := m.kdePaths.DataDir

	if strings.HasPrefix(path, configDir) {
		return strings.TrimPrefix(path, configDir+"/")
	}
	if strings.HasPrefix(path, dataDir) {
		return strings.TrimPrefix(path, dataDir+"/")
	}
	return filepath.Base(path)
}

// Restore restores configurations from backup
func (m *Manager) Restore(profile string) error {
	dotfilesDir := m.cfg.GetProfileDotfilesDir()

	// Check if profile directory exists
	if _, err := os.Stat(dotfilesDir); os.IsNotExist(err) {
		return fmt.Errorf("profile '%s' does not exist", profile)
	}

	// Get all available categories from backup
	categories := []string{"shortcuts", "themes", "window_management", "languages", "widgets", "panels", "system_settings"}

	for _, category := range categories {
		backupDir := filepath.Join(dotfilesDir, category)
		if _, err := os.Stat(backupDir); os.IsNotExist(err) {
			continue
		}

		var destPaths map[string]string
		switch category {
		case "shortcuts":
			destPaths = m.kdePaths.ShortcutPaths()
		case "themes":
			destPaths = m.kdePaths.ThemePaths()
		case "window_management":
			destPaths = m.kdePaths.KWinPaths()
		case "languages":
			destPaths = m.kdePaths.LocalePaths()
		case "widgets":
			destPaths = m.kdePaths.WidgetPaths()
		case "panels":
			destPaths = m.kdePaths.PanelPaths()
		case "system_settings":
			destPaths = m.kdePaths.SystemSettingsPaths()
		default:
			continue
		}

		if err := m.restoreCategory(category, backupDir, destPaths); err != nil {
			return fmt.Errorf("failed to restore %s: %w", category, err)
		}
	}

	return nil
}

// restoreCategory restores files for a single category
func (m *Manager) restoreCategory(name, backupDir string, destPaths map[string]string) error {
	for name, destPath := range destPaths {
		// Determine source path in backup
		relPath := m.getRelativePath(destPath)
		srcFile := filepath.Join(backupDir, relPath)

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

// copyFile copies a single file or symbolic link
func copyFile(src, dst string) error {
	// Use Lstat to not follow symlinks
	info, err := os.Lstat(src)
	if err != nil {
		return err
	}

	// Handle symbolic links
	if info.Mode()&os.ModeSymlink != 0 {
		linkTarget, err := os.Readlink(src)
		if err != nil {
			return fmt.Errorf("failed to read symlink %s: %w", src, err)
		}
		// Create parent directory if needed
		if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
			return fmt.Errorf("failed to create directory for symlink %s: %w", dst, err)
		}
		if err := os.Symlink(linkTarget, dst); err != nil {
			return fmt.Errorf("failed to create symlink %s: %w", dst, err)
		}
		return nil
	}

	// Regular file
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	// Create parent directory if needed
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("failed to create directory for file %s: %w", dst, err)
	}
	return os.WriteFile(dst, data, info.Mode())
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

		// Get full file info to check for symlinks
		info, err := os.Lstat(srcPath)
		if err != nil {
			return fmt.Errorf("failed to stat %s: %w", srcPath, err)
		}

		// Handle symbolic links
		if info.Mode()&os.ModeSymlink != 0 {
			linkTarget, err := os.Readlink(srcPath)
			if err != nil {
				return fmt.Errorf("failed to read symlink %s: %w", srcPath, err)
			}
			if err := os.Symlink(linkTarget, dstPath); err != nil {
				return fmt.Errorf("failed to create symlink %s: %w", dstPath, err)
			}
			continue
		}

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

// GetCustomWidgetsFromBackup finds custom widgets in backup that are not installed
func (m *Manager) GetCustomWidgetsFromBackup() ([]string, error) {
	dotfilesDir := m.cfg.GetProfileDotfilesDir()
	widgetsBackupPath := filepath.Join(dotfilesDir, "widgets", "plasma", "plasmoids")

	// Check if widgets directory exists
	if _, err := os.Stat(widgetsBackupPath); os.IsNotExist(err) {
		return nil, nil
	}

	// Get list of installed widgets
	installed, err := widgets.ListInstalledWidgets(m.kdePaths.DataDir)
	if err != nil {
		return nil, err
	}

	installedMap := make(map[string]bool)
	for _, w := range installed {
		installedMap[w.Plugin] = true
	}

	// Find widgets in backup that are not installed
	var toInstall []string
	entries, err := os.ReadDir(widgetsBackupPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		widgetName := entry.Name()
		// Skip system widgets
		if strings.HasPrefix(widgetName, "org.kde.") {
			continue
		}

		// Check if already installed
		if !installedMap[widgetName] {
			toInstall = append(toInstall, widgetName)
		}
	}

	return toInstall, nil
}

// InstallCustomWidgets installs custom widgets from backup
func (m *Manager) InstallCustomWidgets(dryRun bool) ([]string, error) {
	dotfilesDir := m.cfg.GetProfileDotfilesDir()
	return widgets.InstallWidgetsFromBackup(dotfilesDir, m.kdePaths.DataDir, dryRun)
}
