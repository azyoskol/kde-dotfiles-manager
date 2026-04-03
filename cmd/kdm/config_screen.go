package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/user/kde-dotfiles-manager/internal/config"
)

// configScreen handles application settings
type configScreen struct {
	parent      *model
	cfg         *config.Config
	settings    []configSetting
	cursor      int
	message     string
	messageType string
	width       int
	height      int
}

// configSetting represents a configurable option
type configSetting struct {
	key         string
	label       string
	value       string
	description string
	toggleable  bool
}

func newConfigScreen(parent *model) *configScreen {
	s := &configScreen{
		parent: parent,
		cfg:    parent.cfg,
		cursor: 0,
	}

	s.settings = []configSetting{
		{
			key:         "dotfiles_dir",
			label:       "Dotfiles Directory",
			value:       parent.cfg.DotfilesDir,
			description: "Directory where dotfiles backups are stored",
			toggleable:  false,
		},
		{
			key:         "git_repo",
			label:       "Git Repository URL",
			value:       parent.cfg.GitRepo,
			description: "Remote git repository for synchronization",
			toggleable:  false,
		},
		{
			key:         "backup_before_restore",
			label:       "Backup Before Restore",
			value:       boolToString(parent.cfg.BackupBeforeRestore),
			description: "Create a backup before restoring configurations",
			toggleable:  true,
		},
		{
			key:         "verbose",
			label:       "Verbose Output",
			value:       boolToString(parent.cfg.Verbose),
			description: "Show detailed output during operations",
			toggleable:  true,
		},
	}

	return s
}

func (s *configScreen) Init() tea.Cmd {
	return nil
}

func (s *configScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if s.cursor < len(s.settings)-1 {
				s.cursor++
			}
		case " ", "enter":
			if s.settings[s.cursor].toggleable {
				s.toggleSetting(s.cursor)
			}
		case "s":
			return s.saveConfig()
		}
	}
	return s, nil
}

func (s *configScreen) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Configuration Settings"))
	b.WriteString("\n\n")
	b.WriteString(subtitleStyle.Render("Press Space to toggle, S to save"))
	b.WriteString("\n\n")

	for i, setting := range s.settings {
		cursor := "  "
		if i == s.cursor {
			cursor = "> "
		}

		var valueDisplay string
		if setting.toggleable {
			if setting.value == "true" {
				valueDisplay = checkStyle.Render("[ON]")
			} else {
				valueDisplay = errorStyle.Render("[OFF]")
			}
		} else {
			valueDisplay = setting.value
			if valueDisplay == "" {
				valueDisplay = warningStyle.Render("(not set)")
			}
		}

		line := fmt.Sprintf("%s%s: %s - %s", cursor, setting.label, valueDisplay, setting.description)
		if i == s.cursor {
			b.WriteString(selectedStyle.Render(line))
		} else {
			b.WriteString(normalStyle.Render(line))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n\n")

	// Save button
	saveLabel := "[ Save Configuration ]"
	b.WriteString(buttonStyle.Render(saveLabel))

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

	b.WriteString("\n\n  Press esc to go back • ↑↓ to navigate • Space to toggle • S to save")

	return b.String()
}

// toggleSetting toggles a boolean setting
func (s *configScreen) toggleSetting(index int) {
	setting := &s.settings[index]
	if setting.value == "true" {
		setting.value = "false"
	} else {
		setting.value = "true"
	}

	// Update the config struct
	switch setting.key {
	case "backup_before_restore":
		s.cfg.BackupBeforeRestore = setting.value == "true"
	case "verbose":
		s.cfg.Verbose = setting.value == "true"
	}
}

// saveConfig writes the configuration to disk
func (s *configScreen) saveConfig() (tea.Model, tea.Cmd) {
	if err := s.cfg.Save(); err != nil {
		s.message = fmt.Sprintf("Failed to save config: %v", err)
		s.messageType = "error"
	} else {
		s.message = "Configuration saved successfully"
		s.messageType = "success"
	}

	return s, nil
}

// boolToString converts a boolean to a string
func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
