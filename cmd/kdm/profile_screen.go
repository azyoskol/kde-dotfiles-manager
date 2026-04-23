package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/user/kde-dotfiles-manager/internal/config"
)

// profileScreen handles profile management functionality
type profileScreen struct {
	parent      *model
	cfg         *config.Config
	profiles    []string
	cursor      int
	message     string
	messageType string
	width       int
	height      int
	mode        string // "list", "create", "rename", "delete"
	input       string
}

func newProfileScreen(parent *model) *profileScreen {
	s := &profileScreen{
		parent:   parent,
		cfg:      parent.cfg,
		cursor:   0,
		mode:     "list",
	}

	// Load available profiles
	s.loadProfiles()

	// Find current profile index
	for i, p := range s.profiles {
		if p == parent.cfg.Profile {
			s.cursor = i
			break
		}
	}

	return s
}

func (s *profileScreen) loadProfiles() {
	profiles, err := config.ListProfiles()
	if err != nil {
		s.message = fmt.Sprintf("Error loading profiles: %v", err)
		s.messageType = "error"
		profiles = []string{"default"}
	}
	s.profiles = profiles
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
			if s.mode != "list" {
				s.mode = "list"
				s.input = ""
				s.message = ""
				return s, nil
			}
			return s.parent, nil

		case "enter":
			if s.mode == "list" {
				return s.switchProfile()
			} else if s.mode == "create" || s.mode == "rename" || s.mode == "delete" {
				return s.confirmAction()
			}

		case "n":
			if s.mode == "list" {
				s.mode = "create"
				s.input = ""
				s.message = ""
				return s, nil
			}

		case "r":
			if s.mode == "list" && len(s.profiles) > 0 {
				s.mode = "rename"
				s.input = ""
				s.message = ""
				return s, nil
			}

		case "d":
			if s.mode == "list" && len(s.profiles) > 1 {
				s.mode = "delete"
				s.message = ""
				return s, nil
			}

		case "up", "k":
			if s.mode == "list" && s.cursor > 0 {
				s.cursor--
			}
		case "down", "j":
			if s.mode == "list" && s.cursor < len(s.profiles)-1 {
				s.cursor++
			}

		default:
			if s.mode == "create" || s.mode == "rename" {
				s.input += msg.String()
				return s, nil
			}
		}

		// Handle backspace for input modes
		if s.mode == "create" || s.mode == "rename" {
			if msg.Type == tea.KeyBackspace || msg.Type == tea.KeyDelete {
				if len(s.input) > 0 {
					s.input = s.input[:len(s.input)-1]
				}
				return s, nil
			}
		}
	}
	return s, nil
}

func (s *profileScreen) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Profile Management"))
	b.WriteString("\n\n")

	switch s.mode {
	case "list":
		b.WriteString(subtitleStyle.Render("Manage your profiles"))
		b.WriteString("\n\n")
		b.WriteString("Actions: [n] Create  [r] Rename  [d] Delete\n\n")

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

	case "create":
		b.WriteString(subtitleStyle.Render("Create new profile"))
		b.WriteString("\n\n")
		b.WriteString("Enter profile name: ")
		b.WriteString(inputStyle.Render(s.input))
		b.WriteString("\n\n")
		b.WriteString("Press Enter to create, Esc to cancel")

	case "rename":
		b.WriteString(subtitleStyle.Render("Rename profile"))
		b.WriteString("\n\n")
		currentProfile := s.profiles[s.cursor]
		b.WriteString(fmt.Sprintf("Renaming '%s'\n\n", currentProfile))
		b.WriteString("Enter new name: ")
		b.WriteString(inputStyle.Render(s.input))
		b.WriteString("\n\n")
		b.WriteString("Press Enter to rename, Esc to cancel")

	case "delete":
		b.WriteString(subtitleStyle.Render("Delete profile"))
		b.WriteString("\n\n")
		currentProfile := s.profiles[s.cursor]
		if currentProfile == "default" {
			b.WriteString(errorStyle.Render("Cannot delete default profile!"))
		} else {
			b.WriteString(fmt.Sprintf("Are you sure you want to delete profile '%s'?\n", currentProfile))
			b.WriteString(warningStyle.Render("This action cannot be undone!"))
		}
		b.WriteString("\n\n")
		b.WriteString("Press Enter to confirm, Esc to cancel")
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
		b.WriteString("\n\n")
	}

	if s.mode == "list" {
		b.WriteString("Press esc to go back • ↑↓ to navigate • Enter to switch")
	}

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

// confirmAction handles confirmation for create/rename/delete actions
func (s *profileScreen) confirmAction() (tea.Model, tea.Cmd) {
	switch s.mode {
	case "create":
		if s.input == "" {
			s.message = "Profile name cannot be empty"
			s.messageType = "error"
			return s, nil
		}
		if err := config.CreateProfile(s.input); err != nil {
			s.message = fmt.Sprintf("Failed to create profile: %v", err)
			s.messageType = "error"
			return s, nil
		}
		s.loadProfiles()
		s.message = fmt.Sprintf("Profile '%s' created successfully", s.input)
		s.messageType = "success"
		s.mode = "list"
		s.input = ""

	case "rename":
		if s.input == "" {
			s.message = "Profile name cannot be empty"
			s.messageType = "error"
			return s, nil
		}
		oldName := s.profiles[s.cursor]
		if err := config.RenameProfile(oldName, s.input); err != nil {
			s.message = fmt.Sprintf("Failed to rename profile: %v", err)
			s.messageType = "error"
			return s, nil
		}
		s.loadProfiles()
		s.message = fmt.Sprintf("Profile renamed to '%s'", s.input)
		s.messageType = "success"
		s.mode = "list"
		s.input = ""

	case "delete":
		profileToDelete := s.profiles[s.cursor]
		if profileToDelete == "default" {
			s.message = "Cannot delete default profile"
			s.messageType = "error"
			return s, nil
		}
		if err := config.DeleteProfile(profileToDelete); err != nil {
			s.message = fmt.Sprintf("Failed to delete profile: %v", err)
			s.messageType = "error"
			return s, nil
		}
		s.loadProfiles()
		if s.cursor >= len(s.profiles) {
			s.cursor = len(s.profiles) - 1
		}
		s.message = fmt.Sprintf("Profile '%s' deleted successfully", profileToDelete)
		s.messageType = "success"
		s.mode = "list"
	}

	return s, nil
}
