// Package fileutil provides common file operations used across the application.
package fileutil

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// CopyFile copies a single file from src to dst, preserving permissions.
// It handles regular files and symbolic links correctly.
func CopyFile(src, dst string) error {
	// Use Lstat to not follow symlinks
	info, err := os.Lstat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source file: %w", err)
	}

	// Check if destination already exists
	if _, err := os.Lstat(dst); err == nil {
		// Destination exists, remove it first
		if err := os.RemoveAll(dst); err != nil {
			return fmt.Errorf("failed to remove existing destination %s: %w", dst, err)
		}
	}

	// Handle symbolic links
	if info.Mode()&os.ModeSymlink != 0 {
		return copySymlink(src, dst)
	}

	// Regular file
	return copyRegularFile(src, dst, info.Mode())
}

// copyRegularFile copies a regular file from src to dst with the given mode.
func copyRegularFile(src, dst string, mode os.FileMode) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// Create parent directory if needed
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("failed to create directory for file: %w", err)
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		dstFile.Close()
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	// Ensure all data is written to disk
	if err := dstFile.Sync(); err != nil {
		dstFile.Close()
		return fmt.Errorf("failed to sync file: %w", err)
	}

	if err := dstFile.Chmod(mode); err != nil {
		dstFile.Close()
		return fmt.Errorf("failed to set file permissions: %w", err)
	}

	if err := dstFile.Close(); err != nil {
		return fmt.Errorf("failed to close destination file: %w", err)
	}

	return nil
}

// copySymlink copies a symbolic link from src to dst.
func copySymlink(src, dst string) error {
	linkTarget, err := os.Readlink(src)
	if err != nil {
		return fmt.Errorf("failed to read symlink: %w", err)
	}

	// Create parent directory if needed
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("failed to create directory for symlink: %w", err)
	}

	if err := os.Symlink(linkTarget, dst); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	return nil
}

// CopyDir recursively copies a directory from src to dst.
func CopyDir(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("failed to read source directory: %w", err)
	}

	if err := os.MkdirAll(dst, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		// Get full file info to check for symlinks
		info, err := os.Lstat(srcPath)
		if err != nil {
			return fmt.Errorf("failed to stat %s: %w", srcPath, err)
		}

		// Check if destination already exists and remove it
		if _, err := os.Lstat(dstPath); err == nil {
			if err := os.RemoveAll(dstPath); err != nil {
				return fmt.Errorf("failed to remove existing destination %s: %w", dstPath, err)
			}
		}

		// Handle symbolic links
		if info.Mode()&os.ModeSymlink != 0 {
			if err := copySymlink(srcPath, dstPath); err != nil {
				return err
			}
			continue
		}

		if entry.IsDir() {
			if err := CopyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyRegularFile(srcPath, dstPath, info.Mode()); err != nil {
				return err
			}
		}
	}

	return nil
}

// EnsureDir creates a directory and all its parent directories if they don't exist.
func EnsureDir(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

// FileExists checks if a file or directory exists at the given path.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// DirExists checks if a directory exists at the given path.
func DirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// CalculateSize calculates the total size of all files in a directory.
// It excludes .git directories by default.
func CalculateSize(rootPath string) (uint64, error) {
	var totalSize uint64
	
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Skip .git directories
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}
		
		if !info.IsDir() {
			totalSize += uint64(info.Size())
		}
		
		return nil
	})
	
	if err != nil {
		return 0, fmt.Errorf("failed to calculate size: %w", err)
	}
	
	return totalSize, nil
}

// FormatSize formats bytes into a human-readable string.
func FormatSize(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
