package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/user/kde-dotfiles-manager/internal/config"
	"github.com/user/kde-dotfiles-manager/internal/kde"
	"github.com/user/kde-dotfiles-manager/internal/widgets"
	"github.com/user/kde-dotfiles-manager/internal/backup"
)

// restoreScreen handles the restore functionality
type restoreScreen struct {
	parent      *model
	cfg         *config.Config
	kdePaths    *kde.Paths
	backupMgr   *backup.Manager
	profiles    []string
	profileSizes map[string]string
	cursor      int
	selected    int
	message     string
	messageType string
	width       int
	height      int
	showWidgetInstall bool
	widgetsToInstall []string
}

func newRestoreScreen(parent *model) *restoreScreen {
	paths, err := kde.NewPaths()
	if err != nil {
		paths = &kde.Paths{}
	}
	
	backupMgr, err := backup.NewManager(parent.cfg)
	if err != nil {
		backupMgr = nil
	}
	
	s := &restoreScreen{
		parent:    parent,
		cfg:       parent.cfg,
		kdePaths:  paths,
		backupMgr: backupMgr,
		selected:  0,
		profileSizes: make(map[string]string),
	}

	// Discover available backup profiles
	s.profiles = s.discoverProfiles()
	if len(s.profiles) == 0 {
		s.profiles = []string{"No backups found"}
	} else {
		// Calculate sizes for each profile
		for _, profile := range s.profiles {
			if size, err := s.backupMgr.GetBackupSize(profile); err == nil {
				s.profileSizes[profile] = backup.FormatSize(size)
			} else {
				s.profileSizes[profile] = "Unknown"
			}
		}
	}

	return s
}

func (s *restoreScreen) Init() tea.Cmd {
	return nil
}

func (s *restoreScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
		return s, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			if s.showWidgetInstall {
				s.showWidgetInstall = false
				return s, nil
			}
			return s.parent, nil
		case "up", "k":
			if s.showWidgetInstall {
				if s.cursor > 0 {
					s.cursor--
				}
			} else {
				if s.cursor > 0 {
					s.cursor--
				}
			}
		case "down", "j":
			if s.showWidgetInstall {
				if s.cursor < len(s.widgetsToInstall)-1 {
					s.cursor++
				}
			} else {
				if s.cursor < len(s.profiles)-1 {
					s.cursor++
				}
			}
		case "enter":
			if s.showWidgetInstall {
				return s.installSelectedWidgets()
			}
			if s.profiles[s.cursor] != "No backups found" {
				return s.executeRestore()
			}
		case " ":
			if s.showWidgetInstall && s.cursor < len(s.widgetsToInstall) {
				// Toggle widget selection (optional feature)
			}
		}
	}
	return s, nil
}

