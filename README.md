# KDE Dotfiles Manager

A Terminal User Interface (TUI) application for managing KDE Plasma 6+ configuration backups, synchronization, and deployment.

## Features

- **Backup & Restore**: Save and restore KDE Plasma 6+ configurations to dotfiles
- **Synchronization**: Sync configurations across multiple machines
- **Deployment**: Quick deployment of saved configurations
- **TUI Interface**: Beautiful terminal interface built with Bubbletea

### Managed Configuration Categories

- **Window Management**: KWin rules, window behavior, virtual desktops, tiling scripts
- **Keyboard Shortcuts**: Global shortcuts, KWin shortcuts, application shortcuts
- **Themes**: Color schemes, window decorations, cursors, icons, wallpapers, GTK themes
- **Languages**: System locale, input methods, spell checking, keyboard layouts
- **Widgets**: Desktop widgets, panel configurations, desktop layout

## Prerequisites

- Linux with KDE Plasma 6+
- Go 1.22+
- Bash 5.0+
- Git (for sync functionality)

## Installation

```bash
# Clone the repository
git clone https://github.com/user/kde-dotfiles-manager.git
cd kde-dotfiles-manager

# Build the TUI application
make build

# Install bash scripts
make install-scripts
```

## Quick Start

```bash
# Launch the TUI application
./bin/kdm

# Or use bash scripts directly
./scripts/backup.sh    # Backup all KDE configurations
./scripts/restore.sh   # Restore from backup
./scripts/sync.sh      # Sync with remote repository
./scripts/deploy.sh    # Deploy configuration to current system
```

## Project Structure

```
kde-dotfiles-manager/
├── cmd/kdm/main.go          # TUI application entry point
├── internal/
│   ├── config/              # Configuration management
│   ├── kde/                 # KDE-specific path definitions
│   ├── sync/                # Git synchronization logic
│   ├── theme/               # Theme configuration handling
│   ├── widgets/             # Widget configuration handling
│   ├── shortcuts/           # Keyboard shortcut handling
│   └── locales/             # Language/locale configuration
├── scripts/
│   ├── backup.sh            # Full backup script
│   ├── restore.sh           # Full restore script
│   ├── sync.sh              # Git sync script
│   ├── deploy.sh            # Deploy script
│   └── common.sh            # Shared functions
├── assets/                  # Static assets
├── Makefile                 # Build automation
└── README.md
```

## Configuration

The application uses a YAML configuration file located at `~/.config/kde-dotfiles-manager/config.yaml`:

```yaml
# Directory to store dotfiles backups
dotfiles_dir: "~/kde-dotfiles"

# Git repository URL for synchronization
# git_repo: "git@github.com:user/kde-dotfiles.git"

# Categories to include in backup
categories:
  - shortcuts
  - themes
  - window_management
  - languages
  - widgets
  - panels
  - system_settings

# Auto-create backup before restore
backup_before_restore: true

# Verbose logging
verbose: false
```

## Usage

### TUI Application

Launch with `./bin/kdm` and navigate using:
- `Arrow keys` / `j/k` - Navigate menu
- `Enter` - Select/confirm
- `Esc` / `q` - Go back / Quit
- `Space` - Toggle selection
- `r` - Refresh status
- `d` - Deploy selected profile

### Bash Scripts

```bash
# Backup specific categories
./scripts/backup.sh --category shortcuts,themes

# Restore with confirmation
./scripts/restore.sh --interactive

# Sync to remote
./scripts/sync.sh --push

# Deploy specific profile
./scripts/deploy.sh --profile default
```

## License

MIT
