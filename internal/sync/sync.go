package sync

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GitSync handles git-based synchronization of dotfiles
type GitSync struct {
	DotfilesDir string
	GitRepo     string
}

// NewGitSync creates a new GitSync instance
func NewGitSync(dotfilesDir, gitRepo string) *GitSync {
	return &GitSync{
		DotfilesDir: dotfilesDir,
		GitRepo:     gitRepo,
	}
}

// IsGitRepo checks if the dotfiles directory is a git repository
func (g *GitSync) IsGitRepo() bool {
	gitDir := filepath.Join(g.DotfilesDir, ".git")
	_, err := os.Stat(gitDir)
	return err == nil
}

// Init initializes a git repository in the dotfiles directory
func (g *GitSync) Init() error {
	if g.IsGitRepo() {
		return nil
	}

	cmd := exec.Command("git", "-C", g.DotfilesDir, "init")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git init failed: %s", string(output))
	}

	return nil
}

// AddRemote adds a remote repository
func (g *GitSync) AddRemote(name, url string) error {
	// Remove existing remote if present
	g.RemoveRemote(name)

	cmd := exec.Command("git", "-C", g.DotfilesDir, "remote", "add", name, url)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git remote add failed: %s", string(output))
	}

	return nil
}

// RemoveRemote removes a remote repository
func (g *GitSync) RemoveRemote(name string) error {
	cmd := exec.Command("git", "-C", g.DotfilesDir, "remote", "remove", name)
	if err := cmd.Run(); err != nil {
		return nil // Remote may not exist, ignore
	}
	return nil
}

// Add stages all changes in the dotfiles directory
func (g *GitSync) Add() error {
	cmd := exec.Command("git", "-C", g.DotfilesDir, "add", "-A")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git add failed: %s", string(output))
	}
	return nil
}

// Commit creates a commit with the given message
func (g *GitSync) Commit(message string) error {
	// Check if there are changes to commit
	statusCmd := exec.Command("git", "-C", g.DotfilesDir, "status", "--porcelain")
	output, err := statusCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git status failed: %s", string(output))
	}

	if len(strings.TrimSpace(string(output))) == 0 {
		return nil // No changes to commit
	}

	cmd := exec.Command("git", "-C", g.DotfilesDir, "commit", "-m", message)
	if commitOutput, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git commit failed: %s", string(commitOutput))
	}

	return nil
}

// Push pushes changes to the remote repository
func (g *GitSync) Push(remote, branch string) error {
	cmd := exec.Command("git", "-C", g.DotfilesDir, "push", "-u", remote, branch)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git push failed: %s", string(output))
	}
	return nil
}

// Pull pulls changes from the remote repository
func (g *GitSync) Pull(remote, branch string) error {
	cmd := exec.Command("git", "-C", g.DotfilesDir, "pull", remote, branch)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git pull failed: %s", string(output))
	}
	return nil
}

// Clone clones a repository to the dotfiles directory
func (g *GitSync) Clone(url string) error {
	// Remove existing directory if it exists
	if _, err := os.Stat(g.DotfilesDir); err == nil {
		if err := os.RemoveAll(g.DotfilesDir); err != nil {
			return fmt.Errorf("failed to remove existing directory: %w", err)
		}
	}

	// Create parent directory
	parent := filepath.Dir(g.DotfilesDir)
	if err := os.MkdirAll(parent, 0755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	cmd := exec.Command("git", "clone", url, g.DotfilesDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git clone failed: %s", string(output))
	}

	return nil
}

// Status returns the current git status
func (g *GitSync) Status() (string, error) {
	cmd := exec.Command("git", "-C", g.DotfilesDir, "status", "--short")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git status failed: %s", string(output))
	}
	return string(output), nil
}

// CurrentBranch returns the current branch name
func (g *GitSync) CurrentBranch() (string, error) {
	cmd := exec.Command("git", "-C", g.DotfilesDir, "branch", "--show-current")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %s", string(output))
	}
	return strings.TrimSpace(string(output)), nil
}

// Sync performs a full sync: add, commit, push
func (g *GitSync) Sync(message, remote, branch string) error {
	if !g.IsGitRepo() {
		if err := g.Init(); err != nil {
			return err
		}
	}

	if err := g.Add(); err != nil {
		return err
	}

	if err := g.Commit(message); err != nil {
		return err
	}

	if g.GitRepo != "" {
		if err := g.Push(remote, branch); err != nil {
			return err
		}
	}

	return nil
}
