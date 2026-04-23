package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/user/kde-dotfiles-manager/internal/config"
	"github.com/user/kde-dotfiles-manager/internal/kde"
	"github.com/user/kde-dotfiles-manager/internal/backup"
)

// backupScreen handles the backup functionality
type backupScreen struct {
	parent     *model
	cfg        *config.Config
	kdePaths   *kde.Paths
	backupMgr  *backup.Manager
	categories []categoryItem
	cursor     int
	selected   map[int]bool
	message    string
	messageType string // "success", "error", "info"
	width      int
	height     int
	spinner    spinner.Model
	isBackingUp bool
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
	
	backupMgr, err := backup.NewManager(parent.cfg)
	if err != nil {
		backupMgr = nil
	}
	
	s := &backupScreen{
		parent:    parent,
		cfg:       parent.cfg,
		kdePaths:  paths,
		backupMgr: backupMgr,
		selected:  make(map[int]bool),
		cursor:    0,
		spinner:   spinner.New(),
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

	case spinner.TickMsg:
		if !s.isBackingUp {
			break
		}
		var cmd tea.Cmd
		s.spinner, cmd = s.spinner.Update(msg)
		return s, cmd

	case backupDoneMsg:
		s.isBackingUp = false
		if msg.err != nil {
			s.message = fmt.Sprintf("Backup failed: %v", msg.err)
			s.messageType = "error"
		} else {
			s.message = fmt.Sprintf("Backup completed successfully for: %s", strings.Join(msg.categories, ", "))
			s.messageType = "success"
		}
		return s, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			if s.isBackingUp {
				return s, nil // Don't allow exit during backup
			}
			return s.parent, nil
		case "up", "k":
			if s.isBackingUp {
				return s, nil // Don't allow navigation during backup
			}
			if s.cursor > 0 {
				s.cursor--
			}
		case "down", "j":
			if s.isBackingUp {
				return s, nil // Don't allow navigation during backup
			}
			if s.cursor < len(s.categories)-1 {
				s.cursor++
			}
		case " ", "enter":
			if s.isBackingUp {
				return s, nil // Don't allow actions during backup
			}
			if s.cursor == len(s.categories)-1 {
				// Execute backup
				return s.executeBackup()
			}
			// Toggle selection
			s.selected[s.cursor] = !s.selected[s.cursor]
		}
	}
	
	// Continue spinner animation during backup
	if s.isBackingUp {
		return s, s.spinner.Tick
	}
	
	return s, nil
}

func (s *backupScreen) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Backup KDE Configurations"))
	b.WriteString("\n\n")
	
	if s.isBackingUp {
		b.WriteString(s.spinner.View() + " Backup in progress...\n\n")
	} else {
		b.WriteString(subtitleStyle.Render("Select categories to backup (Space to toggle, Enter to start)"))
		b.WriteString("\n\n")
	}

	// Category list
	execIndex := len(s.categories) - 1
	for i, cat := range s.categories {
		cursor := "  "
		if i == s.cursor && !s.isBackingUp {
			cursor = "> "
		}

		// Last item is the execute button, not a toggleable category
		if i == execIndex {
			line := fmt.Sprintf("%s[ Execute Backup ]", cursor)
			if i == s.cursor && !s.isBackingUp {
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
		if i == s.cursor && !s.isBackingUp {
			b.WriteString(selectedStyle.Render(line))
		} else {
			b.WriteString(normalStyle.Render(line))
		}
		b.WriteString("\n")
	}

	if !s.isBackingUp {
		b.WriteString("\n")

		// Backup button
		buttonLabel := "[ Execute Backup ]"
		if s.cursor == len(s.categories) {
			b.WriteString(buttonStyle.Render(buttonLabel))
		} else {
			b.WriteString(buttonInactiveStyle.Render(buttonLabel))
		}

		b.WriteString("\n\n")
	}

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

	if !s.isBackingUp {
		b.WriteString("\n\n  Press esc to go back • ↑↓ to navigate • Space to toggle")
	}

	return b.String()
}

// backupDoneMsg is sent when backup operation completes
type backupDoneMsg struct {
	categories []string
	err        error
}

// executeBackup runs the backup using the Go manager
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

	if s.backupMgr == nil {
		s.message = "Backup manager not initialized"
		s.messageType = "error"
		return s, nil
	}

	// Set backing up state and start spinner
	s.isBackingUp = true
	s.spinner = spinner.New()
	s.spinner.Spinner = spinner.Dot
	s.message = "Creating backup..."
	s.messageType = "info"

	// Return command to run backup in background
	return s, func() tea.Msg {
		// Run backup in a separate goroutine to avoid blocking the UI
		err := s.backupMgr.Backup(selectedCats)
		return backupDoneMsg{categories: selectedCats, err: err}
	}
}
