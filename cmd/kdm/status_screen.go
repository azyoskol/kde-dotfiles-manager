package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/user/kde-dotfiles-manager/internal/config"
	"github.com/user/kde-dotfiles-manager/internal/kde"
	"github.com/user/kde-dotfiles-manager/internal/sync"
	"github.com/user/kde-dotfiles-manager/internal/backup"
)

// statusScreen shows the current status of configurations
type statusScreen struct {
	parent    *model
	cfg       *config.Config
	kdePaths  *kde.Paths
	gitSync   *sync.GitSync
	statuses  []statusItem
	width     int
	height    int
}

// statusItem represents a configuration status entry
type statusItem struct {
	category string
	status   string // "exists", "missing", "modified"
	details  string
}

func newStatusScreen(parent *model) *statusScreen {
	s := &statusScreen{
		parent:   parent,
		cfg:      parent.cfg,
		statuses: []statusItem{},
	}

	var err error
	s.kdePaths, err = kde.NewPaths()
	if err != nil {
		s.kdePaths = &kde.Paths{}
	}
	s.gitSync = sync.NewGitSync(parent.cfg.ExpandPath(), parent.cfg.GitRepo)
	s.checkStatuses()

	return s
}

func (s *statusScreen) Init() tea.Cmd {
	return nil
}

func (s *statusScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
		return s, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			return s.parent, nil
		case "r":
			s.checkStatuses()
		}
	}
	return s, nil
}

func (s *statusScreen) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Configuration Status"))
	b.WriteString("\n\n")

	// Git status
	gitStatus := "Git: Not initialized"
	if s.gitSync.IsGitRepo() {
		gitStatus = checkStyle.Render("Git: Active")
	}
	b.WriteString(subtitleStyle.Render(gitStatus))
	b.WriteString("\n\n")

	// Configuration statuses
	for _, item := range s.statuses {
		var statusIcon string
		switch item.status {
		case "exists":
			statusIcon = checkStyle.Render("[OK]")
		case "missing":
			statusIcon = errorStyle.Render("[MISSING]")
		case "modified":
			statusIcon = warningStyle.Render("[MODIFIED]")
		}

		line := fmt.Sprintf("  %s %s - %s", statusIcon, item.category, item.details)
		b.WriteString(line)
		b.WriteString("\n")
	}

	b.WriteString("\n\n")

	// Dotfiles directory status
	dotfilesDir := s.cfg.ExpandPath()
	if _, err := os.Stat(dotfilesDir); err == nil {
		b.WriteString(fmt.Sprintf("  Dotfiles directory: %s", dotfilesDir))
		b.WriteString("\n")

		// Count backup files and size
		count, totalSize := s.countBackupFiles(dotfilesDir)
		b.WriteString(fmt.Sprintf("  Backed up files: %d", count))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("  Total size: %s", backup.FormatSize(totalSize)))
	} else {
		b.WriteString(warningStyle.Render(fmt.Sprintf("  Dotfiles directory not found: %s", dotfilesDir)))
	}

	b.WriteString("\n\n  Press esc to go back • r to refresh")

	return b.String()
}

// checkStatuses checks the status of all KDE configuration files
func (s *statusScreen) checkStatuses() {
	s.statuses = []statusItem{}

	if s.kdePaths == nil {
		return
	}

	// Check shortcuts
	s.checkCategory("Shortcuts", s.kdePaths.ShortcutPaths())

	// Check window management
	s.checkCategory("Window Management", s.kdePaths.KWinPaths())

	// Check themes
	s.checkCategory("Themes", s.kdePaths.ThemePaths())

	// Check locales
	s.checkCategory("Languages", s.kdePaths.LocalePaths())

	// Check widgets
	s.checkCategory("Widgets", s.kdePaths.WidgetPaths())

	// Check panels
	s.checkCategory("Panels", s.kdePaths.PanelPaths())

	// Check system settings
	s.checkCategory("System Settings", s.kdePaths.SystemSettingsPaths())
}

// checkCategory checks all files in a category
func (s *statusScreen) checkCategory(category string, paths map[string]string) {
	exists := 0
	total := len(paths)

	for name, path := range paths {
		// Only check files (not directories)
		if !strings.Contains(path, "/") {
			continue
		}

		if _, err := os.Stat(path); err == nil {
			exists++
		}
		_ = name
	}

	var status string
	if exists == 0 {
		status = "missing"
	} else if exists == total {
		status = "exists"
	} else {
		status = "modified"
	}

	s.statuses = append(s.statuses, statusItem{
		category: category,
		status:   status,
		details:  fmt.Sprintf("%d/%d files found", exists, total),
	})
}

// countBackupFiles counts the number of files and total size in the dotfiles directory
func (s *statusScreen) countBackupFiles(dir string) (int, uint64) {
	count := 0
	var totalSize uint64
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && !strings.Contains(path, "/.git/") {
			count++
			totalSize += uint64(info.Size())
		}
		return nil
	})
	return count, totalSize
}
