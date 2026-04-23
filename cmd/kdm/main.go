package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/user/kde-dotfiles-manager/internal/config"
)

// Main entry point for the KDE Dotfiles Manager TUI
func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load config: %v\n", err)
		cfg = config.DefaultConfig()
	}

	delegate := list.NewDefaultDelegate()

	l := list.New(mainMenuItems(), delegate, 0, 0)
	l.Title = "KDE Dotfiles Manager"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(false)

	m := &model{
		list:   l,
		cfg:    cfg,
		width:  80,
		height: 40,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}

// model is the main TUI model
type model struct {
	list     list.Model
	cfg      *config.Config
	width    int
	height   int
	quitting bool
}

// Init initializes the model
func (m model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			return m.handleSelection()
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View renders the UI
func (m model) View() string {
	if m.quitting {
		return "\n  Thank you for using KDE Dotfiles Manager!\n\n"
	}

	return "\n" + m.list.View() + "\n\n  Press q to quit • Enter to select • ↑↓ to navigate\n"
}

func (m model) handleSelection() (tea.Model, tea.Cmd) {
	if item, ok := m.list.SelectedItem().(mainItem); ok {
		switch item.id {
		case "backup":
			return newBackupScreen(&m), nil
		case "restore":
			return newRestoreScreen(&m), nil
		case "sync":
			return newSyncScreen(&m), nil
		case "deploy":
			return newDeployScreen(&m), nil
		case "config":
			return newConfigScreen(&m), nil
		case "status":
			return newStatusScreen(&m), nil
		case "profile":
			return newProfileScreen(&m), nil
		case "quit":
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

// mainItem represents a menu item in the main menu
type mainItem struct {
	id          string
	title       string
	description string
}

func (i mainItem) Title() string       { return i.title }
func (i mainItem) Description() string { return i.description }
func (i mainItem) FilterValue() string { return i.title }

// mainMenuItems returns the list of main menu items
func mainMenuItems() []list.Item {
	return []list.Item{
		mainItem{
			id:          "backup",
			title:       "📦 Backup",
			description: "Save KDE configurations to dotfiles",
		},
		mainItem{
			id:          "restore",
			title:       "📂 Restore",
			description: "Restore KDE configurations from dotfiles",
		},
		mainItem{
			id:          "sync",
			title:       "🔄 Sync",
			description: "Synchronize dotfiles with Git repository",
		},
		mainItem{
			id:          "deploy",
			title:       "🚀 Deploy",
			description: "Deploy saved configurations to current system",
		},
		mainItem{
			id:          "profile",
			title:       "👤 Profile",
			description: "Switch between configuration profiles",
		},
		mainItem{
			id:          "config",
			title:       "⚙️  Settings",
			description: "Configure backup options and preferences",
		},
		mainItem{
			id:          "status",
			title:       "📊 Status",
			description: "View current configuration status",
		},
		mainItem{
			id:          "quit",
			title:       "Quit",
			description: "Exit the application",
		},
	}
}

// Style definitions for the TUI
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7C5CFC")).
			MarginTop(1).
			MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			MarginBottom(1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#7C5CFC")).
			Padding(0, 1)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#CCCCCC"))

	checkStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000"))

	warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFF00"))

	buttonStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#7C5CFC")).
			Padding(0, 2).
			MarginRight(1)

	buttonInactiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#888888")).
				Background(lipgloss.Color("#333333")).
				Padding(0, 2).
				MarginRight(1)

	inputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#444444")).
			Padding(0, 1)
)
