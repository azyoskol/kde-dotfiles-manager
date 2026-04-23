package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/user/kde-dotfiles-manager/internal/config"
)

// profileScreen handles profile switching functionality
type profileScreen struct {
	parent      *model
	cfg         *config.Config
	profiles    []string
	cursor      int
	message     string
	messageType string
	width       int
	height      int
}

func newProfileScreen(parent *model) *profileScreen {
	s := &profileScreen{
		parent:   parent,
		cfg:      parent.cfg,
		cursor:   0,
	}

	// Load available profiles
	profiles, err := config.ListProfiles()
	if err != nil {
		s.message = fmt.Sprintf("Error loading profiles: %v", err)
		s.messageType = "error"
		profiles = []string{"default"}
	}
	s.profiles = profiles

	// Find current profile index
	for i, p := range s.profiles {
		if p == parent.cfg.Profile {
			s.cursor = i
			break
		}
	}

	return s
}

func (s *profileScreen) Init() tea.Cmd {
	return nil
}

func (s *profileScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			return s.switchProfile()
		}
	}
	return s, nil
}

func (s *profileScreen) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Switch Profile"))
	b.WriteString("\n\n")
	b.WriteString(subtitleStyle.Render("Select a profile to switch to"))
	b.WriteString("\n\n")

	for i, profile := range s.profiles {
		cursor := "  "
		if i == s.cursor {
			cursor = "> "
		}

		currentMarker := ""
		if profile == s.cfg.Profile {
			currentMarker = checkStyle.Render(" [current]")
		}

		line := fmt.Sprintf("%s%s%s", cursor, profile, currentMarker)
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

	b.WriteString("\n\n  Press esc to go back • ↑↓ to navigate • Enter to switch")

	return b.String()
}

// switchProfile changes to the selected profile
func (s *profileScreen) switchProfile() (tea.Model, tea.Cmd) {
	selectedProfile := s.profiles[s.cursor]

	if selectedProfile == s.cfg.Profile {
		s.message = fmt.Sprintf("Already using profile '%s'", selectedProfile)
		s.messageType = "info"
		return s, nil
	}

	// Update config with new profile
	if err := s.cfg.SetProfile(selectedProfile); err != nil {
		s.message = fmt.Sprintf("Failed to switch profile: %v", err)
		s.messageType = "error"
		return s, nil
	}

	s.message = fmt.Sprintf("Successfully switched to profile '%s'", selectedProfile)
	s.messageType = "success"

	// Update parent model config
	s.parent.cfg = s.cfg

	return s, nil
}
