package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/user/kde-dotfiles-manager/internal/config"
	"github.com/user/kde-dotfiles-manager/internal/fileutil"
	"github.com/user/kde-dotfiles-manager/internal/kde"
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
// Strategy: For each source path, preserve its full relative structure from config/data dir
func (m *Manager) Backup(categories []string) error {
	dotfilesDir := m.cfg.GetProfileDotfilesDir()

	// Create base directory
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

// backupCategory backs up all source paths for a single category
// Each source is copied preserving its relative path from config/data directory
func (m *Manager) backupCategory(name string, srcPaths map[string]string, destDir string) error {
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	for _, srcPath := range srcPaths {
		// Check if source exists
		info, err := os.Lstat(srcPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue // Skip non-existent files
			}
			return fmt.Errorf("failed to stat %s: %w", srcPath, err)
		}

		// Calculate relative path from config/data dir
		relPath := m.getRelativePath(srcPath)
		
		// Build destination path: destDir + relative path
		destPath := filepath.Join(destDir, relPath)

		// Create parent directories
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", destPath, err)
		}

		// Determine if it's a directory (including symlinks to directories)
		isDir := info.IsDir()
		if !isDir && info.Mode()&os.ModeSymlink != 0 {
			targetInfo, err := os.Stat(srcPath)
			if err == nil && targetInfo.IsDir() {
				isDir = true
			}
		}

		// Copy based on type
		if isDir {
			if err := fileutil.CopyDir(srcPath, destPath); err != nil {
				return fmt.Errorf("failed to copy directory %s: %w", srcPath, err)
			}
		} else {
			if err := fileutil.CopyFile(srcPath, destPath); err != nil {
				return fmt.Errorf("failed to copy file %s: %w", srcPath, err)
			}
		}
	}

	return nil
}

// getRelativePath returns the path relative to config or data directory
// This preserves the full directory structure for proper restoration
func (m *Manager) getRelativePath(path string) string {
	configDir := m.kdePaths.ConfigDir
	dataDir := m.kdePaths.DataDir

	if strings.HasPrefix(path, configDir+"/") {
		return strings.TrimPrefix(path, configDir+"/")
	}
	if strings.HasPrefix(path, dataDir+"/") {
		return strings.TrimPrefix(path, dataDir+"/")
	}
	// Fallback: just use the base name
	return filepath.Base(path)
}

// Restore restores configurations from backup
// Strategy: Walk through backup directory and restore each file to its original location
func (m *Manager) Restore(profile string) error {
	// Calculate profile path
	baseDir := m.cfg.ExpandPath()
	profilePath := filepath.Join(baseDir, "profiles", profile)

	// Check if profile directory exists
	if _, err := os.Stat(profilePath); os.IsNotExist(err) {
		return fmt.Errorf("profile '%s' does not exist", profile)
	}

	// Get all available categories
	categories := []string{"shortcuts", "themes", "window_management", "languages", "widgets", "panels", "system_settings"}

	for _, category := range categories {
		backupDir := filepath.Join(profilePath, category)
		if _, err := os.Stat(backupDir); os.IsNotExist(err) {
			continue
		}

		if err := m.restoreCategory(category, backupDir); err != nil {
			return fmt.Errorf("failed to restore %s: %w", category, err)
		}
	}

	return nil
}

// restoreCategory restores all files from backup directory to their original locations
// It walks through the backup directory and uses the relative path to determine destination
func (m *Manager) restoreCategory(name, backupDir string) error {
	// Walk through all files in backup directory
	return filepath.Walk(backupDir, func(srcPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the root backup directory itself
		if srcPath == backupDir {
			return nil
		}

		// Calculate relative path within backup
		relPath, err := filepath.Rel(backupDir, srcPath)
		if err != nil {
			return fmt.Errorf("failed to calculate relative path: %w", err)
		}

		// Determine destination based on category and relative path
		destPath := m.getDestinationForRestore(name, relPath)
		if destPath == "" {
			return nil // Skip if we can't determine destination
		}

		// Skip if already processed (for symlinks that might be visited twice)
		if _, err := os.Lstat(destPath); err == nil {
			// Destination exists, will be overwritten
		}

		// Create parent directories
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", destPath, err)
		}

		// Remove existing destination if it exists (to handle type changes)
		if info, err := os.Lstat(destPath); err == nil {
			if info.IsDir() && info.Mode()&os.ModeSymlink == 0 {
				if err := os.RemoveAll(destPath); err != nil {
					return fmt.Errorf("failed to remove existing directory %s: %w", destPath, err)
				}
			} else {
				if err := os.Remove(destPath); err != nil {
					return fmt.Errorf("failed to remove existing file %s: %w", destPath, err)
				}
			}
		}

		// Determine if source is a directory
		isDir := info.IsDir()
		if !isDir && info.Mode()&os.ModeSymlink != 0 {
			targetInfo, err := os.Stat(srcPath)
			if err == nil && targetInfo.IsDir() {
				isDir = true
			}
		}

		// Copy based on type
		if isDir {
			if err := fileutil.CopyDir(srcPath, destPath); err != nil {
				return fmt.Errorf("failed to restore directory %s: %w", destPath, err)
			}
		} else {
			if err := fileutil.CopyFile(srcPath, destPath); err != nil {
				return fmt.Errorf("failed to restore file %s: %w", destPath, err)
			}
		}

		return nil
	})
}

// getDestinationForRestore determines the destination path for a given category and relative path
func (m *Manager) getDestinationForRestore(category, relPath string) string {
	var basePath string
	
	switch category {
	case "shortcuts":
		basePath = m.getFirstExistingPath(m.kdePaths.ShortcutPaths())
	case "themes":
		basePath = m.getFirstExistingPath(m.kdePaths.ThemePaths())
	case "window_management":
		basePath = m.getFirstExistingPath(m.kdePaths.KWinPaths())
	case "languages":
		basePath = m.getFirstExistingPath(m.kdePaths.LocalePaths())
	case "widgets":
		basePath = m.getFirstExistingPath(m.kdePaths.WidgetPaths())
	case "panels":
		basePath = m.getFirstExistingPath(m.kdePaths.PanelPaths())
	case "system_settings":
		basePath = m.getFirstExistingPath(m.kdePaths.SystemSettingsPaths())
	default:
		return ""
	}

	if basePath == "" {
		return ""
	}

	// Reconstruct the full destination path
	// The relative path should be appended to the base directory of the first path
	baseDir := filepath.Dir(basePath)
	return filepath.Join(baseDir, relPath)
}

// getFirstExistingPath returns the first existing path from a map of paths
func (m *Manager) getFirstExistingPath(paths map[string]string) string {
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	// If none exist, return the first one anyway (it will be created)
	for _, path := range paths {
		return path
	}
	return ""
}

// GetBackupSize calculates the total size of a backup profile in bytes
func (m *Manager) GetBackupSize(profile string) (uint64, error) {
	baseDir := m.cfg.ExpandPath()
	profilePath := filepath.Join(baseDir, "profiles", profile)
	
	// Check if profile directory exists
	if _, err := os.Stat(profilePath); os.IsNotExist(err) {
		return 0, fmt.Errorf("profile '%s' does not exist", profile)
	}
	
	totalSize, err := fileutil.CalculateSize(profilePath)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate backup size: %w", err)
	}
	
	return totalSize, nil
}

// FormatSize formats bytes into human-readable string
func FormatSize(bytes uint64) string {
	return fileutil.FormatSize(bytes)
}
