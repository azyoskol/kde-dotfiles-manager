package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/user/kde-dotfiles-manager/internal/config"
)

// deployScreen handles configuration deployment
type deployScreen struct {
	parent      *model
	cfg         *config.Config
	profiles    []string
	cursor      int
	message     string
	messageType string
	width       int
	height      int
}

func newDeployScreen(parent *model) *deployScreen {
	s := &deployScreen{
		parent:   parent,
		cfg:      parent.cfg,
		cursor:   0,
		profiles: s.discoverProfiles(),
	}

	if len(s.profiles) == 0 {
		s.profiles = []string{"No profiles found"}
	}

	return s
}

func (s *deployScreen) Init() tea.Cmd {
	return nil
}

func (s *deployScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if s.profiles[s.cursor] != "No profiles found" {
				return s.executeDeploy()
			}
		}
	}
	return s, nil
}

func (s *deployScreen) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Deploy Configuration"))
	b.WriteString("\n\n")
	b.WriteString(subtitleStyle.Render("Select a profile to deploy to current system"))
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

	b.WriteString("\n\n  Press esc to go back • ↑↓ to navigate • Enter to deploy")

	return b.String()
}

// discoverProfiles finds available deployment profiles
func (s *deployScreen) discoverProfiles() []string {
	// In a real implementation, this would scan the dotfiles directory
	// For now, return placeholder profiles
	return []string{"default", "work", "gaming", "minimal"}
}

// executeDeploy runs the deployment process
func (s *deployScreen) executeDeploy() (tea.Model, tea.Cmd) {
	profile := s.profiles[s.cursor]

	s.message = fmt.Sprintf("Deploying profile '%s'...\n\nThis would execute: scripts/deploy.sh --profile %s", profile, profile)
	s.messageType = "success"

	return s, nil
}