func (s *restoreScreen) View() string {
	var b strings.Builder

	if s.showWidgetInstall {
		return s.viewWidgetInstall(&b)
	}

	b.WriteString(titleStyle.Render("Restore KDE Configurations"))
	b.WriteString("\n\n")
	b.WriteString(subtitleStyle.Render("Select a backup profile to restore"))
	b.WriteString("\n\n")

	for i, profile := range s.profiles {
		cursor := "  "
		if i == s.cursor {
			cursor = "> "
		}

		sizeStr := ""
		if size, ok := s.profileSizes[profile]; ok {
			sizeStr = fmt.Sprintf(" (%s)", size)
		}

		line := fmt.Sprintf("%s%s%s", cursor, profile, sizeStr)
		if i == s.cursor {
			b.WriteString(selectedStyle.Render(line))
		} else {
			b.WriteString(normalStyle.Render(line))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n\n")

	if s.message != "" {
		switch s.messageType {
		case "success":
			b.WriteString(checkStyle.Render(s.message))
		case "error":
			b.WriteString(errorStyle.Render(s.message))
		default:
			b.WriteString(warningStyle.Render(s.message))
		}
	}

	b.WriteString("\n\n  Press esc to go back • ↑↓ to navigate • Enter to restore")

	return b.String()
}

// viewWidgetInstall shows the widget installation screen
func (s *restoreScreen) viewWidgetInstall(b *strings.Builder) string {
	b.WriteString(titleStyle.Render("Install Custom Widgets"))
	b.WriteString("\n\n")
	b.WriteString(subtitleStyle.Render("Custom widgets found in backup. Install them?"))
	b.WriteString("\n\n")

	for i, widget := range s.widgetsToInstall {
		cursor := "  "
		if i == s.cursor {
			cursor = "> "
		}

		line := fmt.Sprintf("%s[ ] %s", cursor, widget)
		if i == s.cursor {
			b.WriteString(selectedStyle.Render(line))
		} else {
			b.WriteString(normalStyle.Render(line))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n\n")

	if s.message != "" {
		switch s.messageType {
		case "success":
			b.WriteString(checkStyle.Render(s.message))
		case "error":
			b.WriteString(errorStyle.Render(s.message))
		default:
			b.WriteString(warningStyle.Render(s.message))
		}
	}

	b.WriteString("\n\n  Press Enter to install selected • Esc to skip")

	return b.String()
}

// discoverProfiles finds available backup profiles in the dotfiles directory
func (s *restoreScreen) discoverProfiles() []string {
	baseDir := s.cfg.ExpandPath()
	profilesDir := filepath.Join(baseDir, "profiles")
	var profiles []string
	
	// Check profiles subdirectory
	if entries, err := os.ReadDir(profilesDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() && entry.Name() != ".git" {
				profiles = append(profiles, entry.Name())
			}
		}
	}
	
	if len(profiles) == 0 {
		return []string{"No backups found"}
	}
	
	// Ensure default is in the list
	hasDefault := false
	for _, p := range profiles {
		if p == "default" {
			hasDefault = true
			break
		}
	}
	
	if !hasDefault {
		profiles = append([]string{"default"}, profiles...)
	}
	
	return profiles
}

// executeRestore restores configurations using the Go manager
func (s *restoreScreen) executeRestore() (tea.Model, tea.Cmd) {
	profile := s.profiles[s.cursor]

	if s.backupMgr == nil {
		s.message = "Backup manager not initialized"
		s.messageType = "error"
		return s, nil
	}

	// Run restore using Go manager - it will handle profile path resolution
	err := s.backupMgr.Restore(profile)

	if err != nil {
		s.message = fmt.Sprintf("Restore failed: %v", err)
		s.messageType = "error"
		return s, nil
	}

	// After successful restore, get the profile path for widget check
	baseDir := s.cfg.ExpandPath()
	profilePath := filepath.Join(baseDir, "profiles", profile)

	// After successful restore, check for custom widgets to install
	s.message = fmt.Sprintf("Restore completed from profile: %s", profile)
	s.messageType = "success"
	
	// Check for custom widgets in backup
	widgetsToInstall, err := s.findCustomWidgets(profilePath)
	if err != nil {
		s.message = fmt.Sprintf("%s\nWarning: Could not check for widgets: %v", s.message, err)
	} else if len(widgetsToInstall) > 0 {
		s.widgetsToInstall = widgetsToInstall
		s.showWidgetInstall = true
		s.cursor = 0
		return s, nil
	}

	return s, nil
}

// findCustomWidgets finds custom widgets in the backup that are not installed
func (s *restoreScreen) findCustomWidgets(profilePath string) ([]string, error) {
	widgetsBackupPath := filepath.Join(profilePath, "widgets", "plasma", "plasmoids")
	
	// Check if widgets directory exists
	if _, err := os.Stat(widgetsBackupPath); os.IsNotExist(err) {
		return nil, nil // No widgets to install
	}
	
	// Get list of installed widgets
	installed, err := widgets.ListInstalledWidgets(s.kdePaths.DataDir)
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

// installSelectedWidgets installs the selected custom widgets
func (s *restoreScreen) installSelectedWidgets() (tea.Model, tea.Cmd) {
	profile := s.profiles[s.selected]
	dotfilesDir := s.cfg.GetProfileDotfilesDir()
	
	var profilePath string
	if s.cfg.Profile == "" || s.cfg.Profile == "default" {
		baseDir := s.cfg.ExpandPath()
		profilePath = filepath.Join(baseDir, profile)
	} else {
		profilePath = dotfilesDir
	}
	
	installed, err := widgets.InstallWidgetsFromBackup(profilePath, s.kdePaths.DataDir, false)
	
	if err != nil {
		s.message = fmt.Sprintf("Widget installation completed with errors: %v", err)
		s.messageType = "warning"
	} else {
		s.message = fmt.Sprintf("Successfully installed %d widgets", len(installed))
		s.messageType = "success"
	}
	
	s.showWidgetInstall = false
	s.widgetsToInstall = nil
	
	return s, nil
}
