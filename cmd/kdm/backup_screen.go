package main

import (
	"fmt"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/user/kde-dotfiles-manager/internal/config"
	"github.com/user/kde-dotfiles-manager/internal/kde"
)

// backupScreen handles the backup functionality
type backupScreen struct {
	parent     *model
	cfg        *config.Config
	kdePaths   *kde.Paths
	categories []categoryItem
	cursor     int
	selected   map[int]bool
	message    string
	messageType string // "success", "error", "info"
	width      int
	height     int
}

// categoryItem represents a backup category
type categoryItem struct {
	name        string
	description string
	enabled     bool
}

func newBackupScreen(parent *model) *backupScreen {
	paths, err := kde.NewPaths()
	if err != nil {
		paths = &kde.Paths{}
	}
	s := &backupScreen{
		parent:   parent,
		cfg:      parent.cfg,
		kdePaths: paths,
		selected: make(map[int]bool),
		cursor:   0,
	}

	// Initialize categories
	s.categories = []categoryItem{
		{name: "shortcuts", description: "Keyboard shortcuts and global hotkeys", enabled: true},
		{name: "themes", description: "Color schemes, icons, cursors, wallpapers", enabled: true},
		{name: "window_management", description: "KWin rules, virtual desktops, tiling", enabled: true},
		{name: "languages", description: "Locale, input methods, keyboard layouts", enabled: true},
		{name: "widgets", description: "Desktop widgets and plasmoids", enabled: true},
		{name: "panels", description: "Panel layout and configuration", enabled: true},
		{name: "system_settings", description: "General system settings", enabled: true},
		{name: "Execute Backup", description: "Start backup process for selected categories", enabled: true},
	}

	// Mark categories from config as selected
	for i, cat := range s.categories {
		for _, cfgCat := range parent.cfg.Categories {
			if cat.name == cfgCat {
				s.selected[i] = true
				break
			}
		}
	}

	return s
}

func (s *backupScreen) Init() tea.Cmd {
	return nil
}

func (s *backupScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if s.cursor < len(s.categories)-1 {
				s.cursor++
			}
		case " ", "enter":
			if s.cursor == len(s.categories)-1 {
				// Execute backup
				return s.executeBackup()
			}
			// Toggle selection
			s.selected[s.cursor] = !s.selected[s.cursor]
		}
	}
	return s, nil
}

func (s *backupScreen) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Backup KDE Configurations"))
	b.WriteString("\n\n")
	b.WriteString(subtitleStyle.Render("Select categories to backup (Space to toggle, Enter to start)"))
	b.WriteString("\n\n")

	// Category list
	execIndex := len(s.categories) - 1
	for i, cat := range s.categories {
		cursor := "  "
		if i == s.cursor {
			cursor = "> "
		}

		// Last item is the execute button, not a toggleable category
		if i == execIndex {
			line := fmt.Sprintf("%s[ Execute Backup ]", cursor)
			if i == s.cursor {
				b.WriteString(selectedStyle.Render(line))
			} else {
				b.WriteString(normalStyle.Render(line))
			}
			b.WriteString("\n")
			continue
		}

		checkbox := "[ ]"
		if s.selected[i] {
			checkbox = checkStyle.Render("[x]")
		}

		line := fmt.Sprintf("%s%s %s - %s", cursor, checkbox, cat.name, cat.description)
		if i == s.cursor {
			b.WriteString(selectedStyle.Render(line))
		} else {
			b.WriteString(normalStyle.Render(line))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Backup button
	buttonLabel := "[ Execute Backup ]"
	if s.cursor == len(s.categories) {
		b.WriteString(buttonStyle.Render(buttonLabel))
	} else {
		b.WriteString(buttonInactiveStyle.Render(buttonLabel))
	}

	b.WriteString("\n\n")

	// Message display
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

	b.WriteString("\n\n  Press esc to go back • ↑↓ to navigate • Space to toggle")

	return b.String()
}

// executeBackup runs the backup bash script
func (s *backupScreen) executeBackup() (tea.Model, tea.Cmd) {
	var selectedCats []string
	for i, cat := range s.categories {
		if s.selected[i] {
			selectedCats = append(selectedCats, cat.name)
		}
	}

	if len(selectedCats) == 0 {
		s.message = "No categories selected for backup"
		s.messageType = "error"
		return s, nil
	}

	// Build the backup command
	scriptPath := "scripts/backup.sh"
	args := []string{
		"--dotfiles-dir", s.cfg.ExpandPath(),
		"--categories", strings.Join(selectedCats, ","),
	}

	if s.cfg.Verbose {
		args = append(args, "--verbose")
	}

	cmd := exec.Command("bash", append([]string{scriptPath}, args...)...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		s.message = fmt.Sprintf("Backup failed: %s", string(output))
		s.messageType = "error"
	} else {
		s.message = fmt.Sprintf("Backup completed successfully for: %s", strings.Join(selectedCats, ", "))
		s.messageType = "success"
	}

	return s, nil
}
