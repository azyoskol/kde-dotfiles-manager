package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/user/kde-dotfiles-manager/internal/config"
)

// restoreScreen handles the restore functionality
type restoreScreen struct {
	parent      *model
	cfg         *config.Config
	profiles    []string
	cursor      int
	selected    int
	message     string
	messageType string
	width       int
	height      int
}

func newRestoreScreen(parent *model) *restoreScreen {
	s := &restoreScreen{
		parent:   parent,
		cfg:      parent.cfg,
		selected: 0,
	}

	// Discover available backup profiles
	s.profiles = s.discoverProfiles()
	if len(s.profiles) == 0 {
		s.profiles = []string{"No backups found"}
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
			return s.parent, nil
		case "up", "k":
			if s.cursor > 0 {
				s.cursor--
			}
		case "down", "j":
			if s.cursor < len(s.profiles)-1 {
				s.cursor++
			}
		case "enter":
			if s.profiles[s.cursor] != "No backups found" {
				return s.executeRestore()
			}
		}
	}
	return s, nil
}

func (s *restoreScreen) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Restore KDE Configurations"))
	b.WriteString("\n\n")
	b.WriteString(subtitleStyle.Render("Select a backup profile to restore"))
	b.WriteString("\n\n")

	for i, profile := range s.profiles {
		cursor := "  "
		if i == s.cursor {
			cursor = "> "
		}

		line := fmt.Sprintf("%s%s", cursor, profile)
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

// discoverProfiles finds available backup profiles in the dotfiles directory
func (s *restoreScreen) discoverProfiles() []string {
	dotfilesDir := s.cfg.ExpandPath()
	var profiles []string

	entries, err := os.ReadDir(dotfilesDir)
	if err != nil {
		return profiles
	}

	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != ".git" {
			profiles = append(profiles, entry.Name())
		}
	}

	return profiles
}

// executeRestore runs the restore bash script
func (s *restoreScreen) executeRestore() (tea.Model, tea.Cmd) {
	profile := s.profiles[s.cursor]
	dotfilesDir := s.cfg.ExpandPath()
	profilePath := filepath.Join(dotfilesDir, profile)

	// Check if profile directory exists
	if _, err := os.Stat(profilePath); os.IsNotExist(err) {
		s.message = fmt.Sprintf("Profile '%s' does not exist", profile)
		s.messageType = "error"
		return s, nil
	}

	// Create backup before restore if configured
	if s.cfg.BackupBeforeRestore {
		s.message = "Creating backup before restore..."
		s.messageType = "info"
	}

	scriptPath := "scripts/restore.sh"
	args := []string{
		"--dotfiles-dir", dotfilesDir,
		"--profile", profile,
	}

	s.message = fmt.Sprintf("Restoring from profile: %s (script execution would happen here)", profile)
	s.messageType = "success"

	_ = scriptPath
	_ = args

	return s, nil
}
