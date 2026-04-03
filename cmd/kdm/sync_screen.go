package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/user/kde-dotfiles-manager/internal/config"
	"github.com/user/kde-dotfiles-manager/internal/sync"
)

// syncScreen handles git synchronization
type syncScreen struct {
	parent      *model
	cfg         *config.Config
	gitSync     *sync.GitSync
	menuItems   []syncMenuItem
	cursor      int
	message     string
	messageType string
	width       int
	height      int
}

// syncMenuItem represents a sync action
type syncMenuItem struct {
	id          string
	title       string
	description string
}

func newSyncScreen(parent *model) *syncScreen {
	s := &syncScreen{
		parent: parent,
		cfg:    parent.cfg,
		cursor: 0,
		menuItems: []syncMenuItem{
			{id: "init", title: "Initialize Git Repository", description: "Initialize a new git repository in dotfiles directory"},
			{id: "add", title: "Stage Changes", description: "Stage all changes for commit"},
			{id: "commit", title: "Commit Changes", description: "Commit staged changes with a timestamp message"},
			{id: "push", title: "Push to Remote", description: "Push commits to the remote repository"},
			{id: "pull", title: "Pull from Remote", description: "Pull latest changes from the remote repository"},
			{id: "status", title: "View Git Status", description: "Show current git status"},
			{id: "clone", title: "Clone Repository", description: "Clone remote repository to local dotfiles directory"},
		},
	}

	s.gitSync = sync.NewGitSync(parent.cfg.ExpandPath(), parent.cfg.GitRepo)
	return s
}

func (s *syncScreen) Init() tea.Cmd {
	return nil
}

func (s *syncScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if s.cursor < len(s.menuItems)-1 {
				s.cursor++
			}
		case "enter":
			return s.executeAction()
		}
	}
	return s, nil
}

func (s *syncScreen) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Synchronize Dotfiles"))
	b.WriteString("\n\n")

	// Show git repo status
	repoStatus := "No Git repository"
	if s.gitSync.IsGitRepo() {
		repoStatus = checkStyle.Render("Git repository active")
		if s.cfg.GitRepo != "" {
			repoStatus += fmt.Sprintf(" | Remote: %s", s.cfg.GitRepo)
		}
	}
	b.WriteString(subtitleStyle.Render(repoStatus))
	b.WriteString("\n\n")

	for i, item := range s.menuItems {
		cursor := "  "
		if i == s.cursor {
			cursor = "> "
		}

		line := fmt.Sprintf("%s%s - %s", cursor, item.title, item.description)
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

	b.WriteString("\n\n  Press esc to go back • ↑↓ to navigate • Enter to execute")

	return b.String()
}

// executeAction performs the selected sync action
func (s *syncScreen) executeAction() (tea.Model, tea.Cmd) {
	action := s.menuItems[s.cursor].id

	switch action {
	case "init":
		if err := s.gitSync.Init(); err != nil {
			s.message = fmt.Sprintf("Failed to initialize: %v", err)
			s.messageType = "error"
		} else {
			s.message = "Git repository initialized successfully"
			s.messageType = "success"
		}

	case "add":
		if err := s.gitSync.Add(); err != nil {
			s.message = fmt.Sprintf("Failed to stage changes: %v", err)
			s.messageType = "error"
		} else {
			s.message = "Changes staged successfully"
			s.messageType = "success"
		}

	case "commit":
		msg := "Update KDE dotfiles configuration"
		if err := s.gitSync.Commit(msg); err != nil {
			s.message = fmt.Sprintf("Failed to commit: %v", err)
			s.messageType = "error"
		} else {
			s.message = "Changes committed successfully"
			s.messageType = "success"
		}

	case "push":
		if s.cfg.GitRepo == "" {
			s.message = "No remote repository configured. Set git_repo in config."
			s.messageType = "error"
		} else if err := s.gitSync.Push("origin", "main"); err != nil {
			s.message = fmt.Sprintf("Failed to push: %v", err)
			s.messageType = "error"
		} else {
			s.message = "Pushed to remote successfully"
			s.messageType = "success"
		}

	case "pull":
		if s.cfg.GitRepo == "" {
			s.message = "No remote repository configured"
			s.messageType = "error"
		} else if err := s.gitSync.Pull("origin", "main"); err != nil {
			s.message = fmt.Sprintf("Failed to pull: %v", err)
			s.messageType = "error"
		} else {
			s.message = "Pulled from remote successfully"
			s.messageType = "success"
		}

	case "status":
		status, err := s.gitSync.Status()
		if err != nil {
			s.message = fmt.Sprintf("Failed to get status: %v", err)
			s.messageType = "error"
		} else if status == "" {
			s.message = "Working tree clean - no changes"
			s.messageType = "success"
		} else {
			s.message = fmt.Sprintf("Git status:\n%s", status)
			s.messageType = "info"
		}

	case "clone":
		if s.cfg.GitRepo == "" {
			s.message = "No remote repository configured"
			s.messageType = "error"
		} else if err := s.gitSync.Clone(s.cfg.GitRepo); err != nil {
			s.message = fmt.Sprintf("Failed to clone: %v", err)
			s.messageType = "error"
		} else {
			s.message = "Repository cloned successfully"
			s.messageType = "success"
		}
	}

	return s, nil
}
