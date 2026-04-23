package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

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

// Backup executes backup for specified categories using goroutines for parallel processing
func (m *Manager) Backup(categories []string) error {
	dotfilesDir := m.cfg.GetProfileDotfilesDir()

	// Create base directory structure
	if err := os.MkdirAll(dotfilesDir, 0755); err != nil {
		return fmt.Errorf("failed to create dotfiles directory: %w", err)
	}

	// Define workItem type before using it
	type workItem struct {
		src  string
		dst  string
	}

	// Collect all work items - a file may need to be copied to multiple categories
	var workItems []workItem

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

		// Add all source paths for this category
		for _, srcPath := range srcPaths {
			// Determine destination path within category
			info, err := os.Lstat(srcPath)
			if err != nil && !os.IsNotExist(err) {
				continue
			}
			
			var relPath string
			if err == nil && info.IsDir() {
				relPath = filepath.Base(srcPath)
			} else {
				relPath = m.getRelativePath(srcPath)
			}
			
			destPath := filepath.Join(destDir, relPath)
			
			// Add work item - same source file may be copied to multiple destinations
			workItems = append(workItems, workItem{src: srcPath, dst: destPath})
		}
	}

	// Use WaitGroup to wait for all goroutines to complete
	var wg sync.WaitGroup
	errorChan := make(chan error, len(workItems)+len(categories)) // Sufficient buffer

	// Create a mutex for synchronizing directory creation
	var mkdirMu sync.Mutex

	// Create goroutines for each unique source file
	for _, item := range workItems {
		wg.Add(1)
		go func(src, dst string) {
			defer wg.Done()
			
			info, err := os.Lstat(src)
			if err != nil {
				if os.IsNotExist(err) {
					return // File doesn't exist, skip silently
				}
				errorChan <- fmt.Errorf("failed to stat %s: %w", src, err)
				return
			}

			// Create parent directories with mutex to avoid race conditions
			mkdirMu.Lock()
			if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
				mkdirMu.Unlock()
				errorChan <- fmt.Errorf("failed to create directory for %s: %w", dst, err)
				return
			}
			mkdirMu.Unlock()

			if info.IsDir() {
				if err := fileutil.CopyDir(src, dst); err != nil {
					errorChan <- fmt.Errorf("failed to copy directory %s: %w", src, err)
				}
			} else {
				if err := fileutil.CopyFile(src, dst); err != nil {
					errorChan <- fmt.Errorf("failed to copy file %s: %w", src, err)
				}
			}
		}(item.src, item.dst)
	}

	// Start a separate goroutine to close the channel after all workers complete
	go func() {
		wg.Wait()
		close(errorChan)
	}()

	// Collect any errors - this blocks until all goroutines complete and channel is closed
	var errors []error
	for err := range errorChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		// Return the first error, but log all of them
		for _, err := range errors[1:] {
			fmt.Printf("Additional error: %v\n", err)
		}
		return errors[0]
	}

	return nil
}

// backupCategory backs up files for a single category
func (m *Manager) backupCategory(name string, srcPaths map[string]string, destDir string) error {
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	var errors []string
	for name, srcPath := range srcPaths {
		info, err := os.Stat(srcPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			errors = append(errors, fmt.Sprintf("failed to stat %s: %v", name, err))
			continue
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
			errors = append(errors, fmt.Sprintf("failed to create directory for %s: %v", name, err))
			continue
		}

		if info.IsDir() {
			if err := fileutil.CopyDir(srcPath, destPath); err != nil {
				errors = append(errors, fmt.Sprintf("failed to copy directory %s: %v", name, err))
				continue
			}
		} else {
			if err := fileutil.CopyFile(srcPath, destPath); err != nil {
				errors = append(errors, fmt.Sprintf("failed to copy file %s: %v", name, err))
				continue
			}
		}
	}

	if len(errors) > 0 {
		fmt.Printf("Warnings during backup of %s:\n", name)
		for _, err := range errors {
			fmt.Printf("  - %s\n", err)
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
	// Calculate the correct profile path
	baseDir := m.cfg.ExpandPath()
	profilePath := filepath.Join(baseDir, "profiles", profile)

	// Check if profile directory exists
	if _, err := os.Stat(profilePath); os.IsNotExist(err) {
		return fmt.Errorf("profile '%s' does not exist", profile)
	}

	// Get all available categories from backup
	categories := []string{"shortcuts", "themes", "window_management", "languages", "widgets", "panels", "system_settings"}

	for _, category := range categories {
		backupDir := filepath.Join(profilePath, category)
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
		if err := fileutil.EnsureDir(parent, 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", name, err)
		}

		if err := fileutil.CopyFile(srcFile, destPath); err != nil {
			return fmt.Errorf("failed to restore %s: %w", name, err)
		}
	}

	return nil
}

// copyFile copies a single file or symbolic link
// Deprecated: Use fileutil.CopyFile instead
func copyFile(src, dst string) error {
	return fileutil.CopyFile(src, dst)
}

// copyDir recursively copies a directory
// Deprecated: Use fileutil.CopyDir instead
func copyDir(src, dst string) error {
	return fileutil.CopyDir(src, dst)
}

// GetBackupSize calculates the total size of a backup profile in bytes
func (m *Manager) GetBackupSize(profile string) (uint64, error) {
	baseDir := m.cfg.ExpandPath()
	var profilePath string
	
	if profile == "default" {
		profilePath = filepath.Join(baseDir, "profiles", "default")
	} else {
		profilePath = filepath.Join(baseDir, "profiles", profile)
	}
	
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
// Deprecated: Use fileutil.FormatSize instead
func FormatSize(bytes uint64) string {
	return fileutil.FormatSize(bytes)
}
